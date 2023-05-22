package migrations

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

func MigrateClusterLogging(spec loggingv1.ClusterLoggingSpec) loggingv1.ClusterLoggingSpec {
	spec = MigrateCollectionSpec(spec)
	if spec.Visualization != nil {
		*spec.Visualization = MigrateVisualizationSpec(*spec.Visualization)
	}

	return spec
}

// MigrateCollectionSpec moves fluentd tuning and collection.logs.* to collection for
// spec that is common to all collectors
func MigrateCollectionSpec(spec loggingv1.ClusterLoggingSpec) loggingv1.ClusterLoggingSpec {
	log.V(3).Info("Migrating collectionSpec for reconciliation call", "spec", spec)
	if spec.Collection == nil {
		return spec
	}

	if spec.Forwarder != nil {
		spec.Collection.Fluentd = spec.Forwarder.Fluentd
		spec.Forwarder = nil
	}

	if spec.Collection.Type != "" && spec.Collection.Type.IsSupportedCollector() {
		log.V(3).Info("collectionSpec already using latest. removing spec.Collection.Logs while reconciling")
		spec.Collection.Logs = nil
		return spec
	}

	logSpec := spec.Collection.Logs
	if logSpec != nil {
		spec.Collection.Type = logSpec.Type
		if logSpec.Resources != nil {
			spec.Collection.CollectorSpec.Resources = logSpec.Resources
		}
		if logSpec.NodeSelector != nil {
			spec.Collection.CollectorSpec.NodeSelector = logSpec.NodeSelector
		}
		if logSpec.Tolerations != nil {
			spec.Collection.CollectorSpec.Tolerations = logSpec.Tolerations
		}
	}

	spec.Collection.Logs = nil

	log.V(3).Info("Migrated collectionSpec for reconciliation", "spec", spec)
	return spec
}

func MigrateVisualizationSpec(visSpec loggingv1.VisualizationSpec) loggingv1.VisualizationSpec {
	if visSpec.Type == loggingv1.VisualizationTypeKibana {
		visSpec = MigrateKibanaSpecs(visSpec)
	}
	return visSpec
}

func MigrateKibanaSpecs(visSpec loggingv1.VisualizationSpec) loggingv1.VisualizationSpec {
	if visSpec.Kibana == nil {
		log.V(2).Info("kibana visualization specs empty")
		return visSpec
	}

	log.V(3).Info("Migrating kibana visualization specs")
	// Migrate nodeSelector and Tolerations
	if visSpec.Kibana.NodeSelector != nil {
		visSpec.NodeSelector = visSpec.Kibana.NodeSelector
	}

	if visSpec.Kibana.Tolerations != nil {
		visSpec.Tolerations = visSpec.Kibana.Tolerations
	}

	return visSpec
}
