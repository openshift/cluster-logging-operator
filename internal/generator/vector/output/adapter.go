package output

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	generator "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	nhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	corev1 "k8s.io/api/core/v1"
)

// Output is an adapter between CLF and Config generation
type Output struct {
	spec     obs.OutputSpec
	inputIDs []string
	op       generator.Options
	secrets  map[string]*corev1.Secret
	tuning   internalobs.Tuning
}

func (o Output) GetTuning() *internalobs.Tuning {
	return &o.tuning
}

func NewOutput(spec obs.OutputSpec, secrets map[string]*corev1.Secret, op generator.Options) *Output {
	return &Output{
		spec:    spec,
		op:      op,
		secrets: secrets,
		tuning:  internalobs.NewTuning(spec),
	}
}

func (o *Output) Elements() []generator.Element {
	if o == nil {
		return []generator.Element{}
	}
	el := New(o.spec, o.inputIDs, o.secrets, o, o.op, *o)
	return el
}

// AddInputFrom adds an input to an output regardless if the "input"
// originates directly from a log source or pipeline filter
func (o *Output) AddInputFrom(n nhelpers.InputComponent) {
	if o == nil {
		return
	}
	o.inputIDs = append(o.inputIDs, n.InputIDs()...)
}

func (o Output) Inputs() []string {
	return o.inputIDs
}
