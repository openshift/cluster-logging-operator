package cloudwatch

type AWSKey struct {
	KeyIDPath           string
	KeySecretPath       string
	KeyRoleArn          string
	KeyRoleSessionName  string
	KeyWebIdentityToken string
}

func (a AWSKey) Name() string {
	return "awsKeyTemplate"
}

func (a AWSKey) Template() string {
	// First check if we found a value for role
	if len(a.KeyRoleArn) > 0 {
		return `{{define "` + a.Name() + `" -}}
<web_identity_credentials>
  role_arn "{{ .KeyRoleArn }}"
  web_identity_token_file "{{ .KeyWebIdentityToken }}"
  role_session_name "{{ .KeyRoleSessionName }}"
</web_identity_credentials>
{{end}}`
	}
	// Otherwise use ID/Secret
	return `{{define "` + a.Name() + `" -}}
aws_key_id "#{open({{ .KeyIDPath }},'r') do |f|f.read.strip end}"
aws_sec_key "#{open({{ .KeySecretPath }},'r') do |f|f.read.strip end}"
{{end}}`
}
