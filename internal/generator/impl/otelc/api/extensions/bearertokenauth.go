package extensions

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/types"
)

type BearerTokenAuth struct {
	id string

	Filename string `yaml:"filename,omitempty"`
}

func NewBearerTokenAuth(name, filename string) *BearerTokenAuth {
	return &BearerTokenAuth{
		id:       types.MakeComponentID(string(types.ExtensionTypeBearerTokenAuth), name),
		Filename: filename,
	}
}

func (e *BearerTokenAuth) ID() string {
	return e.id
}

func (e *BearerTokenAuth) ExtensionType() types.ExtensionType {
	return types.ExtensionTypeBearerTokenAuth
}
