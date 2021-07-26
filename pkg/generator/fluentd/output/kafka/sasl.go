package kafka

import "fmt"

type SaslOverSSL bool

func (s SaslOverSSL) Name() string {
	return "kafkaSaslOverSSLTemplate"
}

func (s SaslOverSSL) Template() string {
	return fmt.Sprintf(`{{define "kafkaSaslOverSSLTemplate" -}}
sasl_over_ssl %t
{{- end}}
`, s)
}
