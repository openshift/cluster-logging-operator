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

type Multiline struct {
	ComponentID string
	Desc        string
	Inputs      string
}

func (m Multiline) Name() string {
	return "multilineTemplate"
}

func (m Multiline) JavaReg() string {
	return "r'^\\\\d{4}-\\\\d{2}-\\\\d{2} .+'"
}

func (m Multiline) GoReg() string {
	return "r'\\bpanic:'"
}
func (m Multiline) GoReg2() string {
	return "r'http: panic serving'"
}

func (m Multiline) PythonReg() string {
	return "r'^Traceback \\\\(most recent call last\\\\):$/'"
}

func (m Multiline) RubyReg() string {
	return "r'/Error \\\\(.*\\\\):$/'"
}

func (m Multiline) Template() string {
	return `{{define "multilineTemplate" -}}
{{if .Desc -}}
# {{.Desc}}
{{end -}}
[transforms.{{.ComponentID}}]
type = "reduce"
inputs = {{.Inputs}}
starts_when.type = "vrl"
starts_when.source = "match_any(string!(.message), [{{.JavaReg}}, {{.RubyReg}}, {{.GoReg}}, {{.GoReg2}}, {{.PythonReg}}])"
merge_strategies.message = "concat_newline"
{{end}}
`
}
