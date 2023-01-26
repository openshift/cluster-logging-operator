package http

import (
	"github.com/openshift/cluster-logging-operator/internal/generator"
)

type BasicAuthConf generator.ConfLiteral

func (t BasicAuthConf) Name() string {
	return "httpBasicAuthConf"
}

func (t BasicAuthConf) Template() string {
	return `
{{define "httpBasicAuthConf" -}}
# {{.Desc}}
[sinks.{{.ComponentID}}.auth]
{{- end}}`
}
