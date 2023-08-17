package dashboard

import (
	"context"
	"fmt"

	_ "embed"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DashboardName     = "grafana-dashboard-cluster-logging"
	DashboardNS       = "openshift-config-managed"
	DashboardFileName = "openshift-logging.json"
	DashboardHashName = "contentHash"
)

//go:embed openshift-logging-dashboard.json
var DashboardConfig string

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

func ReconcileDashboards(k8sClient client.Client, reader client.Reader) error {
	var err error
	cm := newDashboardConfigMap()

	current := &corev1.ConfigMap{}
	key := client.ObjectKeyFromObject(cm)
	err = reader.Get(context.TODO(), key, current)

	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get %v configmap: %v", key, err)
	}

	if err := reconcile.Configmap(k8sClient, reader, cm, comparators.CompareLabels); err != nil {
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
