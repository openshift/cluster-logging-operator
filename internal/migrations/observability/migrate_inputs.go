package observability

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MigrateInputs creates instances of inputSpec when a pipeline references one of the reserved input names (e.g. application)
// This function is destructive and replaces any inputs that use the reserved input names
func MigrateInputs(spec obs.ClusterLogForwarderSpec) (obs.ClusterLogForwarderSpec, []metav1.Condition) {
	inputs := internalobs.Inputs(spec.Inputs).Map()
	for _, p := range spec.Pipelines {
		for _, i := range p.InputRefs {
			switch i {
			case string(obs.InputTypeApplication):
				inputs[string(obs.InputTypeApplication)] = obs.InputSpec{
					Type:        obs.InputTypeApplication,
					Name:        string(obs.InputTypeApplication),
					Application: &obs.Application{},
				}
			case string(obs.InputTypeInfrastructure):
				inputs[string(obs.InputTypeInfrastructure)] = obs.InputSpec{
					Type: obs.InputTypeInfrastructure,
					Name: string(obs.InputTypeInfrastructure),
					Infrastructure: &obs.Infrastructure{
						Sources: obs.InfrastructureSources,
					},
				}
			case string(obs.InputTypeAudit):
				inputs[string(obs.InputTypeAudit)] = obs.InputSpec{
					Type: obs.InputTypeAudit,
					Name: string(obs.InputTypeAudit),
					Audit: &obs.Audit{
						Sources: obs.AuditSources,
					},
				}
			}
		}
	}
	spec.Inputs = []obs.InputSpec{}
	for _, i := range inputs {
		spec.Inputs = append(spec.Inputs, i)
	}
	return spec, nil
}
