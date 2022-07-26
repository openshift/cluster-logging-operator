package fluentdforward

type TLS struct {
	InsecureMode bool
}

func (t TLS) Name() string {
	return "fluentdforwardTLSTemplate"
}

func (t TLS) Template() string {
	return `{{define "fluentdforwardTLSTemplate" -}}
transport tls
tls_verify_hostname false
tls_version 'TLSv1_2'
{{- if .InsecureMode }}
tls_insecure_mode true
{{- end }}
{{- end }}
`
}
