package cloudwatch

import (
	. "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
)

type Auth struct {
	KeyID           OptionalPair
	KeySecret       OptionalPair
	CredentialsPath OptionalPair
}

func NewAuth() Auth {
	return Auth{
		KeyID:           NewOptionalPair("auth.access_key_id", nil),
		KeySecret:       NewOptionalPair("auth.secret_access_key", nil),
		CredentialsPath: NewOptionalPair("auth.credentials_file", nil),
	}
}

func (a Auth) Name() string {
	return "awsAuthTemplate"
}

func (a Auth) Template() string {
	return `{{define "` + a.Name() + `" -}}
{{.KeyID}}
{{.KeySecret}}
{{.CredentialsPath}}
{{- end}}`
}
