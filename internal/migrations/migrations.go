package migrations

import (
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/migrations/clusterlogging"
)

func MigrateClusterLogging(spec loggingv1.ClusterLoggingSpec) (loggingv1.ClusterLoggingSpec, []loggingv1.Condition) {
	status := []loggingv1.Condition{}
	for _, migrate := range migrations {
		migratedSpec, migrationStatus := migrate(spec)
		status = append(status, migrationStatus...)
		spec = migratedSpec
	}
	return spec, status
}

// migrations are the set of rules for migrating a ClusterLogging that modify the spec
// for reconciliation and provides warning or informational messages to be added as status
var migrations = []func(cl loggingv1.ClusterLoggingSpec) (loggingv1.ClusterLoggingSpec, []loggingv1.Condition){
	clusterlogging.MigrateCollectionSpec,
	clusterlogging.MigrateVisualizationSpec,
}
