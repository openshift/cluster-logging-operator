package http

type Passphrase struct {
	Passphrase string
}

func (p Passphrase) Name() string {
	return "passphraseTemplate"
}

func (p Passphrase) Template() string {
	return `{{define "` + p.Name() + `" -}}
tls_client_private_key_passphrase "{{.Passphrase}}" 
{{- end}}
`
}
