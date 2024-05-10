package kafka

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

const (
	SASLMechanismPlain = "PLAIN"
)

type SASL struct {
	ComponentID string
	Username    string
	Password    string
	Mechanism   string
}

func (t SASL) Name() string {
	return "vectorKafkaSasl"
}

func (t SASL) Template() string {
	return `{{define "vectorKafkaSasl"}}
[sinks.{{.ComponentID}}.sasl]
username = "{{.Username}}"
password = "{{.Password}}"
mechanism = "{{.Mechanism}}"
{{end}}`
}

func SASLConf(id string, spec *obs.KafkaAuthentication, secrets vectorhelpers.Secrets) Element {
	if spec != nil && spec.SASL != nil {
		sasl := SASL{
			ComponentID: id,
			Username:    secrets.AsString(spec.SASL.Username),
			Password:    secrets.AsString(spec.SASL.Password),
			Mechanism:   SASLMechanismPlain,
		}
		if spec.SASL.Mechanism != "" {
			sasl.Mechanism = spec.SASL.Mechanism
		}
		return sasl
	}

	return Nil
}
