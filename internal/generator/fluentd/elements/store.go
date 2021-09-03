package elements

import (
	"github.com/openshift/cluster-logging-operator/internal/generator"
)

type Store struct {
	Element generator.Element
}

func (s Store) Name() string {
	return "storeTemplate"
}

func (s Store) Template() string {
	return `{{define "` + s.Name() + `" -}}
<store>
{{compose_one .Element| indent 2}}
</store>
{{end}}
`
}
