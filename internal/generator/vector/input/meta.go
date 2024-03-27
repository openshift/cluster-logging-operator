package input

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

func NewLogSourceAndType(id, logSource, logType string, inputs string) framework.Element {
	return elements.Remap{
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs),
		VRL:         fmt.Sprintf(".log_source = %q\n.log_type = %q", logSource, logType),
	}
}
