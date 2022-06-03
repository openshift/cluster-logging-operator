package metrics

import (
	"path"

	"github.com/ViaQ/logerr/v2/log"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/configmaps"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ClusterLoggingDashboardFile = "dashboards/openshift-logging-dashboard.json"
	DashboardName               = "grafana-dashboard-cluster-logging"
	DashboardNS                 = "openshift-config-managed"
	DashboardFileName           = "openshift-logging.json"
	DashboardHashName           = "contentHash"
)

func newDashboardConfigMap() *corev1.ConfigMap {
	spec := string(utils.GetFileContents(path.Join(utils.GetShareDir(), ClusterLoggingDashboardFile)))
	hash, err := utils.CalculateMD5Hash(spec)
	if err != nil {
		log.NewLogger("dashboards").Error(err, "Error calculated hash for metrics dashboard")
	}
	cm := runtime.NewConfigMap(DashboardNS,
		DashboardName,
		map[string]string{
			DashboardFileName: spec,
		},
	)
	runtime.NewConfigMapBuilder(cm).
		AddLabel("console.openshift.io/dashboard", "true").
		AddLabel(DashboardHashName, hash)

	return cm
}

func ReconcileDashboards(writer client.Writer, reader client.Reader) (err error) {
	cm := newDashboardConfigMap()
	if err := reconcile.ReconcileConfigmap(writer, reader, cm, configmaps.CompareLabels); err != nil {
		return err
	}

	return nil
}
