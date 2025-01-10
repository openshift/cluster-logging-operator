package dashboard

import (
	"context"
	_ "embed"
	"fmt"
	staticlog "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators"
	"github.com/openshift/cluster-logging-operator/version"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DashboardName     = "grafana-dashboard-cluster-logging"
	DashboardNS       = "openshift-config-managed"
	DashboardFileName = "openshift-logging.json"
	DashboardHashName = "contentHash"
)

var (
	//go:embed openshift-logging-dashboard.json
	DashboardConfig string
	log             = staticlog.WithName("dashboard")
)

type ReconcileDashboards struct {
	Client client.Client

	// Reader is an initialized client.Reader that reads objects directly from the apiserver
	// instead of the cache. Useful for cases where need to read/write to a namespace other than
	// the deployed namespace (e.g. openshift-config-managed)
	Reader client.Reader
}

func (r *ReconcileDashboards) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	log.V(4).Info("#Reconcile", "dashboard", DashboardName)
	return ctrl.Result{}, ReconcileForDashboards(r.Client, r.Reader)
}

func newDashboardConfigMap() *corev1.ConfigMap {
	hash, err := utils.CalculateMD5Hash(DashboardConfig)
	if err != nil {
		log.Error(err, "Error calculated hash for metrics dashboard")
	}
	cm := runtime.NewConfigMap(DashboardNS,
		DashboardName,
		map[string]string{
			DashboardFileName: DashboardConfig,
		},
	)
	runtime.NewConfigMapBuilder(cm).
		AddLabel("console.openshift.io/dashboard", "true").
		AddLabel(DashboardHashName, hash).
		AddLabel(constants.LabelK8sVersion, version.Version).
		AddLabel(constants.LabelK8sManagedBy, constants.ClusterLoggingOperator)
	return cm
}

func ReconcileForDashboards(k8sClient client.Client, reader client.Reader) error {
	cm := newDashboardConfigMap()

	current := &corev1.ConfigMap{}
	key := client.ObjectKeyFromObject(cm)
	err := reader.Get(context.TODO(), key, current)

	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get %v configmap: %v", key, err)
	}

	if err = reconcile.Configmap(k8sClient, reader, cm, comparators.CompareLabels); err != nil {
		return err
	}

	return nil
}

// RemoveDashboardConfigMap removes the config map in the grafana dashboard
func RemoveDashboardConfigMap(c client.Client, r client.Reader) (err error) {
	cm := newDashboardConfigMap()

	current := &corev1.ConfigMap{}
	key := client.ObjectKeyFromObject(cm)
	if err := r.Get(context.TODO(), key, current); err != nil {
		if !errors.IsNotFound(err) {
			return fmt.Errorf("failed to get %v configmap: %v", key, err)
		}
		return nil
	}

	if runtime.Labels(current).Includes(map[string]string{
		constants.LabelK8sManagedBy: constants.ClusterLoggingOperator,
		constants.LabelK8sVersion:   version.Version,
	}) {
		return c.Delete(context.TODO(), cm)
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager
func (r *ReconcileDashboards) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("logging_dashboard_controller").
		Watches(&corev1.ConfigMap{},
			handler.EnqueueRequestsFromMapFunc(func(cxt context.Context, obj client.Object) []ctrl.Request {
				if obj.GetName() == DashboardName && obj.GetNamespace() == DashboardNS {
					return []ctrl.Request{
						{
							NamespacedName: types.NamespacedName{
								Namespace: obj.GetNamespace(),
								Name:      obj.GetName(),
							},
						},
					}
				}
				return nil
			})).Complete(r)
}
