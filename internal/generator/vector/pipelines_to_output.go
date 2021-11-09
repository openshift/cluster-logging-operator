package vector

import (
	"fmt"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
)

func PipelineToOutputs(spec *logging.ClusterLogForwarderSpec, op generator.Options) logging.RouteMap {
	r := logging.RouteMap{}
	for _, p := range spec.Pipelines {
		for _, o := range p.OutputRefs {
			fmt.Printf("Adding %s to %s\n", o, p.Name)
			r.Insert(o, p.Name)
		}
	}
	return r
}
