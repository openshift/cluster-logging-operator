package cloudwatch

type AWSKey struct {
	AWSAccessKeyID     string
	AWSSecretAccessKey string
}

func (a AWSKey) Name() string {
	return "awsKeyTemplate"
}

func (a AWSKey) Template() string {
	return `{{define "` + a.Name() + `" -}}
auth.access_key_id = "{{.AWSAccessKeyID}}"
auth.secret_access_key = "{{.AWSSecretAccessKey}}"
{{- end}}`
}
