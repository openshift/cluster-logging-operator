package forwarding

import (
	"context"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/internal/api/initialize"
	"github.com/openshift/cluster-logging-operator/internal/controller"

	"github.com/openshift/cluster-logging-operator/internal/validations/clusterlogforwarder"
	"github.com/openshift/cluster-logging-operator/internal/validations/clusterlogforwarder/conditions"

	"github.com/openshift/cluster-logging-operator/internal/k8s/loader"
	"github.com/openshift/cluster-logging-operator/internal/migrations/observability/api"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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

	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=logging.openshift.io,resources=clusterlogforwarders,verbs=get;list;create;watch;update;patch;delete
// +kubebuilder:rbac:groups=logging.openshift.io,resources=clusterlogforwarders/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=logging.openshift.io,resources=clusterlogforwarders/finalizers,verbs=update
// Reconcile reads that state of the cluster for a ClusterLogForwarder object and makes changes based on the state read
// and what is in the Logging.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileForwarder) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	log.V(3).Info("clusterlogforwarder-controller fetching LF instance", "namespace", request.NamespacedName.Namespace, "name", request.NamespacedName.Name)

	// Fetch the logging.ClusterLogForwarder instance
	instance, err := loader.FetchClusterLogForwarder(r.Client, request.NamespacedName.Namespace, request.NamespacedName.Name)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading - requeue the request.
		return ctrl.Result{}, err
	}

	// Fetch ClusterLogging if there is one
	clInstance, err := loader.FetchClusterLogging(r.Client, request.NamespacedName.Namespace, request.NamespacedName.Name)

	if err != nil && !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	// Check if fluentDForward is used & validates output secrets
	outputSecrets, status, err := clusterlogforwarder.ValidateClusterLogForwarderForConversion(instance, r.Client)
	if err != nil {
		instance.Status = *status
		if result, err := r.updateStatus(instance); err != nil {
			return result, err
		}
		return ctrl.Result{}, err
	}

	// Convert to observability.ClusterLogForwarder
	obsClf, err := api.ConvertLoggingToObservability(r.Client, clInstance, instance, outputSecrets)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Fix indices for default elasticsearch to be `app-write`, `infra-write`, `audit-write`
	obsClf.Spec = initialize.DefaultElasticsearch(obsClf.Spec)

	if err := r.Client.Create(context.TODO(), obsClf); err != nil {
		return ctrl.Result{}, err
	}

	// Annotate logging.CLF with migrated
	instance.Annotations[constants.AnnotationCRConverted] = "true"
	if err = r.Client.Update(context.TODO(), instance); err != nil {
		return ctrl.Result{}, err
	}

	// Set CLF with status of ready with reason migrated
	instance.Status = logging.ClusterLogForwarderStatus{}
	instance.Status.Conditions.SetCondition(conditions.CondReadyWithMessage(logging.ReasonMigrated, "ClusterLogForwarder.logging.openshift.io migrated to ClusterLogForwarder.observability.openshift.io"))
	if result, err := r.updateStatus(instance); err != nil {
		return result, err
	}

	return periodicRequeue, nil
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

// AnnotateLoggingClusterLogForwarders gets a list of existing clusterlogforwarders.logging.openshift.io
// and annotates them with "logging.openshift.io/needs-migration"
func AnnotateLoggingClusterLogForwarders(k8sClient client.Client, apiReader client.Reader) error {
	var clfList logging.ClusterLogForwarderList
	if err := apiReader.List(context.TODO(), &clfList); err != nil {
		return err
	}

	for _, clf := range clfList.Items {
		if clf.Annotations == nil {
			clf.Annotations = map[string]string{}
		}
		clf.Annotations[constants.AnnotationNeedsMigration] = "true"
		if err := k8sClient.Update(context.TODO(), &clf); err != nil {
			return err
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReconcileForwarder) SetupWithManager(mgr ctrl.Manager) error {
	controllerBuilder := ctrl.NewControllerManagedBy(mgr)
	return controllerBuilder.
		For(&logging.ClusterLogForwarder{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		// Ignore create events as well as migrated resources
		WithEventFilter(controller.IgnoreMigratedResources()).
		Complete(r)
}
