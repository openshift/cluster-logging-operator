package elements

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
