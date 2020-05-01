package k8shandler

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	cl "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
)

const (
	namespace             = "aNamespace"
	otherTargetName       = "someothername"
	theInternalOutputName = "clo-default-output-es"
)

func HasPipelineStatus(status *logging.ForwardingStatus, pipelineName string, state logging.PipelineState, reason logging.PipelineConditionReason) bool {
	logger.Debugf("Pipeline Status: %v", status.Pipelines)
	for _, pipeline := range status.Pipelines {
		if pipeline.Name == pipelineName && pipeline.State == state {
			for _, condition := range pipeline.Conditions {
				if reason == condition.Reason {
					return true
				} else {
					logger.Debugf("Unable to match reason (%s) in validation ", condition)
				}
			}
		} else {
			logger.Debugf("Unable to match name (%s) or state (%s) in validation ", pipelineName, state)
		}
	}
	return false
}
func HasOutputStatus(status *logging.ForwardingStatus, outputName string, state logging.OutputState, reason logging.OutputConditionReason, skipReason bool) bool {
	logger.Debugf("Output Status: %v", status.Outputs)
	for _, output := range status.Outputs {
		if output.Name == outputName && output.State == state {
			if skipReason {
				return true
			}
			for _, condition := range output.Conditions {
				if reason == condition.Reason {
					return true
				}
			}
		}
	}
	return false
}

var _ = Describe("Normalizing Forwarding", func() {

	var (
		cluster                  *cl.ClusterLogging
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
			client: fake.NewFakeClient(),
			cluster: &cl.ClusterLogging{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ForwardingAnnotation: "enabled",
					},
				},
			},
			ForwardingRequest: &logging.LogForwarding{},
		}
		cluster = request.cluster
	})

	It("should have sourceType application", func() {
		Expect(sourceTypes.Has(string(logging.LogSourceTypeApp))).To(BeTrue())
	})
	It("should have sourceType infra", func() {
		Expect(sourceTypes.Has(string(logging.LogSourceTypeInfra))).To(BeTrue())
	})
	It("should have outputType Elastic", func() {
		Expect(outputTypes.Has(string(logging.OutputTypeElasticsearch))).To(BeTrue())
	})
	It("should have outputType Forward", func() {
		Expect(outputTypes.Has(string(logging.OutputTypeForward))).To(BeTrue())
	})

	Context("while validating ", func() {

		BeforeEach(func() {
			request.ForwardingSpec = logging.ForwardingSpec{
				Outputs: []logging.OutputSpec{
					output,
					otherOutput,
				},
				Pipelines: []logging.PipelineSpec{
					{
						Name:       "aPipeline",
						OutputRefs: []string{output.Name, otherOutput.Name},
						SourceType: logging.LogSourceTypeApp,
					},
				},
			}
		})
		Context("pipelines", func() {

			It("should only include logsources if there is atleast one valid pipeline", func() {
				request.ForwardingSpec.Pipelines = []logging.PipelineSpec{
					{
						Name:       "aPipeline",
						OutputRefs: []string{"someotherendpoint"},
						SourceType: logging.LogSourceTypeApp,
					},
				}
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(0), "Exp. all pipelines to be dropped")
				Expect(request.ForwardingRequest.Status.LogSources).To(Equal([]logging.LogSourceType{}), "Exp. no log sources")
			})

			It("should drop pipelines that do not have unique names", func() {
				request.ForwardingSpec.Pipelines = append(request.ForwardingSpec.Pipelines,
					logging.PipelineSpec{
						Name:       "aPipeline",
						OutputRefs: []string{output.Name, otherOutput.Name},
						SourceType: logging.LogSourceTypeApp,
					})
				//sanity check
				Expect(len(request.ForwardingSpec.Pipelines)).To(Equal(2))
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(1), "Exp. non-unique pipelines to be dropped")
				Expect(HasPipelineStatus(request.ForwardingRequest.Status, "pipeline[1]", logging.PipelineStateDropped, logging.PipelineConditionReasonUniqueName)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should drop pipelines that have empty names", func() {
				request.ForwardingSpec.Pipelines = append(request.ForwardingSpec.Pipelines,
					logging.PipelineSpec{
						Name:       "",
						OutputRefs: []string{output.Name, otherOutput.Name},
						SourceType: logging.LogSourceTypeApp,
					})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				logger.Debug(normalizedForwardingSpec.Pipelines)
				Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(1), "Exp. empty pipelines to be dropped")
				Expect(HasPipelineStatus(request.ForwardingRequest.Status, "pipeline[1]", logging.PipelineStateDropped, logging.PipelineConditionReasonMissingName)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should drop pipelines that conflict with the internally reserved name", func() {
				request.ForwardingSpec.Pipelines = append(request.ForwardingSpec.Pipelines,
					logging.PipelineSpec{
						Name:       defaultAppPipelineName,
						OutputRefs: []string{output.Name, otherOutput.Name},
						SourceType: logging.LogSourceTypeApp,
					})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(1), "Exp. pipelines with an internal name conflict to be dropped")
				Expect(HasPipelineStatus(request.ForwardingRequest.Status, "pipeline[1]", logging.PipelineStateDropped, logging.PipelineConditionReasonReservedNameConflict)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should drop pipelines that have missing sources", func() {
				request.ForwardingSpec.Pipelines = append(request.ForwardingSpec.Pipelines,
					logging.PipelineSpec{
						Name:       "someDefinedPipeline",
						OutputRefs: []string{output.Name, otherOutput.Name},
						SourceType: "",
					})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(1), "Exp. pipelines with an empty source to be dropped")
				Expect(HasPipelineStatus(request.ForwardingRequest.Status, "someDefinedPipeline", logging.PipelineStateDropped, logging.PipelineConditionReasonMissingSource)).To(BeTrue(), "Exp. the status to be updated")
			})
			It("should drop pipelines that have unrecognized sources", func() {
				request.ForwardingSpec.Pipelines = append(request.ForwardingSpec.Pipelines,
					logging.PipelineSpec{
						Name:       "someDefinedPipeline",
						OutputRefs: []string{output.Name, otherOutput.Name},
						SourceType: "foo",
					})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(1), "Exp. outputs with an unrecognized type to be dropped")
				Expect(HasPipelineStatus(request.ForwardingRequest.Status, "someDefinedPipeline", logging.PipelineStateDropped, logging.PipelineConditionReasonUnrecognizedSourceType)).To(BeTrue(), "Exp. the status to be updated")
			})
			It("should drop pipelines that have no outputRefs", func() {
				request.ForwardingSpec.Pipelines = append(request.ForwardingSpec.Pipelines,
					logging.PipelineSpec{
						Name:       "someDefinedPipeline",
						OutputRefs: []string{},
						SourceType: logging.LogSourceTypeApp,
					})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(1), "Exp. outputs with an unrecognized type to be dropped")
				Expect(HasPipelineStatus(request.ForwardingRequest.Status, "someDefinedPipeline", logging.PipelineStateDropped, logging.PipelineConditionReasonMissingOutputs)).To(BeTrue(), "Exp. the status to be updated")
			})
			It("should degrade pipelines where there are fewer outputs then defined outputRefs", func() {
				request.ForwardingSpec.Pipelines = append(request.ForwardingSpec.Pipelines,
					logging.PipelineSpec{
						Name:       "someDefinedPipeline",
						OutputRefs: []string{output.Name, otherOutput.Name, "aMissingOutput"},
						SourceType: logging.LogSourceTypeApp,
					})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(2), "Exp. all defined pipelines")
				Expect(HasPipelineStatus(request.ForwardingRequest.Status, "someDefinedPipeline", logging.PipelineStateDegraded, logging.PipelineConditionReasonMissingOutputs)).To(BeTrue(), "Exp. the status to be updated")
			})

		})

		Context("outputs", func() {

			It("should drop outputs that do not have unique names", func() {
				request.ForwardingSpec.Outputs = append(request.ForwardingSpec.Outputs, logging.OutputSpec{
					Name:     "myOutput",
					Type:     logging.OutputTypeElasticsearch,
					Endpoint: "anOutPut",
				})
				//sanity check
				Expect(len(request.ForwardingSpec.Outputs)).To(Equal(3))
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. non-unique outputs to be dropped")
				Expect(HasOutputStatus(request.ForwardingRequest.Status, "output[2]", logging.OutputStateDropped, logging.OutputConditionReasonNonUniqueName, false)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should drop outputs that have empty names", func() {
				request.ForwardingSpec.Outputs = append(request.ForwardingSpec.Outputs, logging.OutputSpec{
					Name:     "",
					Type:     logging.OutputTypeElasticsearch,
					Endpoint: "anOutPut",
				})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with an empty name to be dropped")
				Expect(HasOutputStatus(request.ForwardingRequest.Status, "output[2]", logging.OutputStateDropped, logging.OutputConditionReasonMissingName, false)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should drop outputs that conflict with the internally reserved name", func() {
				request.ForwardingSpec.Outputs = append(request.ForwardingSpec.Outputs, logging.OutputSpec{
					Name:     internalOutputName,
					Type:     logging.OutputTypeElasticsearch,
					Endpoint: "anOutPut",
				})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with an internal name conflict to be dropped")
				Expect(HasOutputStatus(request.ForwardingRequest.Status, "output[2]", logging.OutputStateDropped, logging.OutputConditionReasonReservedNameConflict, false)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should drop outputs that have empty types", func() {
				request.ForwardingSpec.Outputs = append(request.ForwardingSpec.Outputs, logging.OutputSpec{
					Name:     "aName",
					Type:     "",
					Endpoint: "anOutPut",
				})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with an empty type to be dropped")
				Expect(HasOutputStatus(request.ForwardingRequest.Status, "aName", logging.OutputStateDropped, logging.OutputConditionReasonMissingType, false)).To(BeTrue(), "Exp. the status to be updated")
			})
			It("should drop outputs that have unrecognized types", func() {
				request.ForwardingSpec.Outputs = append(request.ForwardingSpec.Outputs, logging.OutputSpec{
					Name:     "aName",
					Type:     "foo",
					Endpoint: "anOutPut",
				})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with an unrecognized type to be dropped")
				Expect(HasOutputStatus(request.ForwardingRequest.Status, "aName", logging.OutputStateDropped, logging.OutputConditionReasonUnrecognizedType, false)).To(BeTrue(), "Exp. the status to be updated")
			})
			It("should drop outputs that have empty endpoints", func() {
				request.ForwardingSpec.Outputs = append(request.ForwardingSpec.Outputs, logging.OutputSpec{
					Name:     "aName",
					Type:     "fluentForward",
					Endpoint: "",
				})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with an empty endpoint to be dropped")
				Expect(HasOutputStatus(request.ForwardingRequest.Status, "aName", logging.OutputStateDropped, logging.OutputConditionReasonMissingEndpoint, false)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should drop outputs that have secrets with no names", func() {
				request.ForwardingSpec.Outputs = append(request.ForwardingSpec.Outputs, logging.OutputSpec{
					Name:     "aName",
					Type:     logging.OutputTypeElasticsearch,
					Endpoint: "an output",
					Secret:   &logging.OutputSecretSpec{},
				})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with empty secrets to be dropped")
				Expect(HasOutputStatus(request.ForwardingRequest.Status, "aName", logging.OutputStateDropped, logging.OutputConditionReasonMissingSecretName, false)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should drop outputs that have secrets which don't exist", func() {
				request.ForwardingSpec.Outputs = append(request.ForwardingSpec.Outputs, logging.OutputSpec{
					Name:     "aName",
					Type:     logging.OutputTypeElasticsearch,
					Endpoint: "an output",
					Secret: &logging.OutputSecretSpec{
						Name: "mysecret",
					},
				})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with non-existent secrets to be dropped")
				Expect(HasOutputStatus(request.ForwardingRequest.Status, "aName", logging.OutputStateDropped, logging.OutputConditionReasonSecretDoesNotExist, false)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should drop forward outputs that have secrets and is missing shared_key", func() {
				secret := &core.Secret{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: core.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mysecret",
						Namespace: "openshift-logging",
					},
					Type: "Opaque",
					Data: map[string][]byte{
						"foo": []byte("bar"),
					},
				}
				request.client = fake.NewFakeClient(secret)
				request.ForwardingSpec.Outputs = append(request.ForwardingSpec.Outputs, logging.OutputSpec{
					Name:     "aName",
					Type:     logging.OutputTypeForward,
					Endpoint: "an output",
					Secret: &logging.OutputSecretSpec{
						Name: "mysecret",
					},
				})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with missing shared_key to be dropped")
				Expect(HasOutputStatus(request.ForwardingRequest.Status, "aName", logging.OutputStateDropped, logging.OutputConditionReasonSecretMissingSharedKey, false)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should accept well formed outputs", func() {
				request.client = fake.NewFakeClient(&core.Secret{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: core.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mysecret",
						Namespace: "openshift-logging",
					},
				})
				request.ForwardingSpec.Outputs = append(request.ForwardingSpec.Outputs, logging.OutputSpec{
					Name:     "aName",
					Type:     logging.OutputTypeElasticsearch,
					Endpoint: "an output",
					Secret: &logging.OutputSecretSpec{
						Name: "mysecret",
					},
				})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(3))
				Expect(HasOutputStatus(request.ForwardingRequest.Status, "aName", logging.OutputStateAccepted, logging.OutputConditionReasonSecretDoesNotExist, true)).To(BeTrue(), "Exp. the status to be updated")
			})

		})
	})

	Context("and a logstore is not spec'd", func() {
		It("should return an empty forwarding spec", func() {
			spec := request.normalizeLogForwarding(namespace, cluster)
			Expect(spec).To(Equal(logging.ForwardingSpec{Outputs: []logging.OutputSpec{}, Pipelines: nil}))
		})
	})
	Context("and a logstore is spec'd", func() {
		BeforeEach(func() {
			cluster.Spec.LogStore = &cl.LogStoreSpec{
				Type: cl.LogStoreTypeElasticsearch,
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
				Expect(sources).To(ContainElement(string(logging.LogSourceTypeApp)), "Exp. the internal pipeline to include app logs")
				Expect(sources).To(ContainElement(string(logging.LogSourceTypeInfra)), "Exp. the internal pipeline to include infa logs")
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
					Expect(sources).To(ContainElement(string(logging.LogSourceTypeApp)), "Exp. the internal pipeline to include app logs")
					Expect(sources).To(ContainElement(string(logging.LogSourceTypeInfra)), "Exp. the internal pipeline to include infa logs")
				})
			})
			Context("and disableDefaultForwarding is true", func() {

				Context("and a pipline spec'd for undefined outputs", func() {
					BeforeEach(func() {
						request.ForwardingSpec = logging.ForwardingSpec{
							Pipelines: []logging.PipelineSpec{
								{
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
						request.ForwardingSpec = logging.ForwardingSpec{
							Outputs: []logging.OutputSpec{output},
							Pipelines: []logging.PipelineSpec{
								{
									Name:       "mypipeline",
									OutputRefs: []string{otherOutput.Name, output.Name},
									SourceType: logging.LogSourceTypeApp,
								},
							},
						}
						normalizedForwardingSpec = request.normalizeLogForwarding(namespace, cluster)
					})
					It("should drop the undefined outputs", func() {
						Expect(normalizedForwardingSpec.Outputs).To(Equal(request.ForwardingSpec.Outputs), "Exp. outputs to be included in the normalized forwarding")
						Expect(normalizedForwardingSpec.Pipelines[0].OutputRefs).To(Equal([]string{output.Name}))
						Expect(normalizedForwardingSpec.Pipelines[0].OutputRefs).NotTo(ContainElement(theInternalOutputName), "Exp. the internal log store to not be part of the pipelines")
					})
				})
				Context("and a pipline spec'd for defined outputs", func() {
					BeforeEach(func() {
						request.ForwardingSpec = logging.ForwardingSpec{
							Outputs: []logging.OutputSpec{
								output,
								otherOutput,
							},
							Pipelines: []logging.PipelineSpec{
								{
									Name:       "mypipeline",
									OutputRefs: []string{output.Name, otherOutput.Name},
									SourceType: logging.LogSourceTypeApp,
								},
							},
						}
						normalizedForwardingSpec = request.normalizeLogForwarding(namespace, cluster)
					})
					It("should forward the pipeline's source to all the spec'd output", func() {
						Expect(normalizedForwardingSpec.Outputs).To(Equal(request.ForwardingSpec.Outputs), "Exp. outputs to be included in the normalized forwarding")
						Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(1), "Expected the pipeline to forward to its output")
						pipeline := normalizedForwardingSpec.Pipelines[0]
						Expect(pipeline.OutputRefs).To(Equal([]string{output.Name, otherOutput.Name}))
						Expect(pipeline.OutputRefs).NotTo(ContainElement(theInternalOutputName), "Exp. the internal log store to not be part of the pipelines")
					})
				})
			})
		})
	})
})
