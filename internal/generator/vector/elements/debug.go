package elements

import (
	"github.com/openshift/cluster-logging-operator/internal/generator"
)

func Debug(id string, inputs string) generator.Element {
	return generator.ConfLiteral{
		Desc:         "Sending records to stdout for debug purposes",
		ComponentID:  id,
		InLabel:      inputs,
		TemplateName: "debug",
		TemplateStr: `
{{define "debug" -}}
[sinks.{{.ComponentID}}]
inputs = {{.InLabel}}
type = "console"
target = "stdout"
encoding.codec = "json"
{{end}}
`,
	}
}
