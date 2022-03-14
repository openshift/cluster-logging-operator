package loki

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
)

type TLSKeyCert security.TLSCertKey

func (kc TLSKeyCert) Name() string {
	return "lokiCertKeyTemplate"
}

func (kc TLSKeyCert) Template() string {
	return `{{define "` + kc.Name() + `" -}}
key {{.KeyPath}}
cert {{.CertPath}}
{{- end}}`
}
