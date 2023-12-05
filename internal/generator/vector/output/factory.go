package output

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/cloudwatch"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/elasticsearch"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/gcl"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/http"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/kafka"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/loki"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/splunk"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/syslog"
	corev1 "k8s.io/api/core/v1"
)

func New(o logging.OutputSpec, inputs []string, secrets map[string]*corev1.Secret, op Options) []Element {
	secret := helpers.GetOutputSecret(o, secrets)
	helpers.SetTLSProfileOptions(o, op)

	var els []Element
	baseID := helpers.MakeOutputID(o.Name)
	if o.HasPolicy() && o.GetMaxRecordsPerSecond() > 0 {
		// Vector Throttle component cannot have zero threshold
		throttleID := helpers.MakeID(baseID, "throttle")
		els = append(els, normalize.NewThrottle(throttleID, inputs, o.GetMaxRecordsPerSecond(), "")...)
		inputs = []string{throttleID}
	}

	switch o.Type {
	case logging.OutputTypeKafka:
		els = append(els, kafka.New(baseID, o, inputs, secret, op)...)
	case logging.OutputTypeLoki:
		els = append(els, loki.New(baseID, o, inputs, secret, op)...)
	case logging.OutputTypeElasticsearch:
		els = append(els, elasticsearch.New(baseID, o, inputs, secret, op)...)
	case logging.OutputTypeCloudwatch:
		els = append(els, cloudwatch.New(baseID, o, inputs, secret, op)...)
	case logging.OutputTypeGoogleCloudLogging:
		els = append(els, gcl.New(baseID, o, inputs, secret, op)...)
	case logging.OutputTypeSplunk:
		els = append(els, splunk.New(baseID, o, inputs, secret, op)...)
	case logging.OutputTypeHttp:
		els = append(els, http.New(baseID, o, inputs, secret, op)...)
	case logging.OutputTypeSyslog:
		els = append(els, syslog.New(baseID, o, inputs, secret, op)...)
	}
	return els
}
