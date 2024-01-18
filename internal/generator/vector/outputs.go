package vector

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/logstore/lokistack"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"

	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
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

func OutputFromPipelines(spec *logging.ClusterLogForwarderSpec, op generator.Options) logging.RouteMap {
	r := logging.RouteMap{}
	for _, p := range spec.Pipelines {
		for _, o := range p.OutputRefs {
			r.Insert(o, p.Name)
		}
	}
	return r
}

func AddThrottleForSink(spec *logging.OutputSpec, inputs []string) []generator.Element {
	el := []generator.Element{}

	el = append(el, normalize.Throttle{
		ComponentID: fmt.Sprintf(UserDefinedSinkThrottle, spec.Name),
		Inputs:      helpers.MakeInputs(inputs...),
		Threshold:   spec.Limit.MaxRecordsPerSecond,
		KeyField:    "",
	})

	return el
}

func Outputs(clspec *logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
	outputs := []generator.Element{}
	ofp := OutputFromPipelines(clfspec, op)

	for idx, o := range clfspec.Outputs {
		var secret *corev1.Secret
		if s := getSecret(o, secrets); s != nil {
			secret = s
			log.V(9).Info("Using secret configured in output: " + o.Name)
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
				outputs = generator.MergeElements(outputs, elasticsearch.Conf(o, inputs, secret, op))
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

func getSecret(output logging.OutputSpec, secrets map[string]*corev1.Secret) *corev1.Secret {
	key := output.Name
	if output.Type == logging.OutputTypeLoki && lokistack.DefaultLokiOutputNames.Has(output.Name) {
		key = constants.LogCollectorToken
	}
	return secrets[key]
}

func PrometheusOutput(id string, inputs []string, minTlsVersion string, cipherSuites string) generator.Element {
	return PrometheusExporter{
		ID:            id,
		Inputs:        helpers.MakeInputs(inputs...),
		Address:       helpers.ListenOnAllLocalInterfacesAddress() + `:` + PrometheusExporterListenPort,
		TlsMinVersion: minTlsVersion,
		CipherSuites:  cipherSuites,
	}
}

func AddNodeNameToMetric(id string, inputs []string) generator.Element {
	return AddNodenameToMetric{
		ID:     id,
		Inputs: helpers.MakeInputs(inputs...),
	}
}
