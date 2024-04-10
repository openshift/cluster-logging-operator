package dashboard

import (
	"context"
	_ "embed"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlruntime "sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	DashboardName     = "grafana-dashboard-cluster-logging"
	DashboardNS       = "openshift-config-managed"
	DashboardFileName = "openshift-logging.json"
	DashboardHashName = "contentHash"
)

//go:embed openshift-logging-dashboard.json
var DashboardConfig string

var _ ctrlruntime.Reconciler = &ReconcileDashboards{}

type ReconcileDashboards struct {
	Client client.Client

	// Reader is an initialized client.Reader that reads objects directly from the apiserver
	// instead of the cache. Useful for cases where need to read/write to a namespace other than
	// the deployed namespace (e.g. openshift-config-managed)
	Reader client.Reader
}

func (r *ReconcileDashboards) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	log.V(3).Info("reconcile ", "ConfigMap", DashboardName)
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
		AddLabel(DashboardHashName, hash)
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
func RemoveDashboardConfigMap(c client.Client) (err error) {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DashboardName,
			Namespace: DashboardNS,
		},
	}
	return c.Delete(context.TODO(), cm)
}

// SetupWithManager sets up the controller with the Manager
func (r *ReconcileDashboards) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
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
