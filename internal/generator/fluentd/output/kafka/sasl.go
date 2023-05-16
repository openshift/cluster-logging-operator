package kafka

import (
	"fmt"
)

type SASL struct {
	Desc            string
	SaslOverSSL     bool
	SaslKeyPassword string
	ScramMechanism  string
}

func (s SASL) Name() string {
	return "kafkaSaslOverSSLTemplate"
}

func (s SASL) Template() string {
	conf := fmt.Sprintf(`{{define "kafkaSaslOverSSLTemplate" -}} 
sasl_over_ssl %t
`, s.SaslOverSSL)
	if s.SaslKeyPassword != "" {
		conf += "ssl_client_cert_key_password \"#{File.exists?({{.SaslKeyPassword}}) ? open({{.SaslKeyPassword}},'r') do |f|f.read end : ''}\"\n"
	}
	if s.ScramMechanism != "" {
		conf += fmt.Sprintf(`scram_mechanism "%s"`, s.ScramMechanism)
	}
	conf += `{{- end}}`
	return conf
}
