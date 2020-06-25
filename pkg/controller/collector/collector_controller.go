package collector

import (
	"context"
	"time"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	collector "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/pkg/utils"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	promtailAnnotation = "clusterlogging.openshift.io/promtaildevpreview"
)

// Add creates a new Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCollector{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("collector-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Collector
	err = c.Watch(&source.Kind{Type: &collector.Collector{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileCollector{}

// ReconcileCollector reconciles a Collector object
type ReconcileCollector struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

var (
	reconcilePeriod = 30 * time.Second
	reconcileResult = reconcile.Result{RequeueAfter: reconcilePeriod}
)

// Reconcile reads that state of the cluster for a Collector object and makes changes based on the state read
// and what is in the Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileCollector) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger.Debugf("Reconcile 'collector' request: %v", request)
	// Fetch the instance
	instance := &collector.Collector{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			logger.Debug("Collector 'instance' not found")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	value, _ := utils.GetAnnotation(promtailAnnotation, instance.ObjectMeta)
	logger.Debugf("Annotation %q value: %q", promtailAnnotation, value)

	//check for instancename and then update status
	var reconcileErr error

	if instance.Name == constants.SingletonName && value == "enabled" {

		clInstance := &logging.ClusterLogging{}
		clName := types.NamespacedName{Name: constants.SingletonName, Namespace: constants.OpenshiftNS}
		err = r.client.Get(context.TODO(), clName, clInstance)
		if err != nil && !errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, err
		}

		reconcileErr = k8shandler.ReconcileCollector(clInstance, &instance.Spec, r.client)
	} else {
		logger.Debugf("Not requeing request as collector is not named 'instance' or devpreview is not 'enabled'")
		return reconcile.Result{Requeue: false}, nil
	}

	return reconcileResult, reconcileErr
}
