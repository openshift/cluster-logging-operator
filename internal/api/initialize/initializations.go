package initialize

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

const (

	// GeneratedSecrets identifies the list of secrets that may not exist but should be considered when configuration and
	// deploying the collector
	GeneratedSecrets = "generatedSecrets"
)

// clfInitializers are the set of rules for initializing the ClusterLogForwarder spec
var clfInitializers = []func(spec obs.ClusterLogForwarder, migrateContext utils.Options) obs.ClusterLogForwarder{
	MigrateLokiStack,
	MigrateInputs,
}

// ClusterLogForwarder initializes the forwarder for fields that must be set and are inferred from settings already defined.
// It additionally takes a context to pass decisions back such as secrets for receiver inputs that will exist after reconciliation
// but may not exist when the configuration is being generated
func ClusterLogForwarder(spec obs.ClusterLogForwarder, migrateContext utils.Options) obs.ClusterLogForwarder {
	for _, initialize := range clfInitializers {
		spec = initialize(spec, migrateContext)
	}
	return spec
}
