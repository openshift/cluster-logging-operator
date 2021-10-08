package kafka

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
)

type TLSKeyCert security.TLSCertKey

func (kc TLSKeyCert) Name() string {
	return "kafkaCertKeyTemplate"
}

func (kc TLSKeyCert) Template() string {
	return `{{define "` + kc.Name() + `" -}}
tls.key_file = {{.KeyPath}}
tls.crt_file = {{.CertPath}}
{{- end}}`
}
