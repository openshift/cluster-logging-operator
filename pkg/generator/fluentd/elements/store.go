package elements

import (
	. "github.com/openshift/cluster-logging-operator/pkg/generator"
)

type Store struct {
	Element Element
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
