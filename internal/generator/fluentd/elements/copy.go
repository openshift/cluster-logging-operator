package elements

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
)

type Copy struct {
	DeepCopy bool
	Stores   []framework.Element
}

func (c Copy) Name() string {
	return "copySourceTypeToPipeline"
}

func (c Copy) Template() string {
	return `{{define "` + c.Name() + `"  -}}
@type copy
{{if .DeepCopy -}}
copy_mode deep
{{end -}}
{{compose .Stores}}
{{end}}`
}

func CopyToLabels(labels []string) []framework.Element {
	s := []framework.Element{}
	for _, l := range labels {
		s = append(s, Store{
			Element: Relabel{
				OutLabel: l,
			},
		})
	}
	return s
}
