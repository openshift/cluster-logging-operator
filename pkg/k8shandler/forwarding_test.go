package k8shandler

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	helpers "github.com/openshift/cluster-logging-operator/test"
)

const (
	namespace             = "aNamespace"
	otherTargetName       = "someothername"
	theInternalOutputName = "clo-default-output-es"
)

func HasPipelineStatus(status *logging.ForwardingStatus, pipelineName string, state logging.PipelineState, reason logging.PipelineStateReason) bool {
	logger.Debugf("Pipeline Status: %v", status.Pipelines)
	for _, pipeline := range status.Pipelines {
		if pipeline.Name == pipelineName && pipeline.State == state {
			for _, statusReason := range pipeline.Reasons {
				if reason == statusReason {
					return true
				} else {
					logger.Debugf("Unable to match reason (%s) in validation ", statusReason)
				}
			}
		} else {
			logger.Debugf("Unable to match name (%s) or state (%s) in validation ", pipelineName, state)
		}
	}
	return false
}
func HasOutputStatus(status *logging.ForwardingStatus, outputName string, state logging.OutputState, reason logging.OutputStateReason) bool {
	logger.Debugf("Output Status: %v", status.Outputs)
	for _, output := range status.Outputs {
		if output.Name == outputName && output.State == state {
			if string(reason) == "" {
				return true
			}
			for _, statusReason := range output.Reasons {
				if reason == statusReason {
					return true
				}
			}
		}
	}
	return false
}

var _ = Describe("Normalizing Forwarding", func() {

	var (
		cluster                  *logging.ClusterLogging
		normalizedForwardingSpec logging.ForwardingSpec
		output                   logging.OutputSpec
		otherOutput              logging.OutputSpec
		request                  *ClusterLoggingRequest
	)
	BeforeEach(func() {
		output = logging.OutputSpec{
			Name:     "myOutput",
			Type:     logging.OutputTypeElasticsearch,
			Endpoint: "anOutPut",
		}
		otherOutput = logging.OutputSpec{
			Name:     otherTargetName,
			Type:     logging.OutputTypeElasticsearch,
			Endpoint: "someotherendpoint",
		}
		request = &ClusterLoggingRequest{
			client:  fake.NewFakeClient(),
			cluster: &logging.ClusterLogging{},
		}
		cluster = request.cluster
	})

	Context("while validating ", func() {

		BeforeEach(func() {
			cluster.Spec.Forwarding = &logging.ForwardingSpec{
				Outputs: []logging.OutputSpec{
					output,
					otherOutput,
				},
				Pipelines: []logging.PipelineSpec{
					logging.PipelineSpec{
						Name:       "aPipeline",
						OutputRefs: []string{output.Name, otherOutput.Name},
						SourceType: logging.LogSourceTypeApp,
					},
				},
			}
		})
		Context("pipelines", func() {

			It("should drop pipelines that do not have unique names", func() {
				cluster.Spec.Forwarding.Pipelines = append(cluster.Spec.Forwarding.Pipelines,
					logging.PipelineSpec{
						Name:       "aPipeline",
						OutputRefs: []string{output.Name, otherOutput.Name},
						SourceType: logging.LogSourceTypeApp,
					})
				//sanity check
				Expect(len(cluster.Spec.Forwarding.Pipelines)).To(Equal(2))
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(1), "Exp. non-unique pipelines to be dropped")
				Expect(HasPipelineStatus(cluster.Status.Forwarding, "pipeline[1]", logging.PipelineStateDropped, logging.PipelineStateReasonNonUniqueName)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should drop pipelines that have empty names", func() {
				cluster.Spec.Forwarding.Pipelines = append(cluster.Spec.Forwarding.Pipelines,
					logging.PipelineSpec{
						Name:       "",
						OutputRefs: []string{output.Name, otherOutput.Name},
						SourceType: logging.LogSourceTypeApp,
					})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				logger.Debug(normalizedForwardingSpec.Pipelines)
				Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(1), "Exp. empty pipelines to be dropped")
				Expect(HasPipelineStatus(cluster.Status.Forwarding, "pipeline[1]", logging.PipelineStateDropped, logging.PipelineStateReasonMissingName)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should drop pipelines that conflict with the internally reserved name", func() {
				cluster.Spec.Forwarding.Pipelines = append(cluster.Spec.Forwarding.Pipelines,
					logging.PipelineSpec{
						Name:       defaultAppPipelineName,
						OutputRefs: []string{output.Name, otherOutput.Name},
						SourceType: logging.LogSourceTypeApp,
					})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(1), "Exp. pipelines with an internal name conflict to be dropped")
				Expect(HasPipelineStatus(cluster.Status.Forwarding, "pipeline[1]", logging.PipelineStateDropped, logging.PipelineStateReasonReservedNameConflict)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should drop pipelines that have missing sources", func() {
				cluster.Spec.Forwarding.Pipelines = append(cluster.Spec.Forwarding.Pipelines,
					logging.PipelineSpec{
						Name:       "someDefinedPipeline",
						OutputRefs: []string{output.Name, otherOutput.Name},
						SourceType: "",
					})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(1), "Exp. pipelines with an empty source to be dropped")
				Expect(HasPipelineStatus(cluster.Status.Forwarding, "someDefinedPipeline", logging.PipelineStateDropped, logging.PipelineStateReasonMissingSource)).To(BeTrue(), "Exp. the status to be updated")
			})
			It("should drop pipelines that have unrecognized sources", func() {
				cluster.Spec.Forwarding.Pipelines = append(cluster.Spec.Forwarding.Pipelines,
					logging.PipelineSpec{
						Name:       "someDefinedPipeline",
						OutputRefs: []string{output.Name, otherOutput.Name},
						SourceType: "foo",
					})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(1), "Exp. outputs with an unrecognized type to be dropped")
				Expect(HasPipelineStatus(cluster.Status.Forwarding, "someDefinedPipeline", logging.PipelineStateDropped, logging.PipelineStateReasonUnrecognizedSource)).To(BeTrue(), "Exp. the status to be updated")
			})
			It("should drop pipelines that have no outputRefs", func() {
				cluster.Spec.Forwarding.Pipelines = append(cluster.Spec.Forwarding.Pipelines,
					logging.PipelineSpec{
						Name:       "someDefinedPipeline",
						OutputRefs: []string{},
						SourceType: logging.LogSourceTypeApp,
					})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(1), "Exp. outputs with an unrecognized type to be dropped")
				Expect(HasPipelineStatus(cluster.Status.Forwarding, "someDefinedPipeline", logging.PipelineStateDropped, logging.PipelineStateReasonMissingOutputs)).To(BeTrue(), "Exp. the status to be updated")
			})
			It("should degrade pipelines where there are fewer outputs then defined outputRefs", func() {
				cluster.Spec.Forwarding.Pipelines = append(cluster.Spec.Forwarding.Pipelines,
					logging.PipelineSpec{
						Name:       "someDefinedPipeline",
						OutputRefs: []string{output.Name, otherOutput.Name, "aMissingOutput"},
						SourceType: logging.LogSourceTypeApp,
					})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(2), "Exp. all defined pipelines")
				Expect(HasPipelineStatus(cluster.Status.Forwarding, "someDefinedPipeline", logging.PipelineStateDegraded, logging.PipelineStateReasonMissingOutputs)).To(BeTrue(), "Exp. the status to be updated")
			})

		})

		Context("outputs", func() {

			It("should drop outputs that do not have unique names", func() {
				cluster.Spec.Forwarding.Outputs = append(cluster.Spec.Forwarding.Outputs, logging.OutputSpec{
					Name:     "myOutput",
					Type:     logging.OutputTypeElasticsearch,
					Endpoint: "anOutPut",
				})
				//sanity check
				Expect(len(cluster.Spec.Forwarding.Outputs)).To(Equal(3))
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. non-unique outputs to be dropped")
				Expect(HasOutputStatus(cluster.Status.Forwarding, "output[2]", logging.OutputStateDropped, logging.OutputStateNonUniqueName)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should drop outputs that have empty names", func() {
				cluster.Spec.Forwarding.Outputs = append(cluster.Spec.Forwarding.Outputs, logging.OutputSpec{
					Name:     "",
					Type:     logging.OutputTypeElasticsearch,
					Endpoint: "anOutPut",
				})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with an empty name to be dropped")
				Expect(HasOutputStatus(cluster.Status.Forwarding, "output[2]", logging.OutputStateDropped, logging.OutputStateReasonMissingName)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should drop outputs that conflict with the internally reserved name", func() {
				cluster.Spec.Forwarding.Outputs = append(cluster.Spec.Forwarding.Outputs, logging.OutputSpec{
					Name:     internalOutputName,
					Type:     logging.OutputTypeElasticsearch,
					Endpoint: "anOutPut",
				})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with an internal name conflict to be dropped")
				Expect(HasOutputStatus(cluster.Status.Forwarding, "output[2]", logging.OutputStateDropped, logging.OutputStateReservedNameConflict)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should drop outputs that have empty types", func() {
				cluster.Spec.Forwarding.Outputs = append(cluster.Spec.Forwarding.Outputs, logging.OutputSpec{
					Name:     "aName",
					Type:     "",
					Endpoint: "anOutPut",
				})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with an empty type to be dropped")
				Expect(HasOutputStatus(cluster.Status.Forwarding, "aName", logging.OutputStateDropped, logging.OutputStateReasonMissingType)).To(BeTrue(), "Exp. the status to be updated")
			})
			It("should drop outputs that have unrecognized types", func() {
				cluster.Spec.Forwarding.Outputs = append(cluster.Spec.Forwarding.Outputs, logging.OutputSpec{
					Name:     "aName",
					Type:     "foo",
					Endpoint: "anOutPut",
				})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with an unrecognized type to be dropped")
				Expect(HasOutputStatus(cluster.Status.Forwarding, "aName", logging.OutputStateDropped, logging.OutputStateReasonUnrecognizedType)).To(BeTrue(), "Exp. the status to be updated")
			})
			It("should drop outputs that have empty endpoints", func() {
				cluster.Spec.Forwarding.Outputs = append(cluster.Spec.Forwarding.Outputs, logging.OutputSpec{
					Name:     "aName",
					Type:     "foo",
					Endpoint: "",
				})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with an empty endpoint to be dropped")
				Expect(HasOutputStatus(cluster.Status.Forwarding, "aName", logging.OutputStateDropped, logging.OutputStateReasonMissingEndpoint)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should drop outputs that have secrets with no names", func() {
				cluster.Spec.Forwarding.Outputs = append(cluster.Spec.Forwarding.Outputs, logging.OutputSpec{
					Name:     "aName",
					Type:     logging.OutputTypeElasticsearch,
					Endpoint: "an output",
					Secret:   &logging.OutputSecretSpec{},
				})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with empty secrets to be dropped")
				Expect(HasOutputStatus(cluster.Status.Forwarding, "aName", logging.OutputStateDropped, logging.OutputStateReasonMissingSecretName)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should drop outputs that have secrets which don't exist", func() {
				cluster.Spec.Forwarding.Outputs = append(cluster.Spec.Forwarding.Outputs, logging.OutputSpec{
					Name:     "aName",
					Type:     logging.OutputTypeElasticsearch,
					Endpoint: "an output",
					Secret: &logging.OutputSecretSpec{
						Name: "mysecret",
					},
				})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with non-existent secrets to be dropped")
				Expect(HasOutputStatus(cluster.Status.Forwarding, "aName", logging.OutputStateDropped, logging.OutputStateReasonSecretDoesNotExist)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should accept well formed outputs", func() {
				request = &ClusterLoggingRequest{
					client: fake.NewFakeClient(&core.Secret{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Secret",
							APIVersion: core.SchemeGroupVersion.String(),
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "mysecret",
							Namespace: "openshift-logging",
						},
					}),
					cluster: cluster,
				}
				cluster.Spec.Forwarding.Outputs = append(cluster.Spec.Forwarding.Outputs, logging.OutputSpec{
					Name:     "aName",
					Type:     logging.OutputTypeElasticsearch,
					Endpoint: "an output",
					Secret: &logging.OutputSecretSpec{
						Name: "mysecret",
					},
				})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(3))
				Expect(HasOutputStatus(cluster.Status.Forwarding, "aName", logging.OutputStateAccepted, logging.OutputStateReason(""))).To(BeTrue(), "Exp. the status to be updated")
			})

		})
	})

	Context("and a logstore is not spec'd", func() {
		It("should return an empty forwarding spec", func() {
			spec := request.normalizeLogForwarding(namespace, cluster)
			Expect(spec).To(Equal(logging.ForwardingSpec{}))
		})
	})
	Context("and a logstore is spec'd", func() {
		BeforeEach(func() {
			cluster.Spec.LogStore = &logging.LogStoreSpec{
				Type: logging.LogStoreTypeElasticsearch,
			}
		})
		Context("and forwarding not spec'd", func() {
			It("should forward app and infra logs to the logstore", func() {
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(1), "Exp. internal outputs to be included in the normalized forwarding")
				Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(2), "Exp pipelines for application and infra logs")
				sources := []string{}
				for _, pipeline := range normalizedForwardingSpec.Pipelines {
					Expect(pipeline.OutputRefs).To(Equal([]string{theInternalOutputName}))
					sources = append(sources, string(pipeline.SourceType))
				}
				Expect(len(sources)).To(Equal(2), fmt.Sprintf("Sources: %v", sources))
				Expect(helpers.StringsContain(sources, string(logging.LogSourceTypeApp))).To(BeTrue(), "Exp. the internal pipeline to include app logs")
				Expect(helpers.StringsContain(sources, string(logging.LogSourceTypeInfra))).To(BeTrue(), "Exp. the internal pipeline to include infa logs")
			})
		})
		Context("and forwarding is defined", func() {

			Context("and disableDefaultForwarding is false with no other defined pipelines", func() {
				It("should forward app and infra logs to the logstore", func() {
					normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
					Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(1), "Exp. internal outputs to be included in the normalized forwarding")
					Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(2), "Exp pipelines for application and infra logs")
					sources := []string{}
					for _, pipeline := range normalizedForwardingSpec.Pipelines {
						Expect(pipeline.OutputRefs).To(Equal([]string{theInternalOutputName}))
						sources = append(sources, string(pipeline.SourceType))
					}
					Expect(len(sources)).To(Equal(2), fmt.Sprintf("Sources: %v", sources))
					Expect(helpers.StringsContain(sources, string(logging.LogSourceTypeApp))).To(BeTrue(), "Exp. the internal pipeline to include app logs")
					Expect(helpers.StringsContain(sources, string(logging.LogSourceTypeInfra))).To(BeTrue(), "Exp. the internal pipeline to include infa logs")
				})
			})
			Context("and disableDefaultForwarding is true", func() {

				Context("and a pipline spec'd for undefined outputs", func() {
					BeforeEach(func() {
						cluster.Spec.Forwarding = &logging.ForwardingSpec{
							Pipelines: []logging.PipelineSpec{
								logging.PipelineSpec{
									Name:       "mypipeline",
									OutputRefs: []string{output.Name},
									SourceType: logging.LogSourceTypeApp,
								},
							},
						}
						normalizedForwardingSpec = request.normalizeLogForwarding(namespace, cluster)
					})
					It("should drop the pipeline", func() {
						Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(0))
					})
				})
				Context("and a pipline spec'd with some undefined outputs", func() {
					BeforeEach(func() {
						cluster.Spec.Forwarding = &logging.ForwardingSpec{
							Outputs: []logging.OutputSpec{output},
							Pipelines: []logging.PipelineSpec{
								logging.PipelineSpec{
									Name:       "mypipeline",
									OutputRefs: []string{otherOutput.Name, output.Name},
									SourceType: logging.LogSourceTypeApp,
								},
							},
						}
						normalizedForwardingSpec = request.normalizeLogForwarding(namespace, cluster)
					})
					It("should drop the undefined outputs", func() {
						Expect(normalizedForwardingSpec.Outputs).To(Equal(cluster.Spec.Forwarding.Outputs), "Exp. outputs to be included in the normalized forwarding")
						Expect(normalizedForwardingSpec.Pipelines[0].OutputRefs).To(Equal([]string{output.Name}))
						Expect(helpers.StringsContain(normalizedForwardingSpec.Pipelines[0].OutputRefs, theInternalOutputName)).To(Not(BeTrue()), "Exp. the internal log store to not be part of the pipelines")
					})
				})
				Context("and a pipline spec'd for defined outputs", func() {
					BeforeEach(func() {
						cluster.Spec.Forwarding = &logging.ForwardingSpec{
							Outputs: []logging.OutputSpec{
								output,
								otherOutput,
							},
							Pipelines: []logging.PipelineSpec{
								logging.PipelineSpec{
									Name:       "mypipeline",
									OutputRefs: []string{output.Name, otherOutput.Name},
									SourceType: logging.LogSourceTypeApp,
								},
							},
						}
						normalizedForwardingSpec = request.normalizeLogForwarding(namespace, cluster)
					})
					It("should forward the pipeline's source to all the spec'd output", func() {
						Expect(normalizedForwardingSpec.Outputs).To(Equal(cluster.Spec.Forwarding.Outputs), "Exp. outputs to be included in the normalized forwarding")
						Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(1), "Expected the pipeline to forward to its output")
						pipeline := normalizedForwardingSpec.Pipelines[0]
						Expect(pipeline.OutputRefs).To(Equal([]string{output.Name, otherOutput.Name}))
						Expect(helpers.StringsContain(pipeline.OutputRefs, theInternalOutputName)).To(Not(BeTrue()), "Exp. the internal log store to not be part of the pipelines")
					})
				})
			})
		})
	})
})
