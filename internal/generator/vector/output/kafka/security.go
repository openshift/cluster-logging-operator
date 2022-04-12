package kafka

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/security"
)

type TLS security.TLS

func (t TLS) Name() string {
	return "kafkaTLSTemplate"
}

func (t TLS) Template() string {
	return fmt.Sprintf(`{{define "kafkaTLSTemplate" -}}
enabled = %t
{{end}}`, t)
}

type TLSInsecure bool

func (i TLSInsecure) Name() string { return "kafkaInsecureSkipVerifyTemplate" }
func (i TLSInsecure) Template() string {
	return fmt.Sprintf(`{{define %q -}}
verify_certificate = %t
verify_hostname = %t
{{- end}}`, i.Name(), !i, !i)
}

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
