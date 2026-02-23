package observability

import (
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	nhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

// Output is an internal representation of the public API Output
type Output struct {
	obsv1.OutputSpec
	InputIDs []string
	tuning   Tuning
}

func (o *Output) GetTlsSpec() *obsv1.TLSSpec {
	if o.TLS == nil {
		return nil
	}
	return &o.TLS.TLSSpec
}

func (o *Output) IsInsecureSkipVerify() bool {
	if o.TLS == nil {
		return false
	}
	return o.TLS.InsecureSkipVerify
}

func (o *Output) GetTuning() *Tuning {
	return &o.tuning
}

func NewOutput(spec obsv1.OutputSpec) *Output {
	return &Output{
		OutputSpec: spec,
		tuning:     NewTuning(spec),
	}
}

// AddInputFrom adds an input to an output regardless if the "input"
// originates directly from a log source or pipeline filter
func (o *Output) AddInputFrom(n nhelpers.InputComponent) {
	if o == nil {
		return
	}
	o.InputIDs = append(o.InputIDs, n.InputIDs()...)
}

func (o *Output) Inputs() []string {
	return o.InputIDs
}
