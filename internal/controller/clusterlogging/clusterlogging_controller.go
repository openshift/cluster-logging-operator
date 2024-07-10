package clusterlogging

import (
	"context"
	"github.com/openshift/cluster-logging-operator/internal/api/initialize"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/internal/controller"
	"github.com/openshift/cluster-logging-operator/internal/validations/clusterlogforwarder"
	"github.com/openshift/cluster-logging-operator/internal/validations/clusterlogforwarder/conditions"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/openshift/cluster-logging-operator/internal/k8s/loader"
	"github.com/openshift/cluster-logging-operator/internal/migrations/observability/api"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	ctrl "sigs.k8s.io/controller-runtime"

	log "github.com/ViaQ/logerr/v2/log/static"
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
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
	Reader client.Reader
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=logging.openshift.io,resources=clusterloggings,verbs=get;list;create;watch;update;patch;delete
// +kubebuilder:rbac:groups=logging.openshift.io,resources=clusterloggings/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=logging.openshift.io,resources=clusterloggings/finalizers,verbs=update
// Reconcile reads that state of the cluster for a ClusterLogging object and makes changes based on the state read
// and what is in the ClusterLogging.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileClusterLogging) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	log.V(3).Info("Clusterlogging reconcile request.", "namespace", request.Namespace, "name", request.Name)

	// obs.CLF not found so start migration
	// Fetch the ClusterLogging instance
	instance, err := loader.FetchClusterLogging(r.Client, request.NamespacedName.Namespace, request.NamespacedName.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	// Fetch associated logging.ClusterLogForwarder if any
	clf, err := loader.FetchClusterLogForwarder(r.Client, request.NamespacedName.Namespace, request.NamespacedName.Name)
	if err != nil && !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	// Check if fluentDForward is used & validates output secrets
	outputSecrets, status, err := clusterlogforwarder.ValidateClusterLogForwarderForConversion(clf, r.Client)
	if err != nil {
		clf.Status = *status
		if result, err := r.updateStatus(instance); err != nil {
			return result, err
		}
		return ctrl.Result{}, err
	}

	// Convert to observability.ClusterLogForwarder
	obsClf := api.ConvertLoggingToObservability(r.Client, instance, clf, outputSecrets)
	// Fix indices for default elasticsearch to be `app-write`, `infra-write`, `audit-write`
	obsClf.Spec = initialize.DefaultElasticsearch(obsClf.Spec)

	if err = r.Client.Create(context.TODO(), obsClf); err != nil {
		return ctrl.Result{}, err
	}

	// Annotate CL with migrated
	instance.Annotations[constants.AnnotationCRConverted] = "true"
	if err = r.Client.Update(context.TODO(), instance); err != nil {
		return ctrl.Result{}, err
	}

	// Set and update status
	instance.Status = loggingv1.ClusterLoggingStatus{}
	instance.Status.Conditions.SetCondition(conditions.CondReadyWithMessage(loggingv1.ReasonMigrated, "ClusterLogging.logging.openshift.io migrated to ClusterLogForwarder.observability.openshift.io"))
	if result, err := r.updateStatus(instance); err != nil {
		return result, err
	}

	return ctrl.Result{}, err
}

func (r *ReconcileClusterLogging) updateStatus(instance *loggingv1.ClusterLogging) (ctrl.Result, error) {
	if err := r.Client.Status().Update(context.TODO(), instance); err != nil {

		if strings.Contains(err.Error(), constants.OptimisticLockErrorMsg) {
			// do manual retry without error
			// more information about this error here: https://github.com/kubernetes/kubernetes/issues/28149
			return reconcile.Result{RequeueAfter: time.Second * 1}, nil
		}

		log.Error(err, "clusterlogging-controller error updating status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReconcileClusterLogging) SetupWithManager(mgr ctrl.Manager) error {
	controllerBuilder := ctrl.NewControllerManagedBy(mgr).
		For(&loggingv1.ClusterLogging{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		// Ignore create events as well as migrated resources
		WithEventFilter(controller.IgnoreMigratedResources(constants.AnnotationCRConverted))

	return controllerBuilder.Complete(r)
}
