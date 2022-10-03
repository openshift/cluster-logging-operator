package syslog

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
)

type Passphrase security.Passphrase

func (p Passphrase) Name() string {
	return "passphraseTemplate"
}

func (p Passphrase) Template() string {
	return `{{define "` + p.Name() + `" -}}
client_cert_key_password "#{File.exists?({{.PassphrasePath}}) ? open({{.PassphrasePath}},'r') do |f|f.read end : ''}"
{{- end}}
`
}
