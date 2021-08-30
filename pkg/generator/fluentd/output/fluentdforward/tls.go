package fluentdforward

import (
	"github.com/openshift/cluster-logging-operator/pkg/generator/fluentd/output/security"
)

type TLS security.TLS

func (t TLS) Name() string {
	return "fluentdforwardTLSTemplate"
}

func (t TLS) Template() string {
	https := `{{define "fluentdforwardTLSTemplate" -}}
transport tls
tls_verify_hostname false
tls_version 'TLSv1_2'
{{- end}}
`
	http := `{{define "fluentdforwardTLSTemplate" -}}
tls_insecure_mode true
{{- end}}
`
	if t {
		return https
	}
	return http
}
