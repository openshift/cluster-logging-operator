package migrations

import (
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/migrations/clusterlogforwarder"
	"github.com/openshift/cluster-logging-operator/internal/migrations/clusterlogging"
)

func MigrateClusterLogging(spec loggingv1.ClusterLoggingSpec) (loggingv1.ClusterLoggingSpec, []loggingv1.Condition) {
	status := []loggingv1.Condition{}
	for _, migrate := range clMigrations {
		migratedSpec, migrationStatus := migrate(spec)
		status = append(status, migrationStatus...)
		spec = migratedSpec
	}
	return spec, status
}

// migrations are the set of rules for migrating a ClusterLogging that modify the spec
// for reconciliation and provides warning or informational messages to be added as status
var clMigrations = []func(cl loggingv1.ClusterLoggingSpec) (loggingv1.ClusterLoggingSpec, []loggingv1.Condition){
	clusterlogging.MigrateVisualizationSpec,
}

func MigrateClusterLogForwarder(namespace, name string, spec loggingv1.ClusterLogForwarderSpec, logStore *loggingv1.LogStoreSpec, extras map[string]bool, logstoreSecretName, saTokenSecret string) (loggingv1.ClusterLogForwarderSpec, map[string]bool, []loggingv1.Condition) {
	conditions := []loggingv1.Condition{}
	for _, migrate := range clfMigrations {
		var migrationConditions []loggingv1.Condition
		spec, extras, migrationConditions = migrate(namespace, name, spec, logStore, extras, logstoreSecretName, saTokenSecret)
		conditions = append(conditions, migrationConditions...)
	}
	return spec, extras, conditions
}

// migrations are the set of rules for migrating a ClusterLogForwarder that modify the spec
var clfMigrations = []func(namespace, name string, spec loggingv1.ClusterLogForwarderSpec, logStore *loggingv1.LogStoreSpec, extras map[string]bool, logstoreSecretName, saTokenSecret string) (loggingv1.ClusterLogForwarderSpec, map[string]bool, []loggingv1.Condition){
	clusterlogforwarder.EnsureInputsHasType,
	clusterlogforwarder.MigrateClusterLogForwarderSpec,
	clusterlogforwarder.MigrateInputs,
	clusterlogforwarder.DropUnreferencedOutputs,
}
