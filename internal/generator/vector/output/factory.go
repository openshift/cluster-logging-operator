package output

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/azuremonitor"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/cloudwatch"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/elasticsearch"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/gcl"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/http"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/kafka"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/loki"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/splunk"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/syslog"
	corev1 "k8s.io/api/core/v1"
)

func New(o obs.OutputSpec, inputs []string, secrets map[string]*corev1.Secret, strategy common.ConfigStrategy, op Options) []Element {
	op.SetTLSProfileOptionsFrom(o)

	var els []Element
	baseID := helpers.MakeOutputID(o.Name)
	if threshold, hasPolicy := internalobs.Threshold(o.Limit); hasPolicy && threshold > 0 {
		// Vector Throttle component cannot have zero threshold
		throttleID := helpers.MakeID(baseID, "throttle")
		els = append(els, normalize.NewThrottle(throttleID, inputs, threshold, "")...)
		inputs = []string{throttleID}

	}

	switch o.Type {
	case logging.OutputTypeKafka:
		els = append(els, kafka.New(baseID, o, inputs, secrets, strategy, op)...)
	case logging.OutputTypeLoki:
		els = append(els, loki.New(baseID, o, inputs, secrets, strategy, op)...)
	case logging.OutputTypeElasticsearch:
		els = append(els, elasticsearch.New(baseID, o, inputs, secrets, strategy, op)...)
	case logging.OutputTypeCloudwatch:
		els = append(els, cloudwatch.New(baseID, o, inputs, secrets, strategy, op)...)
	case logging.OutputTypeGoogleCloudLogging:
		els = append(els, gcl.New(baseID, o, inputs, secrets, strategy, op)...)
	case logging.OutputTypeSplunk:
		els = append(els, splunk.New(baseID, o, inputs, secrets, strategy, op)...)
	case logging.OutputTypeHttp:
		els = append(els, http.New(baseID, o, inputs, secrets, strategy, op)...)
	case logging.OutputTypeSyslog:
		els = append(els, syslog.New(baseID, o, inputs, secrets, strategy, op)...)
	case logging.OutputTypeAzureMonitor:
		els = append(els, azuremonitor.New(baseID, o, inputs, secrets, strategy, op)...)
	}
	return els
}
