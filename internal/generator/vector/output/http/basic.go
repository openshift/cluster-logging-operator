package http

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
)

type BasicAuthConf framework.ConfLiteral

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
