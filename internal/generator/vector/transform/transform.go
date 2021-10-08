package transform

import "github.com/openshift/cluster-logging-operator/internal/generator"

type Transform interface {
	ComponentID() string
	SourceType() string
	TransformType() string
	generator.Element
}
