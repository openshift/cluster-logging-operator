package cloudwatch

type AWSKey struct {
	KeyIDPath string
	KeyPath   string
}

func (a AWSKey) Name() string {
	return "awsKeyTemplate"
}

func (a AWSKey) Template() string {
	return `{{define "` + a.Name() + `" -}}
aws_key_id "#{open({{ .KeyIDPath }},'r') do |f|f.read.strip end}"
aws_sec_key "#{open({{ .KeyPath }},'r') do |f|f.read.strip end}"
{{end}}`
}
