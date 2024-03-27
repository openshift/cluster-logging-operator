package input

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
)

// Input is an adapter between CLF.input and any collector config segments
type Input struct {
	spec     logging.InputSpec
	ids      []string
	elements []framework.Element
}

func NewInput(spec logging.InputSpec, collectorNS string, resNames *factory.ForwarderResourceNames, op framework.Options) *Input {
	elements, ids := NewSource(spec, collectorNS, resNames, op)
	return &Input{
		spec:     spec,
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

// Add is a convenience function to concat elements and ids
func (i *Input) Add(elements []framework.Element, ids []string) *Input {
	i.ids = append(i.ids, ids...)
	i.elements = append(i.elements, elements...)
	return i
}
