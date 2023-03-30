package elements

import (
	"github.com/openshift/cluster-logging-operator/internal/generator"
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
{{- range $route_name, $route_expr := .Routes}}
route.{{$route_name}} = {{$route_expr}}
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
condition = "{{ .Condition }}"
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

type DetectExceptions struct {
	ComponentID string
	Inputs      string
}

func (d DetectExceptions) Name() string{
  return "detectExceptions"
}

func (d DetectExceptions) Template() string{
  return `{{define "detectExceptions" -}}
[transforms.{{.ComponentID}}]
type = "detect_exceptions"
inputs = {{.Inputs}}
languages = ["All"]
group_by = ["kubernetes.namespace_name","kubernetes.pod_name","kubernetes.container_name", "kubernetes.pod_id"]
expire_after_ms = 2000
multiline_flush_interval_ms = 1000
{{end}}`
}

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
