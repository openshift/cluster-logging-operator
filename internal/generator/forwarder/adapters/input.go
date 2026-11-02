package adapters

import (
	openshiftv1 "github.com/openshift/api/config/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
)

// Input is an internal representation of the public API input
type Input struct {
	obs.InputSpec
	Ids []string
}

func (i *Input) InputIDs() []string {
	return i.Ids
}

func (i *Input) GetTlsSpec() *obs.TLSSpec {
	if i.Receiver == nil || i.Receiver.TLS == nil {
		return nil
	}
	tlsSpec := obs.TLSSpec(*i.Receiver.TLS)
	return &tlsSpec
}

func (i *Input) GetTlsSecurityProfile() *openshiftv1.TLSSecurityProfile {
	return nil
}

func (i *Input) IsInsecureSkipVerify() bool {
	return false
}

func NewInput(spec obs.InputSpec) *Input {
	i := Input{
		InputSpec: spec,
	}
	return &i
}

func (i *Input) IsContainerSource() bool {
	return internalobs.Inputs([]obs.InputSpec{i.InputSpec}).HasContainerSource()
}
