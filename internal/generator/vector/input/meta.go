package input

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

func NewLogSourceAndType(id string, logSource, logType interface{}, inputs string, visit func(remap *elements.Remap)) framework.Element {
	ele := elements.Remap{
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs),
		VRL:         fmt.Sprintf(".log_source = %q\n.log_type = %q", logSource, logType),
	}
	if visit != nil {
		visit(&ele)
	}
	return ele
}
