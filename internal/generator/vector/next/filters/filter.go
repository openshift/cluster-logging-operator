package filters

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
)

// Filter is a transformation of logs
type Filter interface {

	// Generate the configuration elements for the given pipeline
	Elements(inputs []string, pipeline logging.PipelineSpec, spec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element

	// TranformsName is the name of the transform in the config that is used as input to other sections
	TranformsName(pipeline logging.PipelineSpec) string
}
