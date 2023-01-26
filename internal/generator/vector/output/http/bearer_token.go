package http

import "github.com/openshift/cluster-logging-operator/internal/generator/vector/output/security"

type BearerToken security.BearerToken

func (bt BearerToken) Name() string {
	return "httpBearerTokenTemplate"
}

func (bt BearerToken) Template() string {
	return `{{define "` + bt.Name() + `" -}}
strategy = "bearer"
token = "{{.Token}}"
{{end}}
`
}
