package elasticsearch

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/security"
)

type TLSKeyCert security.TLSCertKey

func (kc TLSKeyCert) Name() string {
	return "elasticsearchCertKeyTemplate"
}

func (kc TLSKeyCert) Template() string {
	return `{{define "` + kc.Name() + `" -}}
key_file = {{.KeyPath}}
crt_file = {{.CertPath}}
{{- end}}`
}
