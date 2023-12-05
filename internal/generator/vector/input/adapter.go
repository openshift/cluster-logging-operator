package input

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
)

// Input is an adapter between CLF.input and any collector config segments
type Input struct {
	ids      []string
	elements []framework.Element
}

func NewInput(spec logging.InputSpec, collectorNS string, op framework.Options) *Input {
	elements, ids := NewViaQ(spec, collectorNS, op)
	return &Input{
		ids:      ids,
		elements: elements,
	}
}

func (i Input) Elements() []framework.Element {
	return i.elements
}

func (i Input) InputIDs() []string {
	return i.ids
}
