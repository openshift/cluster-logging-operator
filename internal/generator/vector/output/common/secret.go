package common

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
)

type VectorSecret struct {
	framework.ComponentID
	Desc    string
	Command string
	Timeout int
}

func (sec VectorSecret) Name() string {
	return "vector_secret_template"
}

func (sec VectorSecret) Template() string {
	return `{{define "` + sec.Name() + `" -}}
# {{.Desc}}
[secret.{{.ComponentID}}]
type = "exec"
command = ["sh", "{{.Command}}"]
{{end}}`
}

func NewVectorSecret(id, command string) VectorSecret {

	secret := VectorSecret{
		ComponentID: id,
		Desc:        fmt.Sprintf("Load sensitive data from secret mount with script: %s", command),
		Command:     command,
	}
	return secret
}
