package fluentd

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/cloudwatch"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/elasticsearch"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/fluentdforward"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/kafka"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/loki"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/syslog"
	corev1 "k8s.io/api/core/v1"
)

func Outputs(clspec *logging.ClusterLoggingSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, op Options) []Element {
	outputs := []Element{
		Comment("Ship logs to specific outputs"),
	}
	var bufspec *logging.FluentdBufferSpec = nil
	if clspec != nil &&
		clspec.Collection != nil &&
		clspec.Collection.FluentdSpec != nil &&
		clspec.Collection.FluentdSpec.Tuning != nil &&
		clspec.Collection.FluentdSpec.Tuning.Buffer != nil {
		bufspec = clspec.Collection.FluentdSpec.Tuning.Buffer
	}
	for _, o := range clfspec.Outputs {
		secret := secrets[o.Name]
		switch o.Type {
		case logging.OutputTypeElasticsearch:
			outputs = MergeElements(outputs, elasticsearch.Conf(bufspec, secret, o, op))
		case logging.OutputTypeFluentdForward:
			outputs = MergeElements(outputs, fluentdforward.Conf(bufspec, secret, o, op))
		case logging.OutputTypeKafka:
			outputs = MergeElements(outputs, kafka.Conf(bufspec, secret, o, op))
		case logging.OutputTypeCloudwatch:
			outputs = MergeElements(outputs, cloudwatch.Conf(bufspec, secret, o, op))
		case logging.OutputTypeSyslog:
			outputs = MergeElements(outputs, syslog.Conf(bufspec, secret, o, op))
		case logging.OutputTypeLoki:
			outputs = MergeElements(outputs, loki.Conf(bufspec, secret, o, op))
		}
	}

	return outputs
}
