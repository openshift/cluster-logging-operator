package cloudwatch

type AWSKey struct {
	KeyID     string
	KeySecret string
}

func (a AWSKey) Name() string {
	return "awsKeyTemplate"
}

func (a AWSKey) Template() string {
	return `{{define "` + a.Name() + `" -}}
auth.access_key_id = "{{.KeyID}}"
auth.secret_access_key = "{{.KeySecret}}"
{{- end}}
`
}
