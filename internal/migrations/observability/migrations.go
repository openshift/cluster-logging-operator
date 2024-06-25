package observability

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (

	// GeneratedSecrets identifies the list of secrets that may not exist but should be considered when configuration and
	// deploying the collector
	GeneratedSecrets = "generatedSecrets"
)

// TODO: Refactor other migrations to observability
// migrations are the set of rules for migrating a ClusterLogForwarder that modify the spec
var clfMigrations = []func(spec obs.ClusterLogForwarder, migrateContext utils.Options) (obs.ClusterLogForwarder, []metav1.Condition){
	// clusterlogforwarder.EnsureInputsHasType,
	MigrateLokiStack,
	MigrateInputs,
	// clusterlogforwarder.DropUnreferencedOutputs,
}

// MigrateClusterLogForwarder initializes the forwarder for fields that must be set and are inferred from settings already defined.
// It additionally takes a context to pass decisions back such as secrets for receiver inputs that will exist after reconciliation
// but may not exist when the configuration is being generated
func MigrateClusterLogForwarder(spec obs.ClusterLogForwarder, migrateContext utils.Options) (obs.ClusterLogForwarder, []metav1.Condition) {
	conditions := []metav1.Condition{}
	for _, migrate := range clfMigrations {
		var migrationConditions []metav1.Condition
		spec, migrationConditions = migrate(spec, migrateContext)
		conditions = append(conditions, migrationConditions...)
	}
	return spec, conditions
}
