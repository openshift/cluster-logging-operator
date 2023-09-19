package normalize

type Throttle struct {
	ComponentID string
	Desc        string
	Inputs      string
	Threshold   int64
	KeyField    string
}

func (t Throttle) Name() string {
	return "throttleTemplate"
}

func (t Throttle) Template() string {
	return `
{{define "throttleTemplate" -}}
{{- if .Desc}}
# {{.Desc}}
{{- end}}
[transforms.{{.ComponentID}}]
type = "throttle"
inputs = {{.Inputs}}
window_secs = 1
threshold = {{.Threshold}}
{{- if .KeyField}}
key_field = {{ .KeyField }}
{{- end}}
{{end}}
`
}
