package source

import "github.com/openshift/cluster-logging-operator/internal/generator"

type LogSource interface {
	ComponentID() string
	Type() string
	generator.Element
}
