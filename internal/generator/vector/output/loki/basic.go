package loki

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
)

type BasicAuthConf framework.ConfLiteral

func (t BasicAuthConf) Name() string {
	return "lokiBasicAuthConf"
}

func (t BasicAuthConf) Template() string {
	return `
{{define "lokiBasicAuthConf" -}}
# {{.Desc}}
[sinks.{{.ComponentID}}.auth]
{{- end}}`
}
