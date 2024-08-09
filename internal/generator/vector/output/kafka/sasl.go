package kafka

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
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
enabled = true
username = "{{.Username}}"
password = "{{.Password}}"
mechanism = "{{.Mechanism}}"
{{end}}`
}

func SASLConf(id string, spec *obs.KafkaAuthentication, secrets observability.Secrets) Element {
	if spec != nil {
		saslAuth := spec.SASL
		if saslAuth != nil && saslAuth.Username != nil && saslAuth.Password != nil {
			sasl := SASL{
				ComponentID: id,
				Username:    vectorhelpers.SecretFrom(saslAuth.Username),
				Password:    vectorhelpers.SecretFrom(saslAuth.Password),
				Mechanism:   SASLMechanismPlain,
			}
			if saslAuth.Mechanism != "" {
				sasl.Mechanism = saslAuth.Mechanism
			}
			return sasl
		}
	}

	return Nil
}
