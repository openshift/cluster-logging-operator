package clusterlogging

import (
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
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
		// Case when spec.Visualization.Kibana.NodeSelector or Tolerations are defined
		// Need to move to top level spec.Visualization.NodeSelector/Tolerations
	} else if spec.Visualization != nil && spec.Visualization.Type == logging.VisualizationTypeKibana {
		spec.Visualization = MigrateKibanaSpec(spec.Visualization)
	}

	log.V(3).Info("Migrated visualizationSpec for reconciliation", "spec", spec)
	return spec, nil
}

func MigrateKibanaSpec(visSpec *logging.VisualizationSpec) *logging.VisualizationSpec {
	if visSpec.Kibana == nil {
		log.V(3).Info("kibana visualization specs empty")
		return visSpec
	}

	log.V(3).Info("Migrating kibana visualization specs from spec.Visualization.Kibana")
	// Migrate nodeSelector and Tolerations
	if visSpec.Kibana.NodeSelector != nil {
		visSpec.NodeSelector = visSpec.Kibana.NodeSelector
	}

	if visSpec.Kibana.Tolerations != nil {
		visSpec.Tolerations = visSpec.Kibana.Tolerations
	}

	return visSpec

}
