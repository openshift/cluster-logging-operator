package conf

import (
	"fmt"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/cloudwatch"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/elasticsearch"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/gcl"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/http"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/kafka"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/loki"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/metrics"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/splunk"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/syslog"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
	corev1 "k8s.io/api/core/v1"
)

func OutputFromPipelines(spec *logging.ClusterLogForwarderSpec, op framework.Options) logging.RouteMap {
	r := logging.RouteMap{}
	for _, p := range spec.Pipelines {
		for _, o := range p.OutputRefs {
			r.Insert(o, p.Name)
		}
	}
	return r
}

func Outputs(clspec *logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, op framework.Options) []framework.Element {
	outputs := []framework.Element{}
	ofp := OutputFromPipelines(clfspec, op)

	for idx, o := range clfspec.Outputs {
		secret := helpers.GetOutputSecret(o, secrets)
		helpers.SetTLSProfileOptions(o, op)

		inputs := ofp[o.Name].List()
		if o.HasPolicy() && o.GetMaxRecordsPerSecond() > 0 {
			// Vector Throttle component cannot have zero threshold
			outputs = append(outputs, common.AddThrottleForSink(&clfspec.Outputs[idx], inputs)...)
			inputs = []string{fmt.Sprintf(common.UserDefinedSinkThrottle, o.Name)}
		}

		if !o.HasPolicy() || (o.HasPolicy() && o.GetMaxRecordsPerSecond() > 0) {
			switch o.Type {
			case logging.OutputTypeKafka:
				outputs = framework.MergeElements(outputs, kafka.Conf(o, inputs, secret, op))
			case logging.OutputTypeLoki:
				outputs = framework.MergeElements(outputs, loki.Conf(o, inputs, secret, op))
			case logging.OutputTypeElasticsearch:
				outputs = framework.MergeElements(outputs, elasticsearch.Conf(o, inputs, secret, op))
			case logging.OutputTypeCloudwatch:
				outputs = framework.MergeElements(outputs, cloudwatch.Conf(o, inputs, secret, op))
			case logging.OutputTypeGoogleCloudLogging:
				outputs = framework.MergeElements(outputs, gcl.Conf(o, inputs, secret, op))
			case logging.OutputTypeSplunk:
				outputs = framework.MergeElements(outputs, splunk.Conf(o, inputs, secret, op))
			case logging.OutputTypeHttp:
				outputs = framework.MergeElements(outputs, http.Conf(o, inputs, secret, op))
			case logging.OutputTypeSyslog:
				outputs = framework.MergeElements(outputs, syslog.Conf(o, inputs, secret, op))
			}
		}
	}

	minTlsVersion, cipherSuites := op.TLSProfileInfo(logging.OutputSpec{}, ",")
	outputs = append(outputs,
		metrics.AddNodeNameToMetric(metrics.AddNodenameToMetricTransformName, []string{source.InternalMetricsSourceName}),
		metrics.PrometheusOutput(metrics.PrometheusOutputSinkName, []string{metrics.AddNodenameToMetricTransformName}, minTlsVersion, cipherSuites))
	return outputs
}
