package adapters

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

// InternalFilterSpec is a wrapper to allow separation of public and internal filters
type InternalFilterSpec struct {
	*obs.FilterSpec

	// Factory creates a new instance of a transform
	Factory func(inputs ...string) types.Transform
}
