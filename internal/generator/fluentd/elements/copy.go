package elements

import (
	"github.com/openshift/cluster-logging-operator/internal/generator"
)

type Copy struct {
	Stores []generator.Element
}

func (c Copy) Name() string {
	return "copySourceTypeToPipeline"
}

func (c Copy) Template() string {
	return `{{define "` + c.Name() + `"  -}}
@type copy
{{compose .Stores}}
{{end}}`
}

func CopyToLabels(labels []string) []generator.Element {
	s := []generator.Element{}
	for _, l := range labels {
		s = append(s, Store{
			Element: Relabel{
				OutLabel: l,
			},
		})
	}
	return s
}
