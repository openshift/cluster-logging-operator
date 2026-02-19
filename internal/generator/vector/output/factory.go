package output

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/adapters"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/aws/cloudwatch"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/aws/s3"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/azuremonitor"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
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

func New(o *adapters.Output, inputs []string, secrets map[string]*corev1.Secret, strategy common.ConfigStrategy, op utils.Options) []framework.Element {
	framwework.SetTLSProfileOptionsFrom(op, o.OutputSpec)

	var els []framework.Element
	baseID := helpers.MakeOutputID(o.Name)
	if threshold, hasPolicy := internalobs.Threshold(o.Limit); hasPolicy && threshold > 0 {
		// Vector Throttle component cannot have zero threshold
		throttleID := helpers.MakeID(baseID, "throttle")
		els = append(els, normalize.NewThrottle(throttleID, inputs, threshold, "")...)
		inputs = []string{throttleID}

	}

	switch o.Type {
	case obs.OutputTypeKafka:
		els = append(els, kafka.New(baseID, o, inputs, secrets, op)...)
	case obs.OutputTypeLoki:
		els = append(els, loki.New(baseID, o, inputs, secrets, op)...)
	case obs.OutputTypeLokiStack:
		els = append(els, lokistack.New(baseID, o, inputs, secrets, op)...)
	case obs.OutputTypeElasticsearch:
		els = append(els, elasticsearch.New(baseID, o, inputs, secrets, op)...)
	case obs.OutputTypeCloudwatch:
		els = append(els, cloudwatch.New(baseID, o, inputs, secrets, op)...)
	case obs.OutputTypeS3:
		els = append(els, s3.New(baseID, o, inputs, secrets, op)...)
	case obs.OutputTypeGoogleCloudLogging:
		els = append(els, gcl.New(baseID, o, inputs, secrets, op)...)
	case obs.OutputTypeSplunk:
		els = append(els, splunk.New(baseID, o, inputs, secrets, op)...)
	case obs.OutputTypeHTTP:
		els = append(els, http.New(baseID, o, inputs, secrets, op)...)
	case obs.OutputTypeSyslog:
		els = append(els, syslog.New(baseID, o.OutputSpec, inputs, secrets, strategy, op)...)
	case obs.OutputTypeAzureMonitor:
		els = append(els, azuremonitor.New(baseID, o, inputs, secrets, op)...)
	case obs.OutputTypeOTLP:
		els = append(els, otlp.New(baseID, o, inputs, secrets, op)...)
	}
	return els
}
