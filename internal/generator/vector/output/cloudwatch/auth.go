package cloudwatch

type AWSKey struct {
	KeyID      string
	KeySecret  string
	KeyRoleArn string
}

func (a AWSKey) Name() string {
	return "awsKeyTemplate"
}

func (a AWSKey) Template() string {
	// First check if we found a value for role
	if len(a.KeyRoleArn) > 0 {
		return `{{define "` + a.Name() + `" -}}
# role_arn and identity token set via env vars
{{- end}}
`
	}
	return `{{define "` + a.Name() + `" -}}
auth.access_key_id = "{{.KeyID}}"
auth.secret_access_key = "{{.KeySecret}}"
{{- end}}
`
}
