package elasticsearch

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
)

type BasicAuthConf framework.ConfLiteral

func (t BasicAuthConf) Name() string {
	return "elasticsearchBasicAuthConf"
}

func (t BasicAuthConf) Template() string {
	return `
{{define "elasticsearchBasicAuthConf" -}}
# {{.Desc}}
[sinks.{{.ComponentID}}.auth]
{{- end}}`
}
