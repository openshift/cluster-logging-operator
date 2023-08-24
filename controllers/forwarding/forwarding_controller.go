package forwarding

import (
	"context"
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/collector"
	"github.com/openshift/cluster-logging-operator/internal/status"
	v1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/k8s/loader"

	"github.com/openshift/cluster-logging-operator/internal/metrics/telemetry"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"

	"github.com/openshift/cluster-logging-operator/internal/k8shandler"
	loggingruntime "github.com/openshift/cluster-logging-operator/internal/runtime"
	validationerrors "github.com/openshift/cluster-logging-operator/internal/validations/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	_ reconcile.Reconciler = &ReconcileForwarder{}

	// periodicRequeue to ensure CLF collection permissions are still valid.  We can not watch
	// ClusterRoleBindings since there is no effective way to associate known CLF with a given binding to
	// avoid needing to reconcile all CRB events
	periodicRequeue = ctrl.Result{
		RequeueAfter: time.Minute * 5,
	}

	ValidationCondition     status.ConditionType   = "Validation"
	ValidationFailureReason status.ConditionReason = "ValidationFailure"
)

// ReconcileForwarder reconciles a ClusterLogForwarder object
type ReconcileForwarder struct {
	// This Client, initialized using mgr.Client() above, is a split Client
	// that reads objects from the cache and writes to the apiserver
	Client client.Client

	// Reader is an initialized client.Reader that reads objects directly from the apiserver
	// instead of the cache. Useful for cases where need to read/write to a namespace other than
	// the deployed namespace (e.g. openshift-config-managed)
	Reader client.Reader

	Scheme   *runtime.Scheme
	Recorder record.EventRecorder

	//ClusterVersion is the semantic version of the cluster
	ClusterVersion string
	//ClusterID is the unique identifier of the cluster in which the operator is deployed
	ClusterID string
}

// Reconcile reads that state of the cluster for a ClusterLogForwarder object and makes changes based on the state read
// and what is in the Logging.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileForwarder) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	log.V(3).Info("clusterlogforwarder-controller fetching LF instance", "namespace", request.NamespacedName.Namespace, "name", request.NamespacedName.Name)
	r.Recorder.Event(loggingruntime.NewClusterLogForwarder(request.NamespacedName.Namespace, request.NamespacedName.Name), corev1.EventTypeNormal, constants.EventReasonReconcilingLoggingCR, "Reconciling logging resource")

	telemetry.SetCLFMetrics(0) // Cancel previous info metric
	defer func() { telemetry.SetCLFMetrics(1) }()

	cl, err := r.fetchOrStubClusterLogging(request)
	if err != nil && !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	// Fetch the ClusterLogForwarder instance
	instance, err, status := loader.FetchClusterLogForwarder(r.Client, request.NamespacedName.Namespace, request.NamespacedName.Name, true, func() logging.ClusterLogging { return *cl })
	if status != nil {
		instance.Status = *status
	}
	if err != nil {
		log.V(3).Info("clusterlogforwarder-controller Error getting instance. It will be retried if other then 'NotFound'", "error", err.Error())
		if validationerrors.MustUndeployCollector(err) {
			name := factory.GenerateResourceNames(instance).DaemonSetName()
			if deleteErr := collector.Remove(r.Client, instance.Namespace, name); deleteErr != nil {
				log.V(0).Error(deleteErr, "Unable to remove collector deployment")
			}
		}
		if validationerrors.IsValidationError(err) {
			condition := logging.NewCondition(ValidationCondition, corev1.ConditionTrue, ValidationFailureReason, "%v", err)
			instance.Status.Conditions.SetCondition(condition)
			r.Recorder.Event(&instance, "Warning", string(logging.ReasonInvalid), condition.Message)
			telemetry.Data.CLFInfo.Set("healthStatus", constants.UnHealthyStatus)
			return r.updateStatus(&instance)
		} else if !errors.IsNotFound(err) {
			// Error reading - requeue the request.
			return ctrl.Result{}, err
		}

		// else the object is not found -- meaning it was removed so stop reconciliation
		return ctrl.Result{}, nil
	}

	log.V(3).Info("clusterlogforwarder-controller run reconciler...")

	resourceNames := factory.GenerateResourceNames(instance)
	reconcileErr := k8shandler.Reconcile(cl, &instance, r.Client, r.Reader, r.Recorder, r.ClusterVersion, r.ClusterID, resourceNames)
	if reconcileErr != nil {
		// if cluster is set to fail to reconcile then set healthStatus as 0
		telemetry.Data.CLFInfo.Set("healthStatus", constants.UnHealthyStatus)
		log.V(2).Error(reconcileErr, "clusterlogforwarder-controller returning, error")
		logging.SetCondition(&instance.Status.Conditions, logging.CollectorDeadEnd, corev1.ConditionTrue, logging.ReasonInvalid, "error reconciling clusterlogforwarder instance: %v", reconcileErr)
	} else {
		if instance.Status.Conditions.IsTrueFor(logging.ConditionReady) {
			logging.SetCondition(&instance.Status.Conditions, logging.CollectorDeadEnd, corev1.ConditionFalse, "", "")
			// This returns False if SetCondition updates the condition instead of setting it.
			// For condReady, it will always be updating the status.
			if !instance.Status.Conditions.SetCondition(logging.CondReady) {
				telemetry.Data.CLFInfo.Set("healthStatus", constants.HealthyStatus)
				r.Recorder.Event(&instance, "Normal", string(logging.CondReady.Type), "ClusterLogForwarder is valid")
			}
		}
	}

	if result, err := r.updateStatus(&instance); err != nil {
		return result, err
	}

	return periodicRequeue, reconcileErr
}

// fetchOrStubClusterLogging retrieves ClusterLogging as one of:
// * ClusterLogging <Namespace>/<Name> for ClusterLogForwarder <Namespace>/<Name>
// * ClusterLogging only providing spec.collection.type=vector  for ClusterLogForwarder <Namespace>/<Name> when CL NotFound
func (r *ReconcileForwarder) fetchOrStubClusterLogging(request ctrl.Request) (*logging.ClusterLogging, error) {
	cl, err, _ := loader.FetchClusterLogging(r.Client, request.NamespacedName.Namespace, request.NamespacedName.Name, false)
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}
		if request.NamespacedName.Namespace != constants.OpenshiftNS || (request.NamespacedName.Namespace == constants.OpenshiftNS && request.NamespacedName.Name != constants.SingletonName) {
			cl = *loggingruntime.NewClusterLogging(request.NamespacedName.Namespace, request.NamespacedName.Name)
			cl.Spec = logging.ClusterLoggingSpec{
				Collection: &logging.CollectionSpec{
					Type: logging.LogCollectionTypeVector,
				},
			}
		} else {
			return nil, fmt.Errorf("ClusterLogging (%s/%s) to support ClusterLogForwarder of same namespace and name not found", request.NamespacedName.Namespace, request.NamespacedName.Name)
		}
	}
	return &cl, nil
}

func (r *ReconcileForwarder) updateStatus(instance *logging.ClusterLogForwarder) (ctrl.Result, error) {
	if err := r.Client.Status().Update(context.TODO(), instance); err != nil {

		if strings.Contains(err.Error(), constants.OptimisticLockErrorMsg) {
			// do manual retry without error
			// more information about this error here: https://github.com/kubernetes/kubernetes/issues/28149
			return reconcile.Result{RequeueAfter: time.Second * 1}, nil
		}

		log.Error(err, "clusterlogforwarder-controller error updating status")
		return ctrl.Result{}, err
	}

	return periodicRequeue, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReconcileForwarder) SetupWithManager(mgr ctrl.Manager) error {
	controllerBuilder := ctrl.NewControllerManagedBy(mgr)
	return controllerBuilder.
		For(&logging.ClusterLogForwarder{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.DaemonSet{}).
		Owns(&appsv1.Deployment{}).
		Owns(&rbacv1.Role{}).
		Owns(&rbacv1.RoleBinding{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ServiceAccount{}).
		Owns(&v1.ServiceMonitor{}).
		Complete(r)
}
