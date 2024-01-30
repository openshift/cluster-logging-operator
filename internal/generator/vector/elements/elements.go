package elements

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
)

type Route struct {
	ComponentID string
	Desc        string
	Inputs      string
	Routes      map[string]string
}

func (r Route) Name() string {
	return "routeTemplate"
}

func (r Route) Template() string {
	return `{{define "routeTemplate" -}}
{{if .Desc -}}
# {{.Desc}}
{{end -}}
[transforms.{{.ComponentID}}]
type = "route"
inputs = {{.Inputs}}
{{- $values := .Routes -}}
{{- $keys := getSortedKeyFromMap .Routes -}}
{{ range $route_name := $keys}}
route.{{$route_name}} = {{index $values $route_name}}
{{- end}}
{{end}}
`
}

type Filter struct {
	ComponentID string
	Desc        string
	Inputs      string
	Condition   string
}

func (r Filter) Name() string {
	return "filterTemplate"
}

func (r Filter) Template() string {
	return `{{define "filterTemplate" -}}
{{if .Desc -}}
# {{.Desc}}
{{end -}}
[transforms.{{.ComponentID}}]
type = "filter"
inputs = {{.Inputs}}
condition = '''
{{ .Condition }}
'''
{{end}}
`
}

type Remap struct {
	ComponentID string
	Desc        string
	Inputs      string
	VRL         string
}

func (r Remap) Name() string {
	return "remapTemplate"
}

func (r Remap) Template() string {
	return `{{define "remapTemplate" -}}
{{if .Desc -}}
# {{.Desc}}
{{end -}}
[transforms.{{.ComponentID}}]
type = "remap"
inputs = {{.Inputs}}
source = '''
{{.VRL | indent 2}}
'''
{{end}}
`
}

func Debug(id string, inputs string) framework.Element {
	return framework.ConfLiteral{
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
