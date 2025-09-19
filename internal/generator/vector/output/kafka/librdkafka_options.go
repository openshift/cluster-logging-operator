package kafka

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"
)

type libRDKafkaOptions struct {
	ComponentID               string
	EnableSSLCertVerification helpers.OptionalPair
	MessageMaxBytes           helpers.OptionalPair
}

func (i libRDKafkaOptions) isEmpty() bool {
	return (i.EnableSSLCertVerification.Value == nil ||
		i.EnableSSLCertVerification.Value == true) &&
		i.MessageMaxBytes.Value == nil
}

func (i libRDKafkaOptions) Name() string {
	return "kafkaLibRDKafkaOptionsTemplate"
}

func (i libRDKafkaOptions) Template() string {
	if i.isEmpty() {
		return `{{define "` + i.Name() + `" -}}{{end}}`
	}
	return `{{define "` + i.Name() + `" -}}
[sinks.{{.ComponentID}}.librdkafka_options]
{{.EnableSSLCertVerification}}
{{.MessageMaxBytes}}
{{- end}}`
}

func newLibRDKafkaOptions(id string, o obs.OutputSpec, tuningSpec *obs.KafkaTuningSpec) *libRDKafkaOptions {
	formatter := framework.Option{Name: helpers.OptionFormatter, Value: `%q = "%v"`}
	libOptions := &libRDKafkaOptions{
		ComponentID:               id,
		EnableSSLCertVerification: helpers.NewOptionalPair("enable.ssl.certificate.verification", nil, formatter),
		MessageMaxBytes:           helpers.NewOptionalPair("message.max.bytes", nil, formatter),
	}
	if o.TLS != nil && isTlsBrokers(o) && o.TLS.InsecureSkipVerify {
		libOptions.EnableSSLCertVerification.Value = false
	}
	if tuningSpec != nil && tuningSpec.MaxWrite != nil {
		libOptions.MessageMaxBytes.Value = tuningSpec.MaxWrite.Value()
	}
	return libOptions
}
