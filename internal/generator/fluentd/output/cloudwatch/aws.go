package cloudwatch

type AWSKey struct {
	KeyIDPath     string
	KeySecretPath string
	KeyRoleArn    string
}

func (a AWSKey) Name() string {
	return "awsKeyTemplate"
}

func (a AWSKey) Template() string {
	// First check if we found a value for role
	if len(a.KeyRoleArn) > 0 {
		return `{{define "` + a.Name() + `" -}}
<web_identity_credentials>
  role_arn "#{ENV['AWS_ROLE_ARN']}"
  web_identity_token_file "#{ENV['AWS_WEB_IDENTITY_TOKEN_FILE']}"
  role_session_name "#{ENV['AWS_ROLE_SESSION_NAME']}"
</web_identity_credentials>
{{end}}`
	}
	// Otherwise use ID/Secret
	return `{{define "` + a.Name() + `" -}}
aws_key_id "#{open({{ .KeyIDPath }},'r') do |f|f.read.strip end}"
aws_sec_key "#{open({{ .KeySecretPath }},'r') do |f|f.read.strip end}"
{{end}}`
}
