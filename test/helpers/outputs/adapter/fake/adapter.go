package fake

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	generator "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
)

// Output is an adapter between CLF and Config generation
type Output struct {
	obs.OutputSpec
	op      generator.Options
	secrets map[string]*corev1.Secret
	tuning  internalobs.Tuning
}

func NewOutput(spec obs.OutputSpec, secrets map[string]*corev1.Secret, op utils.Options) *Output {
	return &Output{
		OutputSpec: spec,
		op:         op,
		secrets:    secrets,
		tuning:     internalobs.NewTuning(spec),
	}
}

func (o Output) GetTuning() *internalobs.Tuning {
	return &o.tuning
}

func (o Output) GetTlsSpec() *obs.TLSSpec {
	if o.TLS == nil {
		return nil
	}
	return &o.TLS.TLSSpec
}

func (o Output) IsInsecureSkipVerify() bool {
	if o.TLS == nil {
		return false
	}
	return o.TLS.InsecureSkipVerify
}
