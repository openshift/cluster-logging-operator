package elasticsearch

type TLS bool
type EsTLS struct {
	TLS
	SSLVerify bool
}

func (t EsTLS) Name() string {
	return "elasticsearchTLSTemplate"
}

func (t EsTLS) Template() string {
	https := `{{define "elasticsearchTLSTemplate" -}}
scheme https
ssl_version TLSv1_2
{{- if not .SSLVerify }}
ssl_verify false
{{- end }}
{{- end}}
`
	http := `{{define "elasticsearchTLSTemplate" -}}
scheme http
{{- end}}
`
	if t.TLS {
		return https
	}
	return http
}
