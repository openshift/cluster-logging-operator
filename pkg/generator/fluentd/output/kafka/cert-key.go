package kafka

import (
	"github.com/openshift/cluster-logging-operator/pkg/generator/fluentd/output/security"
)

type TLSKeyCert security.TLSCertKey

func (kc TLSKeyCert) Name() string {
	return "kafkaCertKeyTemplate"
}

func (kc TLSKeyCert) Template() string {
	return `{{define "` + kc.Name() + `" -}}
ssl_client_cert_key {{.KeyPath}}
ssl_client_cert {{.CertPath}}
{{- end}}`
}
