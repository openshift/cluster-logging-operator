package k8shandler

import (
	"path"

	"github.com/openshift/cluster-logging-operator/internal/utils"
)

const (
	clusterLoggingDashboardFile = "dashboards/openshift-logging-dashboard.json"
)

// CreateOrUpdateDashboards reconciles metrics dashboards component for cluster logging
func (clusterRequest *ClusterLoggingRequest) CreateOrUpdateDashboards() (err error) {
	spec := string(utils.GetFileContents(path.Join(utils.GetShareDir(), clusterLoggingDashboardFile)))

	cm := NewConfigMap("grafana-dashboard-cluster-logging",
		"openshift-config-managed",
		map[string]string{
			"openshift-logging.json": spec,
		})
	if cm.Labels == nil {
		cm.Labels = map[string]string{}
	}
	cm.Labels["console.openshift.io/dashboard"] = "true"

	if err := clusterRequest.CreateOrUpdateConfigMap(cm); err != nil {
		return err
	}

	return nil
}
