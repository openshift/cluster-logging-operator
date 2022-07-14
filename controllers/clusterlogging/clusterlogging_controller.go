package clusterlogging

import (
	"context"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	log "github.com/ViaQ/logerr/v2/log/static"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/k8shandler"
	"github.com/openshift/cluster-logging-operator/internal/metrics"
	"github.com/openshift/cluster-logging-operator/internal/telemetry"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
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
}

// Reconcile reads that state of the cluster for a ClusterLogging object and makes changes based on the state read
// and what is in the ClusterLogging.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileClusterLogging) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	log.V(3).Info("Clusterlogging reconcile request.", "name", request.Name)

	// Fetch the ClusterLogging instance
	instance := &loggingv1.ClusterLogging{}
	err := r.Client.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			if err := metrics.RemoveDashboardConfigMap(r.Client); err != nil {
				log.V(1).Error(err, "error deleting grafana configmap")
			}
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	if instance.Spec.ManagementState == loggingv1.ManagementStateUnmanaged {
		// if cluster is set to unmanaged then set managedStatus as 0
		telemetry.Data.CLInfo.Set("managedStatus", constants.UnManagedStatus)
		telemetry.UpdateCLMetricsNoErr()
		return ctrl.Result{}, nil
	}

	if _, err = k8shandler.Reconcile(r.Client, r.Reader, r.Recorder); err != nil {
		telemetry.Data.CLInfo.Set("healthStatus", constants.UnHealthyStatus)
		telemetry.UpdateCLMetricsNoErr()
		log.Error(err, "Error reconciling clusterlogging instance")
	}

	if result, err := r.updateStatus(instance); err != nil {
		return result, err
	}

	return ctrl.Result{}, err
}

func (r *ReconcileClusterLogging) updateStatus(instance *loggingv1.ClusterLogging) (ctrl.Result, error) {
	if err := r.Client.Status().Update(context.TODO(), instance); err != nil {
		telemetry.Data.CLInfo.Set("healthStatus", constants.UnHealthyStatus)
		telemetry.UpdateCLMetricsNoErr()
		log.Error(err, "clusterlogging-controller error updating status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func IsElasticsearchRelatedObj(obj client.Object) (bool, ctrl.Request) {
	if obj.GetNamespace() == "" {
		// ignore cluster scope objects
		return false, ctrl.Request{}
	}

	if constants.OpenshiftNS != obj.GetNamespace() {
		// ignore object in another namespace
		return false, ctrl.Request{}
	}

	if value, exists := obj.GetLabels()["component"]; !exists || value != constants.ElasticsearchName {
		return false, ctrl.Request{}
	}

	return true, ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      constants.SingletonName,
			Namespace: constants.OpenshiftNS,
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReconcileClusterLogging) SetupWithManager(mgr ctrl.Manager) error {
	onAllExceptGenericEventsPredicate := predicate.Funcs{
		UpdateFunc: func(evt event.UpdateEvent) bool {
			return true
		},
		CreateFunc: func(evt event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(evt event.DeleteEvent) bool {
			return true
		},
		GenericFunc: func(evt event.GenericEvent) bool {
			return false
		},
	}

	var toElasticsearchRelatedObjRequestMapper handler.MapFunc = func(obj client.Object) []ctrl.Request {
		isElasticsearchRelatedObj, reconcileRequest := IsElasticsearchRelatedObj(obj)
		if isElasticsearchRelatedObj {
			return []ctrl.Request{reconcileRequest}
		}
		return []ctrl.Request{}
	}

	controllerBuilder := ctrl.NewControllerManagedBy(mgr).
		Watches(&source.Kind{Type: &loggingv1.ClusterLogging{}}, &handler.EnqueueRequestForObject{}).
		Watches(&source.Kind{Type: &loggingv1.ClusterLogForwarder{}}, &handler.EnqueueRequestForObject{}).
		Watches(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &loggingv1.ClusterLogging{},
		}).
		Watches(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &loggingv1.ClusterLogging{},
		}).
		Watches(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &loggingv1.ClusterLogging{},
		}).
		Watches(&source.Kind{Type: &rbacv1.Role{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &loggingv1.ClusterLogging{},
		}).
		Watches(&source.Kind{Type: &rbacv1.RoleBinding{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &loggingv1.ClusterLogging{},
		}).
		Watches(&source.Kind{Type: &corev1.ServiceAccount{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &loggingv1.ClusterLogging{},
		}).
		Watches(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &loggingv1.ClusterLogging{},
		}).
		Watches(&source.Kind{Type: &appsv1.DaemonSet{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &loggingv1.ClusterLogging{},
		}).
		Watches(&source.Kind{Type: &corev1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(toElasticsearchRelatedObjRequestMapper),
			builder.WithPredicates(onAllExceptGenericEventsPredicate),
		)

	return controllerBuilder.
		For(&loggingv1.ClusterLogging{}).
		Complete(r)
}
