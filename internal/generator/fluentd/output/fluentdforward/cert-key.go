package fluentdforward

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
)

type TLSCertKey security.TLSCertKey

func (kc TLSCertKey) Name() string {
	return "fluentdforwardCertKeyTemplate"
}

func (kc TLSCertKey) Template() string {
	return `{{define "` + kc.Name() + `" -}}
tls_client_private_key_path {{.KeyPath}}
tls_client_cert_path {{.CertPath}}
{{- end}}`
}
