package elements

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
)

type Store struct {
	Element framework.Element
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
