package clusterlogging

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/status"
	corev1 "k8s.io/api/core/v1"
)

var (
	DeprecatedCondition status.ConditionType = "Deprecated"

	DeprecatedForwarderSpecReason      status.ConditionReason = "DeprecatedForwarderSpec"
	DeprecatedCollectionLogsSpecReason status.ConditionReason = "DeprecatedCollectionLogsSpec"
)

// MigrateCollectionSpec moves fluentd tuning and collection.logs.* to collection for
// spec that is common to all collectors
func MigrateCollectionSpec(spec logging.ClusterLoggingSpec) (logging.ClusterLoggingSpec, []logging.Condition) {
	warnings := []logging.Condition{}
	log.V(3).Info("Migrating collectionSpec for reconciliation call", "spec", spec)
	if spec.Collection == nil {
		return spec, nil
	}

	if spec.Forwarder != nil {
		spec.Collection.Fluentd = spec.Forwarder.Fluentd
		spec.Forwarder = nil
		warnings = append(warnings, logging.NewCondition(DeprecatedCondition, corev1.ConditionTrue, DeprecatedForwarderSpecReason, "spec.forwarder is deprecated in favor of spec.collection.fluentd"))
	}

	logSpec := spec.Collection.Logs
	if logSpec != nil {
		warnings = append(warnings, logging.NewCondition(DeprecatedCondition, corev1.ConditionTrue, DeprecatedCollectionLogsSpecReason, "spec.collection.logs.* is deprecated in favor of spec.collection.*"))
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
	return spec, warnings
}
