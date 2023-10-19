package outputs

import (
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	rhes "github.com/openshift/cluster-logging-operator/internal/generator/vector/next/outputs/elasticsearch/rhinternal"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/next/registry"
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

func SinkInputsFromPipelinesOrFilters(spec *logging.ClusterLogForwarderSpec, op generator.Options) logging.RouteMap {
	r := logging.RouteMap{}
	filterSpecs := spec.FilterMap()
	for _, p := range spec.Pipelines {
		for _, o := range p.OutputRefs {
			if len(p.FilterRefs) > 0 {
				last := p.FilterRefs[len(p.FilterRefs)-1]
				if f := registry.LookupProto(last, filterSpecs); f != nil {
					for _, n := range f.TranformsNames(p) {
						r.Insert(o, n)
					}
				}
			} else {
				r.Insert(o, p.Name)
			}
		}
	}
	return r
}

func New(clspec *logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
	outputs := []generator.Element{}
	ofp := SinkInputsFromPipelinesOrFilters(clfspec, op)

	for idx, o := range clfspec.Outputs {
		var secret *corev1.Secret
		if s, ok := secrets[o.Name]; ok {
			secret = s
			log.V(9).Info("Using secret configured in output: " + o.Name)
		} else {
			secret = secrets[constants.LogCollectorToken]
			if secret != nil {
				log.V(9).Info("Using secret configured in " + constants.LogCollectorToken)
			} else {
				log.V(9).Info("No Secret found in " + constants.LogCollectorToken)
			}
		}

		if o.Name == logging.OutputNameDefault && o.Type == logging.OutputTypeElasticsearch {
			op[generator.MinTLSVersion] = ""
			op[generator.Ciphers] = ""
		} else {
			outMinTlsVersion, outCiphers := op.TLSProfileInfo(o, ",")
			op[generator.MinTLSVersion] = outMinTlsVersion
			op[generator.Ciphers] = outCiphers
		}

		inputs := ofp[o.Name].List()
		if o.HasPolicy() && o.GetMaxRecordsPerSecond() > 0 {
			// Vector Throttle component cannot have zero threshold
			outputs = append(outputs, AddThrottleForSink(&clfspec.Outputs[idx], inputs)...)
			inputs = []string{fmt.Sprintf(UserDefinedSinkThrottle, o.Name)}
		}

		if !o.HasPolicy() || (o.HasPolicy() && o.GetMaxRecordsPerSecond() > 0) {
			switch o.Type {
			case logging.OutputTypeKafka:
				outputs = generator.MergeElements(outputs, kafka.Conf(o, inputs, secret, op))
			case logging.OutputTypeLoki:
				outputs = generator.MergeElements(outputs, loki.Conf(o, inputs, secret, op))
			case logging.OutputTypeElasticsearch:
				if o.Name == logging.OutputNameDefault {
					outputs = generator.MergeElements(outputs, rhes.Conf(o, inputs, secret, op))
				} else {
					outputs = generator.MergeElements(outputs, elasticsearch.Conf(o, inputs, secret, op))
				}
			case logging.OutputTypeCloudwatch:
				outputs = generator.MergeElements(outputs, cloudwatch.Conf(o, inputs, secret, op))
			case logging.OutputTypeGoogleCloudLogging:
				outputs = generator.MergeElements(outputs, gcl.Conf(o, inputs, secret, op))
			case logging.OutputTypeSplunk:
				outputs = generator.MergeElements(outputs, splunk.Conf(o, inputs, secret, op))
			case logging.OutputTypeHttp:
				outputs = generator.MergeElements(outputs, http.Conf(o, inputs, secret, op))
			case logging.OutputTypeSyslog:
				outputs = generator.MergeElements(outputs, syslog.Conf(o, inputs, secret, op))
			}
		}
	}

	minTlsVersion, cipherSuites := op.TLSProfileInfo(logging.OutputSpec{}, ",")
	outputs = append(outputs,
		AddNodeNameToMetric(AddNodenameToMetricTransformName, []string{InternalMetricsSourceName}),
		PrometheusOutput(PrometheusOutputSinkName, []string{AddNodenameToMetricTransformName}, minTlsVersion, cipherSuites))
	return outputs
}

func MakeID(spec logging.OutputSpec) string {
	return fmt.Sprintf("output_%s", vectorhelpers.FormatComponentID(spec.Name))
}
