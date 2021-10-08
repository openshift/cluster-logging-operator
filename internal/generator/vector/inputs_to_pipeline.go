package vector

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
	transform "github.com/openshift/cluster-logging-operator/internal/generator/vector/transform"
)

func InputsToPipeline(spec *logging.ClusterLogForwarderSpec, op Options, sources []source.LogSource) []transform.Transform {
 transformerList := []transform.Transform{}
 var instance transform.Transform

 for _, source := range sources {

  if source.Type() == "journald" {
    instance = transform.JournalTransform{
    SourceID:      helpers.PipelineName("transform_" + source.Type()),
    SrcType:       source.ComponentID(),
    InputPipeline: []string{"kubernetes_logs", helpers.PipelineName(source.Type())},
    TranType:      "route",
   }
  }
  if source.Type() == "kubernetes_logs" && source.ComponentID() == logging.InputNameApplication {
    instance = transform.KubernetesTransform{
    SourceID:      helpers.PipelineName("transform_" + source.Type()),
    SrcType:       source.ComponentID(),
    InputPipeline: []string{helpers.PipelineName(source.Type())},
    TranType:      "route",
   }
  }
  transformerList = append(transformerList, instance)
 }
return transformerList
}
