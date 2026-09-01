package elements

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
