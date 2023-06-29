package elasticsearch

import "github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"

type Passphrase security.Passphrase

func (p Passphrase) Name() string {
	return "passphraseTemplate"
}

func (p Passphrase) Template() string {
	return `{{define "` + p.Name() + `" -}}
client_key_pass "#{File.exists?({{.PassphrasePath}}) ? open({{.PassphrasePath}},'r') do |f|f.read end : ''}" 
{{- end}}
`
}
