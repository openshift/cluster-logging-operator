package vector

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
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
	"github.com/openshift/cluster-logging-operator/internal/logstore/lokistack"
	corev1 "k8s.io/api/core/v1"
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

func Outputs(clspec *logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
	outputs := []generator.Element{}
	ofp := OutputFromPipelines(clfspec, op)

	for _, o := range clfspec.Outputs {
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

		if op.Has(constants.PreviewTLSSecurityProfile) || lokistack.DefaultLokiOuputNames.Has(o.Name) {
			if o.Name == logging.OutputNameDefault && o.Type == logging.OutputTypeElasticsearch {
				op[generator.MinTLSVersion] = ""
				op[generator.Ciphers] = ""
			} else {
				outMinTlsVersion, outCiphers := op.TLSProfileInfo(clfspec.TLSSecurityProfile, o, ",")
				op[generator.MinTLSVersion] = outMinTlsVersion
				op[generator.Ciphers] = outCiphers
			}
		}

		inputs := ofp[o.Name].List()
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

	minTlsVersion, cipherSuites := op.TLSProfileInfo(nil, logging.OutputSpec{}, ",")
	outputs = append(outputs,
		AddNodeNameToMetric(AddNodenameToMetricTransformName, []string{InternalMetricsSourceName}),
		PrometheusOutput(PrometheusOutputSinkName, []string{AddNodenameToMetricTransformName}, minTlsVersion, cipherSuites))
	return outputs
}

func PrometheusOutput(id string, inputs []string, minTlsVersion string, cipherSuites string) generator.Element {
	return PrometheusExporter{
		ID:            id,
		Inputs:        helpers.MakeInputs(inputs...),
		Address:       PrometheusExporterAddress,
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
