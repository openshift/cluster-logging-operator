package cloudwatch

import (
	"github.com/openshift/cluster-logging-operator/internal/generator"
)

type BasicAuthConf generator.ConfLiteral

func (t BasicAuthConf) Name() string {
	return "cloudwatchBasicAuthConf"
}

func (t BasicAuthConf) Template() string {
	return `
{{define "cloudwatchBasicAuthConf" -}}
# {{.Desc}}
[sinks.{{.ComponentID}}.auth]
{{- end}}`
}
