package fluentd

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	. "github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/cloudwatch"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/elasticsearch"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/fluentdforward"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/http"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/kafka"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/loki"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/syslog"
	"github.com/openshift/cluster-logging-operator/internal/logstore/lokistack"
	corev1 "k8s.io/api/core/v1"
)

func Outputs(clspec *logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, op Options) []Element {
	outputs := []Element{
		Comment("Ship logs to specific outputs"),
	}
	var bufspec *logging.FluentdBufferSpec = nil
	if clspec != nil &&
		clspec.Fluentd != nil &&
		clspec.Fluentd.Buffer != nil {
		bufspec = clspec.Fluentd.Buffer
	}
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
		if lokistack.DefaultLokiOuputNames.Has(o.Name) {
			outMinTlsVersion, outCiphers := op.TLSProfileInfo(clfspec.TLSSecurityProfile, o, ":")
			op[MinTLSVersion] = outMinTlsVersion
			op[Ciphers] = outCiphers
		}
		switch o.Type {
		case logging.OutputTypeElasticsearch:
			if _, ok := op[elements.CharEncoding]; !ok {
				op[elements.CharEncoding] = elements.DefaultCharEncoding
			}
			outputs = MergeElements(outputs, elasticsearch.Conf(bufspec, secret, o, op))
		case logging.OutputTypeFluentdForward:
			outputs = MergeElements(outputs, fluentdforward.Conf(bufspec, secret, o, op))
		case logging.OutputTypeKafka:
			outputs = MergeElements(outputs, kafka.Conf(bufspec, secret, o, op))
		case logging.OutputTypeCloudwatch:
			if _, ok := op[elements.CharEncoding]; !ok {
				op[elements.CharEncoding] = elements.DefaultCharEncoding
			}
			outputs = MergeElements(outputs, cloudwatch.Conf(bufspec, secret, o, op))
		case logging.OutputTypeSyslog:
			outputs = MergeElements(outputs, syslog.Conf(bufspec, secret, o, op))
		case logging.OutputTypeLoki:
			outputs = MergeElements(outputs, loki.Conf(bufspec, secret, o, op))
		case logging.OutputTypeHttp:
			outputs = MergeElements(outputs, http.Conf(bufspec, secret, o, op))
		}
	}

	return outputs
}
