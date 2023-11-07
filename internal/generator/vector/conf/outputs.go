package conf

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
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

var (
	SinkTransformThrottle = "sink_throttle"

	UserDefinedSinkThrottle = fmt.Sprintf(`%s_%%s`, SinkTransformThrottle)
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

func AddThrottleForSink(spec *logging.OutputSpec, inputs []string) []framework.Element {
	el := []framework.Element{}

	el = append(el, normalize.Throttle{
		ComponentID: fmt.Sprintf(UserDefinedSinkThrottle, spec.Name),
		Inputs:      helpers.MakeInputs(inputs...),
		Threshold:   spec.Limit.MaxRecordsPerSecond,
		KeyField:    "",
	})

	return el
}

func Outputs(clspec *logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, op framework.Options) []framework.Element {
	outputs := []framework.Element{}
	ofp := OutputFromPipelines(clfspec, op)

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
			op[framework.MinTLSVersion] = ""
			op[framework.Ciphers] = ""
		} else {
			outMinTlsVersion, outCiphers := op.TLSProfileInfo(o, ",")
			op[framework.MinTLSVersion] = outMinTlsVersion
			op[framework.Ciphers] = outCiphers
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
		AddNodeNameToMetric(source.AddNodenameToMetricTransformName, []string{source.InternalMetricsSourceName}),
		PrometheusOutput(source.PrometheusOutputSinkName, []string{source.AddNodenameToMetricTransformName}, minTlsVersion, cipherSuites))
	return outputs
}

func PrometheusOutput(id string, inputs []string, minTlsVersion string, cipherSuites string) framework.Element {
	return source.PrometheusExporter{
		ID:            id,
		Inputs:        helpers.MakeInputs(inputs...),
		Address:       helpers.ListenOnAllLocalInterfacesAddress() + `:` + source.PrometheusExporterListenPort,
		TlsMinVersion: minTlsVersion,
		CipherSuites:  cipherSuites,
	}
}

func AddNodeNameToMetric(id string, inputs []string) framework.Element {
	return source.AddNodenameToMetric{
		ID:     id,
		Inputs: helpers.MakeInputs(inputs...),
	}
}
