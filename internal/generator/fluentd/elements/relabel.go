package elements

import (
	"github.com/openshift/cluster-logging-operator/internal/generator"
)

type Relabel struct {
	generator.OutLabel
}

func (r Relabel) Name() string {
	return "relabel"
}

func (r Relabel) Template() string {
	return `{{define "` + r.Name() + `"  -}}
@type relabel
@label {{.OutLabel}}
{{end}}`
}
