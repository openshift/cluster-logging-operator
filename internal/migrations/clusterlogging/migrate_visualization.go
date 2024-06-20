package clusterlogging

import (
	"fmt"
	"regexp"

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
	} else if spec.Visualization != nil && spec.Visualization.Type == logging.VisualizationTypeOCPConsole {
		spec.Visualization = MigrateOcpSpec(spec.Visualization)
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

func MigrateOcpSpec(visSpec *logging.VisualizationSpec) *logging.VisualizationSpec {

	var pattern = regexp.MustCompile(`^[0-9]+$`)
	log.V(3).Info("Migrating OCP visualization specs from spec.Visualization.OCPConsole")
	// need to fix Timeout format, because OCP Plugin expect it in format: ^([0-9]+)([smhd])$
	// in same time in our code base used validation in format: "^([0-9]+)([smhd]{0,1})$"
	// set 's' because seconds is default value for OCP Plugin timeout
	if visSpec.OCPConsole.Timeout != "" && pattern.MatchString(string(visSpec.OCPConsole.Timeout)) {
		visSpec.OCPConsole.Timeout = visSpec.OCPConsole.Timeout + "s"
	}
	return visSpec
}
