package clusterlogging

import (
	"context"

	"github.com/openshift/cluster-logging-operator/internal/hostedcontrolplane"
	"github.com/openshift/cluster-logging-operator/internal/logstore/lokistack"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/k8s/loader"
	"github.com/openshift/cluster-logging-operator/internal/metrics/telemetry"
	validationerrors "github.com/openshift/cluster-logging-operator/internal/validations/errors"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	log "github.com/ViaQ/logerr/v2/log/static"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/k8shandler"
	loggingruntime "github.com/openshift/cluster-logging-operator/internal/runtime"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ reconcile.Reconciler = &ReconcileClusterLogging{}

// ReconcileClusterLogging reconciles a ClusterLogging object
type ReconcileClusterLogging struct {
	// This Client, initialized using mgr.Client() above, is a split Client
	// that reads objects from the cache and writes to the apiserver
	Client client.Client

	// Reader is an initialized client.Reader that reads objects directly from the apiserver
	// instead of the cache. Useful for cases where need to read/write to a namespace other than
	// the deployed namespace (e.g. openshift-config-managed)
	Reader   client.Reader
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder

	//ClusterVersion is the semantic version of the cluster
	ClusterVersion string
	//ClusterID is the unique identifier of the cluster in which the operator is deployed
	ClusterID string
}

// Reconcile reads that state of the cluster for a ClusterLogging object and makes changes based on the state read
// and what is in the ClusterLogging.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileClusterLogging) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	log.V(3).Info("Clusterlogging reconcile request.", "namespace", request.Namespace, "name", request.Name)
	r.Recorder.Event(loggingruntime.NewClusterLogging(request.NamespacedName.Namespace, request.NamespacedName.Name), corev1.EventTypeNormal, constants.EventReasonReconcilingLoggingCR, "Reconciling logging resource")

	telemetry.SetCLMetrics(0) // Cancel previous info metric
	defer func() { telemetry.SetCLMetrics(1) }()

	// Fetch the ClusterLogging instance
	instance, err, migrationConditions := loader.FetchClusterLogging(r.Client, request.NamespacedName.Namespace, request.NamespacedName.Name, false)
	if err != nil {
		if errors.IsNotFound(err) {
			removeClusterLogging(r.Client)
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		setMigrationStatusConditions(&instance, migrationConditions)
		if validationerrors.IsValidationError(err) {
			instance.Status.Conditions.SetCondition(loggingv1.CondInvalid("validation failed: %v", err))
			return r.updateStatus(&instance)
		}

		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}
	setMigrationStatusConditions(&instance, migrationConditions)
	defer func() {
		if _, err := r.updateStatus(&instance); err != nil {
			log.Error(err, "Error while updating status for ClusterLogging %s/%s", instance.Namespace, instance.Name)
		}
	}()

	if instance.GetDeletionTimestamp() != nil {
		removeClusterLogging(r.Client)
		return ctrl.Result{}, nil
	}

	if err := lokistack.CheckFinalizer(ctx, r.Client, &instance); err != nil {
		return ctrl.Result{}, err
	}

	if instance.Spec.ManagementState == loggingv1.ManagementStateUnmanaged {
		// if cluster is set to unmanaged then set managedStatus as 0
		telemetry.Data.CLInfo.Set("managedStatus", constants.UnManagedStatus)
		return ctrl.Result{}, nil
	}
	clf, err, _ := loader.FetchClusterLogForwarder(r.Client, request.NamespacedName.Namespace, request.NamespacedName.Name, false, func() loggingv1.ClusterLogging { return instance })
	if err != nil && !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	clusterVersion, clusterID := r.GetClusterVersionID(ctx, request.Namespace)
	resourceNames := factory.GenerateResourceNames(clf)
	if err = k8shandler.Reconcile(&instance, &clf, r.Client, r.Reader, r.Recorder, clusterVersion, clusterID, resourceNames); err != nil {
		telemetry.Data.CLInfo.Set("healthStatus", constants.UnHealthyStatus)
		log.Error(err, "Error reconciling clusterlogging instance")
		instance.Status.Conditions.SetCondition(loggingv1.CondInvalid("error reconciling clusterlogging instance: %v", err))
	} else {
		// Set condition ready if no errors
		instance.Status.Conditions.SetCondition(loggingv1.CondReady)
	}

	return ctrl.Result{}, err
}

func removeClusterLogging(k8Client client.Client) {
	// ClusterLogging is being deleted, remove resources that can not be garbage-collected.
	if err := lokistack.RemoveRbac(k8Client); err != nil {
		log.Error(err, "Error removing RBAC for accessing LokiStack.")
	}
}

func setMigrationStatusConditions(cl *loggingv1.ClusterLogging, conditions []loggingv1.Condition) {
	for _, cond := range conditions {
		cl.Status.Conditions.SetCondition(cond)
	}
}

func (r *ReconcileClusterLogging) updateStatus(instance *loggingv1.ClusterLogging) (ctrl.Result, error) {
	if err := r.Client.Status().Update(context.TODO(), instance); err != nil {

		if strings.Contains(err.Error(), constants.OptimisticLockErrorMsg) {
			// do manual retry without error
			// more information about this error here: https://github.com/kubernetes/kubernetes/issues/28149
			return reconcile.Result{RequeueAfter: time.Second * 1}, nil
		}

		telemetry.Data.CLInfo.Set("healthStatus", constants.UnHealthyStatus)
		log.Error(err, "clusterlogging-controller error updating status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReconcileClusterLogging) SetupWithManager(mgr ctrl.Manager) error {
	controllerBuilder := ctrl.NewControllerManagedBy(mgr).
		For(&loggingv1.ClusterLogging{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&appsv1.Deployment{}).
		Owns(&appsv1.DaemonSet{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&monitoringv1.ServiceMonitor{}).
		Watches(&source.Kind{Type: &corev1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []reconcile.Request {
				if obj.GetNamespace() == constants.OpenshiftNS && obj.GetLabels()["component"] == constants.ElasticsearchName {
					return []reconcile.Request{
						{
							NamespacedName: types.NamespacedName{
								Namespace: obj.GetNamespace(),
								Name:      obj.GetName(),
							},
						},
					}
				}
				return nil
			}))

	return controllerBuilder.Complete(r)
}

func (r *ReconcileClusterLogging) GetClusterVersionID(ctx context.Context, namespace string) (version, id string) {
	// If reconciling in a hosted control plane namespace, use the guest cluster version/id
	// provided by the hostedcontrolplane resource.
	if info := hostedcontrolplane.GetVersionID(ctx, r.Client, namespace); info != nil {
		return info.Version, info.ID
	}
	// Otherwise use the cluster-ID established at start-up.
	return r.ClusterVersion, r.ClusterID
}
