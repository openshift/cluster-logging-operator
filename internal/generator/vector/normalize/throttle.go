package normalize

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

type Throttle struct {
	ComponentID string
	Desc        string
	Inputs      string
	Threshold   int64
	KeyField    string
}

func NewThrottle(id string, inputs []string, threshhold int64, throttleKey string) []framework.Element {
	el := []framework.Element{}

	el = append(el, Throttle{
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
		Threshold:   threshhold,
		KeyField:    throttleKey,
	})

	return el
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
