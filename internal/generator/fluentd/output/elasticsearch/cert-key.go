package elasticsearch

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
)

type TLSKeyCert security.TLSCertKey

func (kc TLSKeyCert) Name() string {
	return "elasticsearchCertKeyTemplate"
}

func (kc TLSKeyCert) Template() string {
	return `{{define "` + kc.Name() + `" -}}
client_key {{.KeyPath}}
client_cert {{.CertPath}}
{{- end}}`
}
