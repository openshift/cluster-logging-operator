package clusterlogging

import (
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

// MigrateVisualizationSpec initialize spec.visualization with ocp-console plugin,
// if spec.LogStore was defined as 'lokistack' and the spec.visualization not defined
func MigrateVisualizationSpec(spec logging.ClusterLoggingSpec) (logging.ClusterLoggingSpec, []logging.Condition) {
	log.V(3).Info("Migrating visualizationSpec for reconciliation call", "spec", spec)
	if spec.Visualization == nil && spec.LogStore != nil && spec.LogStore.Type == logging.LogStoreTypeLokiStack {
		log.V(3).Info(fmt.Sprintf("Migrating: visualisation spec not set but LogStore is %s, going to set %s as default visualisation spec", spec.LogStore.Type, logging.VisualizationTypeOCPConsole))
		spec.Visualization = &logging.VisualizationSpec{
			Type:       logging.VisualizationTypeOCPConsole,
			OCPConsole: &logging.OCPConsoleSpec{},
		}
	}

	log.V(3).Info("Migrated visualizationSpec for reconciliation", "spec", spec)
	return spec, nil
}
