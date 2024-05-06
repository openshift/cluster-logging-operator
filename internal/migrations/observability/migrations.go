package observability

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: Refactor other migrations to observability
// migrations are the set of rules for migrating a ClusterLogForwarder that modify the spec
var clfMigrations = []func(spec obs.ClusterLogForwarderSpec) (obs.ClusterLogForwarderSpec, []metav1.Condition){
	// clusterlogforwarder.EnsureInputsHasType,
	MigrateLokiStack,
	// clusterlogforwarder.MigrateInputs,
	// clusterlogforwarder.DropUnreferencedOutputs,
}

func MigrateClusterLogForwarder(spec obs.ClusterLogForwarderSpec) (obs.ClusterLogForwarderSpec, []metav1.Condition) {
	conditions := []metav1.Condition{}
	for _, migrate := range clfMigrations {
		var migrationConditions []metav1.Condition
		spec, migrationConditions = migrate(spec)
		conditions = append(conditions, migrationConditions...)
	}
	return spec, conditions
}
