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

type insecureTLS struct {
	ComponentID string
}

func (i insecureTLS) Name() string {
	return "kafkaInsecureTLSTemplate"
}

func (i insecureTLS) Template() string {
	return `{{define "` + i.Name() + `" -}}
[sinks.{{.ComponentID}}.librdkafka_options]
"enable.ssl.certificate.verification" = "false"
{{- end}}`
}
