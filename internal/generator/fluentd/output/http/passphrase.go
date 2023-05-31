package http

import "github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"

type Passphrase security.Passphrase

func (p Passphrase) Name() string {
	return "passphraseTemplate"
}

func (p Passphrase) Template() string {
	return `{{define "` + p.Name() + `" -}}
tls_private_key_passphrase "#{File.exists?({{.PassphrasePath}}) ? open({{.PassphrasePath}},'r') do |f|f.read end : ''}" 
{{- end}}
`
}
