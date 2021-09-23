package syslog

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
)

type TLS security.TLS

func (t TLS) Name() string {
	return "syslogTLSTemplate"
}

func (t TLS) Template() string {
	return fmt.Sprintf(`{{define "syslogTLSTemplate" -}}
tls %t
{{- end}}`, t)
}

type HostnameVerify security.HostnameVerify

func (h HostnameVerify) Name() string {
	return "syslogHostNameVerify"
}

func (h HostnameVerify) Template() string {
	if h {
		return `{{define "` + h.Name() + `" -}}
verify_mode 1 #VERIFY_NONE:0, VERIFY_PEER:1
{{- end}}
`
	}
	return `{{define "` + h.Name() + `" -}}
verify_mode 0 #VERIFY_NONE:0, VERIFY_PEER:1
{{- end}}
`
}
