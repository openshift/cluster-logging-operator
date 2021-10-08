package kafka

import "fmt"

type SaslOverSSL bool

func (s SaslOverSSL) ID() string {
	return "kafkaSaslOverSSLTemplateID"
}

func (s SaslOverSSL) Name() string {
	return "kafkaSaslOverSSLTemplate"
}

func (s SaslOverSSL) Template() string {
	return fmt.Sprintf(`{{define "kafkaSaslOverSSLTemplate" -}}
sasl.enabled = %t
{{- end}}
`, s)
}
