package sources

import (
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

type SourceGenerator func(inputSources sets.String) (results []string, err error)

const fluentForwardSource = `
<source>
  @type forward
  port 24224
  bind 0.0.0.0
  @label @MEASURE
</source>
`
const localfluentForwardSource = `
<source>
  @type forward
  port 24224
  bind 127.0.0.1
  @label @MEASURE
</source>
`

//generateLocalFluentForwardSource creates a constant source config that assumes all log
//messages are received from a co-located collector
func GenerateLocalFluentForwardSource(sources sets.String) (results []string, err error) {
	return []string{localfluentForwardSource}, nil
}
func GenerateFluentForwardSource(sources sets.String) (results []string, err error) {
	return []string{fluentForwardSource}, nil
}

//GatherSources walks pipelines to collect all sources being used in forwarding
func GatherSources(forwarder *logging.ClusterLogForwarderSpec) (types, namespaces sets.String) {
	types, namespaces = sets.NewString(), sets.NewString()
	specs := forwarder.InputMap()
	for inputName := range logging.NewRoutes(forwarder.Pipelines).ByInput {
		if logging.ReservedInputNames.Has(inputName) {
			types.Insert(inputName) // Use name as type.
		} else if spec, ok := specs[inputName]; ok {
			if app := spec.Application; app != nil {
				types.Insert(logging.InputNameApplication)
				namespaces.Insert(app.Namespaces...)
			}
			if spec.Infrastructure != nil {
				types.Insert(logging.InputNameInfrastructure)
			}
			if spec.Audit != nil {
				types.Insert(logging.InputNameAudit)
			}
		}
	}
	return types, namespaces
}
