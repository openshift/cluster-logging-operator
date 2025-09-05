package initialize

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("migrateOutputs", func() {
	var (
		initContext utils.Options
	)
	BeforeEach(func() {
		initContext = utils.Options{}
	})
	Context("for LokiStack outputs", func() {

		Context("when the OTLP Tech-preview feature is not enabled", func() {
			It("should set the datamodel to 'Viaq' regardless of of the value spec'd for the datamodel", func() {
				forwarder := obs.ClusterLogForwarder{
					Spec: obs.ClusterLogForwarderSpec{
						Outputs: []obs.OutputSpec{
							{
								Name: "anOutput",
								Type: obs.OutputTypeLokiStack,
								LokiStack: &obs.LokiStack{
									DataModel: obs.LokiStackDataModelOpenTelemetry,
								}},
						},
					},
				}
				result := MigrateOutputs(forwarder, initContext)
				Expect(result.Spec.Outputs).To(HaveLen(1))
				Expect(result.Spec.Outputs[0]).To(Equal(obs.OutputSpec{
					Name: "anOutput",
					Type: obs.OutputTypeLokiStack,
					LokiStack: &obs.LokiStack{
						DataModel: obs.LokiStackDataModelViaq,
					},
				}))
			})

		})

		Context("when the OTLP Tech-preview feature is enabled", func() {

			It("should set the datamodel to 'Otel' when the datamodel is not spec'd", func() {
				forwarder := obs.ClusterLogForwarder{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							constants.AnnotationOtlpOutputTechPreview: "enabled",
						},
					},
					Spec: obs.ClusterLogForwarderSpec{
						Outputs: []obs.OutputSpec{
							{
								Name:      "anOutput",
								Type:      obs.OutputTypeLokiStack,
								LokiStack: &obs.LokiStack{},
							},
						},
					},
				}
				result := MigrateOutputs(forwarder, initContext)
				Expect(result.Spec.Outputs).To(HaveLen(1))
				Expect(result.Spec.Outputs[0]).To(Equal(obs.OutputSpec{
					Name: "anOutput",
					Type: obs.OutputTypeLokiStack,
					LokiStack: &obs.LokiStack{
						DataModel: obs.LokiStackDataModelOpenTelemetry,
					},
				}))
			})
			It("should honor the datamodel 'ViaQ' when the datamodel is spec'd to ViaQ", func() {
				forwarder := obs.ClusterLogForwarder{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							constants.AnnotationOtlpOutputTechPreview: "true",
						},
					},
					Spec: obs.ClusterLogForwarderSpec{
						Outputs: []obs.OutputSpec{
							{
								Name: "anOutput",
								Type: obs.OutputTypeLokiStack,
								LokiStack: &obs.LokiStack{
									DataModel: obs.LokiStackDataModelViaq,
								},
							},
						},
					},
				}
				result := MigrateOutputs(forwarder, initContext)
				Expect(result.Spec.Outputs).To(HaveLen(1))
				Expect(result.Spec.Outputs[0]).To(Equal(obs.OutputSpec{
					Name: "anOutput",
					Type: obs.OutputTypeLokiStack,
					LokiStack: &obs.LokiStack{
						DataModel: obs.LokiStackDataModelViaq,
					},
				}))
			})
			It("should honor the datamodel 'Otel' when the datamodel is spec'd to Otel", func() {
				forwarder := obs.ClusterLogForwarder{
					ObjectMeta: v1.ObjectMeta{
						Annotations: map[string]string{
							constants.AnnotationOtlpOutputTechPreview: "true",
						},
					},
					Spec: obs.ClusterLogForwarderSpec{
						Outputs: []obs.OutputSpec{
							{
								Name: "anOutput",
								Type: obs.OutputTypeLokiStack,
								LokiStack: &obs.LokiStack{
									DataModel: obs.LokiStackDataModelOpenTelemetry,
								},
							},
						},
					},
				}
				result := MigrateOutputs(forwarder, initContext)
				Expect(result.Spec.Outputs).To(HaveLen(1))
				Expect(result.Spec.Outputs[0]).To(Equal(obs.OutputSpec{
					Name: "anOutput",
					Type: obs.OutputTypeLokiStack,
					LokiStack: &obs.LokiStack{
						DataModel: obs.LokiStackDataModelOpenTelemetry,
					},
				}))
			})
		})

	})
})
