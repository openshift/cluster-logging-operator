package migrations

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

// MigrateCollectionSpec moves fluentd tuning and collection.logs.* to collection for
// spec that is common to all collectors
func MigrateCollectionSpec(spec logging.ClusterLoggingSpec) logging.ClusterLoggingSpec {
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
		spec.Collection.Logs = logging.LogCollectionSpec{}
		return spec
	}

	logSpec := spec.Collection.Logs
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
	spec.Collection.Logs = logging.LogCollectionSpec{}

	log.V(3).Info("Migrated collectionSpec for reconciliation", "spec", spec)
	return spec
}
