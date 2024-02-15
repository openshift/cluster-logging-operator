package output

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	generator "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	nhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	corev1 "k8s.io/api/core/v1"
)

// Output is an adapter between CLF and Config generation
type Output struct {
	spec     logging.OutputSpec
	inputIDs []string
	op       generator.Options
	secrets  map[string]*corev1.Secret
}

func NewOutput(spec logging.OutputSpec, secrets map[string]*corev1.Secret, op generator.Options) *Output {
	return &Output{
		spec:    spec,
		op:      op,
		secrets: secrets,
	}
}

func (o *Output) Elements() []generator.Element {
	el := New(o.spec, o.inputIDs, o.secrets, o, o.op)
	return el
}

// AddInputFrom adds an input to an output regardless if the "input"
// originates directly from a log source or pipeline filter
func (o *Output) AddInputFrom(n nhelpers.InputComponent) {
	o.inputIDs = append(o.inputIDs, n.InputIDs()...)
}

func (o Output) Inputs() []string {
	return o.inputIDs
}
