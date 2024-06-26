package observability

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"sort"
)

var _ = Describe("migrateInputs", func() {
	var (
		initContext utils.Options
	)
	BeforeEach(func() {
		initContext = utils.Options{}
	})
	Context("for reserved input names", func() {

		It("should stub 'application' as an input when referenced", func() {
			spec := obs.ClusterLogForwarder{
				Spec: obs.ClusterLogForwarderSpec{
					Pipelines: []obs.PipelineSpec{
						{InputRefs: []string{string(obs.InputTypeApplication)}},
					},
				},
			}
			result, _ := MigrateInputs(spec, initContext)
			Expect(result.Spec.Inputs).To(HaveLen(1))
			Expect(result.Spec.Inputs[0]).To(Equal(obs.InputSpec{Type: obs.InputTypeApplication, Name: string(obs.InputTypeApplication), Application: &obs.Application{}}))
		})
		It("should stub 'infrastructure' as an input when referenced", func() {
			spec := obs.ClusterLogForwarder{
				Spec: obs.ClusterLogForwarderSpec{
					Pipelines: []obs.PipelineSpec{
						{InputRefs: []string{string(obs.InputTypeInfrastructure)}},
					},
				},
			}
			result, _ := MigrateInputs(spec, initContext)
			Expect(result.Spec.Inputs).To(HaveLen(1))
			Expect(result.Spec.Inputs[0]).To(Equal(obs.InputSpec{Type: obs.InputTypeInfrastructure, Name: string(obs.InputTypeInfrastructure),
				Infrastructure: &obs.Infrastructure{
					Sources: obs.InfrastructureSources,
				}}))
		})
		It("should stub 'audit' as an input when referenced", func() {
			spec := obs.ClusterLogForwarder{
				Spec: obs.ClusterLogForwarderSpec{
					Pipelines: []obs.PipelineSpec{
						{InputRefs: []string{string(obs.InputTypeAudit)}},
					},
				},
			}
			result, _ := MigrateInputs(spec, initContext)
			Expect(result.Spec.Inputs).To(HaveLen(1))
			Expect(result.Spec.Inputs[0]).To(Equal(obs.InputSpec{Type: obs.InputTypeAudit, Name: string(obs.InputTypeAudit),
				Audit: &obs.Audit{
					Sources: obs.AuditSources,
				}}))
		})
	})
	Context("when input name is the same as a reserved name", func() {
		It("should replace the input with the 'reserved' representation", func() {
			spec := obs.ClusterLogForwarder{
				Spec: obs.ClusterLogForwarderSpec{
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
				},
			}
			result, _ := MigrateInputs(spec, initContext)
			Expect(result.Spec.Inputs).To(HaveLen(2))
			inputs := internalobs.Inputs(result.Spec.Inputs).Map()
			Expect(inputs[string(obs.InputTypeApplication)]).To(Equal(obs.InputSpec{Name: string(obs.InputTypeApplication), Application: &obs.Application{}, Type: obs.InputTypeApplication}))
		})
	})

	Context("for receiver inputs", func() {
		const forwarderName = "myforwarder"
		var (
			spec obs.ClusterLogForwarder

			migrate = func(spec, exp obs.ClusterLogForwarder) obs.ClusterLogForwarder {
				migratedSpec, _ := MigrateInputs(spec, initContext)
				sort.Slice(migratedSpec.Spec.Inputs, func(i, j int) bool {
					return migratedSpec.Spec.Inputs[i].Name < migratedSpec.Spec.Inputs[j].Name
				})
				sort.Slice(exp.Spec.Inputs, func(i, j int) bool {
					return exp.Spec.Inputs[i].Name < exp.Spec.Inputs[j].Name
				})
				return migratedSpec
			}
		)
		It("should ignore non-receiver inputs", func() {
			spec = obs.ClusterLogForwarder{
				Spec: obs.ClusterLogForwarderSpec{
					Inputs: []obs.InputSpec{
						{Name: "anapp", Type: obs.InputTypeApplication},
						{Name: "anInfra", Type: obs.InputTypeInfrastructure},
						{Name: "anAudit", Type: obs.InputTypeAudit},
					},
				},
			}
			migratedSpec := migrate(spec, spec)
			Expect(migratedSpec.Spec.Inputs).To(Equal(spec.Spec.Inputs))
		})
		It("should ignore migration when receiver specs TLS settings", func() {
			spec = obs.ClusterLogForwarder{
				Spec: obs.ClusterLogForwarderSpec{
					Inputs: []obs.InputSpec{
						{
							Name: "anapp",
							Type: obs.InputTypeReceiver,
							Receiver: &obs.ReceiverSpec{
								TLS: &obs.InputTLSSpec{},
							},
						},
					},
				},
			}

			migratedSpec := migrate(spec, spec)
			Expect(migratedSpec.Spec.Inputs).To(Equal(spec.Spec.Inputs))
		})
		It("should add TLS settings that match the cert signing service when TLS is not spec'd", func() {
			spec = obs.ClusterLogForwarder{
				Spec: obs.ClusterLogForwarderSpec{
					Inputs: []obs.InputSpec{
						{
							Name:     "anapp",
							Type:     obs.InputTypeReceiver,
							Receiver: &obs.ReceiverSpec{},
						},
					},
				},
			}
			spec.Name = forwarderName
			secretName := fmt.Sprintf("%s-%s", forwarderName, "anapp")

			exp := obs.ClusterLogForwarderSpec{
				Inputs: []obs.InputSpec{
					{
						Name: "anapp",
						Type: obs.InputTypeReceiver,
						Receiver: &obs.ReceiverSpec{
							TLS: &obs.InputTLSSpec{
								Key: &obs.SecretKey{
									Key: constants.ClientPrivateKey,
									Secret: &corev1.LocalObjectReference{
										Name: secretName,
									},
								},
								Certificate: &obs.ConfigMapOrSecretKey{
									Key: constants.ClientCertKey,
									Secret: &corev1.LocalObjectReference{
										Name: secretName,
									},
								},
							},
						},
					},
				},
			}

			migratedSpec := migrate(spec, spec)
			Expect(migratedSpec.Spec.Inputs).To(Equal(exp.Inputs))
			secrets, found := utils.GetOption[[]*corev1.Secret](initContext, GeneratedSecrets, []*corev1.Secret{})
			Expect(found).To(BeTrue(), fmt.Sprintf("Exp. TLS secrets to add to deployment to be identified in the context: %v", initContext))
			Expect(secrets).To(BeComparableTo([]*corev1.Secret{
				runtime.NewSecret("", secretName,
					map[string][]byte{
						constants.ClientPrivateKey: {},
						constants.ClientCertKey:    {},
					}),
			}), "Exp. context to include the secrets to mount to the deployment")
		})
	})

})
