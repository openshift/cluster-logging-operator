package output

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/adapters"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/aws/cloudwatch"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/aws/s3"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/azuremonitor"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/elasticsearch"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/gcl"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/http"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/kafka"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/loki"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/lokistack"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/otlp"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/splunk"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/syslog"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
)

func New(o *adapters.Output, inputs []string, secrets map[string]*corev1.Secret, op utils.Options) (sinks api.Sinks, transforms api.Transforms) {
	transforms = api.Transforms{}
	sinks = api.Sinks{}

	framework.SetTLSProfileOptionsFrom(op, o)
	baseID := helpers.MakeOutputID(o.Name)
	if threshold, hasPolicy := internalobs.Threshold(o.Limit); hasPolicy && threshold > 0 {
		// Vector Throttle component cannot have zero threshold
		throttleID := helpers.MakeID(baseID, "throttle")
		transforms.Add(throttleID, common.NewThrottle(inputs, threshold, ""))
		inputs = []string{throttleID}
	}
	var sinkId string
	var sink types.Sink
	var sinkTransforms api.Transforms
	switch o.Type {
	case obs.OutputTypeKafka:
		sinkId, sink, sinkTransforms = kafka.New(baseID, o, inputs, secrets, op)
	case obs.OutputTypeLoki:
		sinkId, sink, sinkTransforms = loki.New(baseID, o, inputs, secrets, op)
	case obs.OutputTypeLokiStack:
		outputSinks, outputTransforms := lokistack.New(baseID, o, inputs, secrets, op)
		sinks.Merge(outputSinks)
		transforms.Merge(outputTransforms)
	case obs.OutputTypeElasticsearch:
		sinkId, sink, sinkTransforms = elasticsearch.New(baseID, o, inputs, secrets, op)
	case obs.OutputTypeCloudwatch:
		sinkId, sink, sinkTransforms = cloudwatch.New(baseID, o, inputs, secrets, op)
	case obs.OutputTypeS3:
		sinkId, sink, sinkTransforms = s3.New(baseID, o, inputs, secrets, op)
	case obs.OutputTypeGoogleCloudLogging:
		sinkId, sink, sinkTransforms = gcl.New(baseID, o, inputs, secrets, op)
	case obs.OutputTypeSplunk:
		sinkId, sink, sinkTransforms = splunk.New(baseID, o, inputs, secrets, op)
	case obs.OutputTypeHTTP:
		sinkId, sink, sinkTransforms = http.New(baseID, o, inputs, secrets, op)
	case obs.OutputTypeSyslog:
		sinkId, sink, sinkTransforms = syslog.New(baseID, o, inputs, secrets, op)
	case obs.OutputTypeAzureMonitor:
		sinkId, sink, sinkTransforms = azuremonitor.New(baseID, o, inputs, secrets, op)
	case obs.OutputTypeOTLP:
		sinkId, sink, sinkTransforms = otlp.New(baseID, o, inputs, secrets, op)
	}

	if sinkId != "" {
		sinks.Add(sinkId, sink)
		transforms.Merge(sinkTransforms)
	}

	return sinks, transforms
}
