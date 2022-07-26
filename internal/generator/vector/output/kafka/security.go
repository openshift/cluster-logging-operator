package kafka

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/security"
)

type Passphrase security.Passphrase

func (p Passphrase) Name() string {
	return "passphraseTemplate"
}

func (p Passphrase) Template() string {
	return `{{define "` + p.Name() + `" -}}
key_pass = "{{p.PassphrasePath}}"
{{- end}}
`
}
