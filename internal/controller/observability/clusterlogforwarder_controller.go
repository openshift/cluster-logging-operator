package observability

import (
	"context"
	log "github.com/ViaQ/logerr/v2/log/static"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/collector"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	obsload "github.com/openshift/cluster-logging-operator/internal/k8s/observability"
	"github.com/openshift/cluster-logging-operator/internal/metrics/telemetry"
	validationerrors "github.com/openshift/cluster-logging-operator/internal/validations/errors"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
	"time"

	observabilityv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// periodicRequeue to ensure CLF collection permissions are still valid.  We can not watch
	// ClusterRoleBindings since there is no effective way to associate known CLF with a given binding to
	// avoid needing to reconcile all CRB events
	periodicRequeue = ctrl.Result{
		RequeueAfter: time.Minute * 5,
	}

	defaultRequeue = ctrl.Result{}
)

// ClusterLogForwarderReconciler reconciles a ClusterLogForwarder object
type ClusterLogForwarderReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=observability.openshift.io,resources=clusterlogforwarders,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=observability.openshift.io,resources=clusterlogforwarders/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=observability.openshift.io,resources=clusterlogforwarders/finalizers,verbs=update
func (r *ClusterLogForwarderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//TODO: this was stubbed. Replace with needed with project log abstraction
	//_ = log.FromContext(ctx)
	log.V(3).Info("obs.clf controller reconciling resource", "namespace", req.NamespacedName.Namespace, "name", req.NamespacedName.Name)

	//TODO: enable telemetry
	//telemetry.SetCLFMetrics(0) // Cancel previous info metric
	//defer func() { telemetry.SetCLFMetrics(1) }()

	// Fetch the ClusterLogForwarder instance
	instance, err, status := obsload.FetchClusterLogForwarder(r.Client, req.NamespacedName.Namespace, req.NamespacedName.Name)
	if status != nil {
		//if err := instance.Status.Synchronize(status); err != nil {
		return defaultRequeue, err
		//}
	}

	if result, err := processFetchError(err, r.Client, req, &instance); err != nil {
		return result, err
	}

	// TODO: Deploy Collector

	return periodicRequeue, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterLogForwarderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&observabilityv1.ClusterLogForwarder{}).
		Complete(r)
}

func processFetchError(err error, k8Client client.Client, req ctrl.Request, instance *obsv1.ClusterLogForwarder) (ctrl.Result, error) {
	if err == nil {
		return periodicRequeue, nil
	}
	log.V(3).Info("obs-clusterlogforwarder-controller Error getting instance. It will be retried if other then 'NotFound'", "error", err.Error())
	if validationerrors.MustUndeployCollector(err) {
		if deleteErr := collector.Remove(k8Client, req.Namespace, req.Name); deleteErr != nil {
			log.V(0).Error(deleteErr, "Unable to remove collector deployment")
		}
	}
	if validationerrors.IsValidationError(err) {
		//condition := logging.NewCondition(logging.ValidationCondition, corev1.ConditionTrue, logging.ValidationFailureReason, "%v", err)
		//instance.Status.Conditions.SetCondition(condition)
		//instance.Status.Conditions.SetCondition(logging.CondNotReady(logging.ValidationFailureReason, ""))
		// TODO: Add in event recording?
		//r.Recorder.Event(&instance, "Warning", string(logging.ReasonInvalid), condition.Message)
		telemetry.Data.CLFInfo.Set("healthStatus", constants.UnHealthyStatus)
		return updateStatus(k8Client, instance)
	} else if !errors.IsNotFound(err) {
		// Error reading - requeue the request.
		return defaultRequeue, err
	}
	return defaultRequeue, err
}

func updateStatus(k8Client client.Client, instance *obsv1.ClusterLogForwarder) (ctrl.Result, error) {
	if err := k8Client.Status().Update(context.TODO(), instance); err != nil {

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
