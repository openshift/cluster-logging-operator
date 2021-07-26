package elasticsearch

import (
	"github.com/openshift/cluster-logging-operator/pkg/generator/fluentd/output/security"
)

type TLS security.TLS

func (t TLS) Name() string {
	return "elasticsearchTLSTemplate"
}

func (t TLS) Template() string {
	https := `{{define "elasticsearchTLSTemplate" -}}
scheme https
ssl_version TLSv1_2
{{- end}}
`
	http := `{{define "elasticsearchTLSTemplate" -}}
scheme http
{{- end}}
`
	if t {
		return https
	}
	return http
}
