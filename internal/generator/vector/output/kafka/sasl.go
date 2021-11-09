package kafka

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/generator"
)

type Sasl bool

func (s Sasl) Name() string {
	return "kafkaSaslTemplate"
}

func (s Sasl) Template() string {
	return fmt.Sprintf(`{{define "kafkaSaslTemplate" -}}
enabled = %t
{{end}}
`, s)
}

type SaslConf generator.ConfLiteral

func (t SaslConf) Name() string {
	return "kafkaSasl"
}

func (t SaslConf) Template() string {
	return `
{{define "kafkaSasl" -}}
# {{.Desc}}
[sinks.{{.ComponentID}}.sasl]
{{- end}}`
}
