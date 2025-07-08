package common

import (
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

type VectorSecret struct {
	framework.ComponentID
	Desc     string
	BasePath string
}

func (sec VectorSecret) Name() string {
	return "vector_secret_template"
}

func (sec VectorSecret) Template() string {
	return `{{define "` + sec.Name() + `" -}}
# {{.Desc}}
[secret.{{.ComponentID}}]
type = "directory"
path = "{{.BasePath}}"
{{end}}`
}

func NewVectorSecret() VectorSecret {
	secret := VectorSecret{
		ComponentID: helpers.VectorSecretID,
		Desc:        "Load sensitive data from files",
		BasePath:    constants.CollectorSecretsDir,
	}
	return secret
}
