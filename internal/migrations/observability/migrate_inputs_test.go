package observability

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
)

var _ = Describe("migrateInputs", func() {
	Context("for reserved input names", func() {

		It("should stub 'application' as an input when referenced", func() {
			spec := obs.ClusterLogForwarderSpec{
				Pipelines: []obs.PipelineSpec{
					{InputRefs: []string{string(obs.InputTypeApplication)}},
				},
			}
			result, _ := MigrateInputs(spec)
			Expect(result.Inputs).To(HaveLen(1))
			Expect(result.Inputs[0]).To(Equal(obs.InputSpec{Type: obs.InputTypeApplication, Name: string(obs.InputTypeApplication), Application: &obs.Application{}}))
		})
		It("should stub 'infrastructure' as an input when referenced", func() {
			spec := obs.ClusterLogForwarderSpec{
				Pipelines: []obs.PipelineSpec{
					{InputRefs: []string{string(obs.InputTypeInfrastructure)}},
				},
			}
			result, _ := MigrateInputs(spec)
			Expect(result.Inputs).To(HaveLen(1))
			Expect(result.Inputs[0]).To(Equal(obs.InputSpec{Type: obs.InputTypeInfrastructure, Name: string(obs.InputTypeInfrastructure),
				Infrastructure: &obs.Infrastructure{
					Sources: obs.InfrastructureSources,
				}}))
		})
		It("should stub 'audit' as an input when referenced", func() {
			spec := obs.ClusterLogForwarderSpec{
				Pipelines: []obs.PipelineSpec{
					{InputRefs: []string{string(obs.InputTypeAudit)}},
				},
			}
			result, _ := MigrateInputs(spec)
			Expect(result.Inputs).To(HaveLen(1))
			Expect(result.Inputs[0]).To(Equal(obs.InputSpec{Type: obs.InputTypeAudit, Name: string(obs.InputTypeAudit),
				Audit: &obs.Audit{
					Sources: obs.AuditSources,
				}}))
		})
	})
	Context("when input name is the same as a reserved name", func() {
		It("should replace the input with the 'reserved' representation", func() {
			spec := obs.ClusterLogForwarderSpec{
				Inputs: []obs.InputSpec{
					{
						Name: string(obs.InputTypeApplication),
						Type: obs.InputTypeApplication,
						Application: &obs.Application{
							Includes: []obs.NamespaceContainerSpec{
								{
									Container: "foo",
								},
							},
						},
					},
					{
						Name: "my-bar",
						Type: obs.InputTypeApplication,
						Application: &obs.Application{
							Includes: []obs.NamespaceContainerSpec{
								{
									Container: "foo",
								},
							},
						},
					},
				},
				Pipelines: []obs.PipelineSpec{
					{InputRefs: []string{string(obs.InputTypeApplication)}},
				},
			}
			result, _ := MigrateInputs(spec)
			Expect(result.Inputs).To(HaveLen(2))
			inputs := internalobs.Inputs(result.Inputs).Map()
			Expect(inputs[string(obs.InputTypeApplication)]).To(Equal(obs.InputSpec{Name: string(obs.InputTypeApplication), Application: &obs.Application{}, Type: obs.InputTypeApplication}))
		})
	})

})
