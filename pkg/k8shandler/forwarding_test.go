package k8shandler

import (
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	cl "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	helpers "github.com/openshift/cluster-logging-operator/test"
	"k8s.io/client-go/kubernetes/scheme"
)

const (
	namespace             = "aNamespace"
	otherTargetName       = "someothername"
	theInternalOutputName = "clo-default-output-es"
)

func HasPipelineStatus(status *loggingv1alpha1.ForwardingStatus, pipelineName string, state loggingv1alpha1.PipelineState, reason loggingv1alpha1.PipelineConditionReason) bool {
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
func HasOutputStatus(status *loggingv1alpha1.ForwardingStatus, outputName string, state loggingv1alpha1.OutputState, reason loggingv1alpha1.OutputConditionReason, skipReason bool) bool {
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
		normalizedForwardingSpec loggingv1alpha1.ForwardingSpec
		output                   loggingv1alpha1.OutputSpec
		otherOutput              loggingv1alpha1.OutputSpec
		request                  *ClusterLoggingRequest
	)
	BeforeEach(func() {
		output = loggingv1alpha1.OutputSpec{
			Name:     "myOutput",
			Type:     loggingv1alpha1.OutputTypeElasticsearch,
			Endpoint: "anOutPut",
		}
		otherOutput = loggingv1alpha1.OutputSpec{
			Name:     otherTargetName,
			Type:     loggingv1alpha1.OutputTypeElasticsearch,
			Endpoint: "someotherendpoint",
		}
		request = &ClusterLoggingRequest{
			Client: fake.NewFakeClient(),
			Cluster: &cl.ClusterLogging{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						ForwardingAnnotation: "enabled",
					},
				},
			},
			ForwardingRequest: &loggingv1alpha1.LogForwarding{},
		}
		cluster = request.Cluster
	})

	It("should have sourceType application", func() {
		Expect(sourceTypes.Has(string(loggingv1alpha1.LogSourceTypeApp))).To(BeTrue())
	})
	It("should have sourceType infra", func() {
		Expect(sourceTypes.Has(string(loggingv1alpha1.LogSourceTypeInfra))).To(BeTrue())
	})
	It("should have outputType Elastic", func() {
		Expect(outputTypes.Has(string(loggingv1alpha1.OutputTypeElasticsearch))).To(BeTrue())
	})
	It("should have outputType Forward", func() {
		Expect(outputTypes.Has(string(loggingv1alpha1.OutputTypeForward))).To(BeTrue())
	})

	Context("while validating ", func() {

		BeforeEach(func() {
			request.ForwardingSpec = loggingv1alpha1.ForwardingSpec{
				Outputs: []loggingv1alpha1.OutputSpec{
					output,
					otherOutput,
				},
				Pipelines: []loggingv1alpha1.PipelineSpec{
					{
						Name:       "aPipeline",
						OutputRefs: []string{output.Name, otherOutput.Name},
						SourceType: loggingv1alpha1.LogSourceTypeApp,
					},
				},
			}
		})
		Context("pipelines", func() {

			It("should only include logsources if there is atleast one valid pipeline", func() {
				request.ForwardingSpec.Pipelines = []loggingv1alpha1.PipelineSpec{
					{
						Name:       "aPipeline",
						OutputRefs: []string{"someotherendpoint"},
						SourceType: loggingv1alpha1.LogSourceTypeApp,
					},
				}
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(0), "Exp. all pipelines to be dropped")
				Expect(request.ForwardingRequest.Status.LogSources).To(Equal([]loggingv1alpha1.LogSourceType{}), "Exp. no log sources")
			})

			It("should drop pipelines that do not have unique names", func() {
				request.ForwardingSpec.Pipelines = append(request.ForwardingSpec.Pipelines,
					loggingv1alpha1.PipelineSpec{
						Name:       "aPipeline",
						OutputRefs: []string{output.Name, otherOutput.Name},
						SourceType: loggingv1alpha1.LogSourceTypeApp,
					})
				//sanity check
				Expect(len(request.ForwardingSpec.Pipelines)).To(Equal(2))
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(1), "Exp. non-unique pipelines to be dropped")
				Expect(HasPipelineStatus(request.ForwardingRequest.Status, "pipeline[1]", loggingv1alpha1.PipelineStateDropped, loggingv1alpha1.PipelineConditionReasonUniqueName)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should drop pipelines that have empty names", func() {
				request.ForwardingSpec.Pipelines = append(request.ForwardingSpec.Pipelines,
					loggingv1alpha1.PipelineSpec{
						Name:       "",
						OutputRefs: []string{output.Name, otherOutput.Name},
						SourceType: loggingv1alpha1.LogSourceTypeApp,
					})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				logger.Debug(normalizedForwardingSpec.Pipelines)
				Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(1), "Exp. empty pipelines to be dropped")
				Expect(HasPipelineStatus(request.ForwardingRequest.Status, "pipeline[1]", loggingv1alpha1.PipelineStateDropped, loggingv1alpha1.PipelineConditionReasonMissingName)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should drop pipelines that conflict with the internally reserved name", func() {
				request.ForwardingSpec.Pipelines = append(request.ForwardingSpec.Pipelines,
					loggingv1alpha1.PipelineSpec{
						Name:       defaultAppPipelineName,
						OutputRefs: []string{output.Name, otherOutput.Name},
						SourceType: loggingv1alpha1.LogSourceTypeApp,
					})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(1), "Exp. pipelines with an internal name conflict to be dropped")
				Expect(HasPipelineStatus(request.ForwardingRequest.Status, "pipeline[1]", loggingv1alpha1.PipelineStateDropped, loggingv1alpha1.PipelineConditionReasonReservedNameConflict)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should drop pipelines that have missing sources", func() {
				request.ForwardingSpec.Pipelines = append(request.ForwardingSpec.Pipelines,
					loggingv1alpha1.PipelineSpec{
						Name:       "someDefinedPipeline",
						OutputRefs: []string{output.Name, otherOutput.Name},
						SourceType: "",
					})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(1), "Exp. pipelines with an empty source to be dropped")
				Expect(HasPipelineStatus(request.ForwardingRequest.Status, "someDefinedPipeline", loggingv1alpha1.PipelineStateDropped, loggingv1alpha1.PipelineConditionReasonMissingSource)).To(BeTrue(), "Exp. the status to be updated")
			})
			It("should drop pipelines that have unrecognized sources", func() {
				request.ForwardingSpec.Pipelines = append(request.ForwardingSpec.Pipelines,
					loggingv1alpha1.PipelineSpec{
						Name:       "someDefinedPipeline",
						OutputRefs: []string{output.Name, otherOutput.Name},
						SourceType: "foo",
					})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(1), "Exp. outputs with an unrecognized type to be dropped")
				Expect(HasPipelineStatus(request.ForwardingRequest.Status, "someDefinedPipeline", loggingv1alpha1.PipelineStateDropped, loggingv1alpha1.PipelineConditionReasonUnrecognizedSourceType)).To(BeTrue(), "Exp. the status to be updated")
			})
			It("should drop pipelines that have no outputRefs", func() {
				request.ForwardingSpec.Pipelines = append(request.ForwardingSpec.Pipelines,
					loggingv1alpha1.PipelineSpec{
						Name:       "someDefinedPipeline",
						OutputRefs: []string{},
						SourceType: loggingv1alpha1.LogSourceTypeApp,
					})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(1), "Exp. outputs with an unrecognized type to be dropped")
				Expect(HasPipelineStatus(request.ForwardingRequest.Status, "someDefinedPipeline", loggingv1alpha1.PipelineStateDropped, loggingv1alpha1.PipelineConditionReasonMissingOutputs)).To(BeTrue(), "Exp. the status to be updated")
			})
			It("should degrade pipelines where there are fewer outputs then defined outputRefs", func() {
				request.ForwardingSpec.Pipelines = append(request.ForwardingSpec.Pipelines,
					loggingv1alpha1.PipelineSpec{
						Name:       "someDefinedPipeline",
						OutputRefs: []string{output.Name, otherOutput.Name, "aMissingOutput"},
						SourceType: loggingv1alpha1.LogSourceTypeApp,
					})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Pipelines)).To(Equal(2), "Exp. all defined pipelines")
				Expect(HasPipelineStatus(request.ForwardingRequest.Status, "someDefinedPipeline", loggingv1alpha1.PipelineStateDegraded, loggingv1alpha1.PipelineConditionReasonMissingOutputs)).To(BeTrue(), "Exp. the status to be updated")
			})

		})

		Context("outputs", func() {

			It("should drop outputs that do not have unique names", func() {
				request.ForwardingSpec.Outputs = append(request.ForwardingSpec.Outputs, loggingv1alpha1.OutputSpec{
					Name:     "myOutput",
					Type:     loggingv1alpha1.OutputTypeElasticsearch,
					Endpoint: "anOutPut",
				})
				//sanity check
				Expect(len(request.ForwardingSpec.Outputs)).To(Equal(3))
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. non-unique outputs to be dropped")
				Expect(HasOutputStatus(request.ForwardingRequest.Status, "output[2]", loggingv1alpha1.OutputStateDropped, loggingv1alpha1.OutputConditionReasonNonUniqueName, false)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should drop outputs that have empty names", func() {
				request.ForwardingSpec.Outputs = append(request.ForwardingSpec.Outputs, loggingv1alpha1.OutputSpec{
					Name:     "",
					Type:     loggingv1alpha1.OutputTypeElasticsearch,
					Endpoint: "anOutPut",
				})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with an empty name to be dropped")
				Expect(HasOutputStatus(request.ForwardingRequest.Status, "output[2]", loggingv1alpha1.OutputStateDropped, loggingv1alpha1.OutputConditionReasonMissingName, false)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should drop outputs that conflict with the internally reserved name", func() {
				request.ForwardingSpec.Outputs = append(request.ForwardingSpec.Outputs, loggingv1alpha1.OutputSpec{
					Name:     internalOutputName,
					Type:     loggingv1alpha1.OutputTypeElasticsearch,
					Endpoint: "anOutPut",
				})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with an internal name conflict to be dropped")
				Expect(HasOutputStatus(request.ForwardingRequest.Status, "output[2]", loggingv1alpha1.OutputStateDropped, loggingv1alpha1.OutputConditionReasonReservedNameConflict, false)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should drop outputs that have empty types", func() {
				request.ForwardingSpec.Outputs = append(request.ForwardingSpec.Outputs, loggingv1alpha1.OutputSpec{
					Name:     "aName",
					Type:     "",
					Endpoint: "anOutPut",
				})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with an empty type to be dropped")
				Expect(HasOutputStatus(request.ForwardingRequest.Status, "aName", loggingv1alpha1.OutputStateDropped, loggingv1alpha1.OutputConditionReasonMissingType, false)).To(BeTrue(), "Exp. the status to be updated")
			})
			It("should drop outputs that have unrecognized types", func() {
				request.ForwardingSpec.Outputs = append(request.ForwardingSpec.Outputs, loggingv1alpha1.OutputSpec{
					Name:     "aName",
					Type:     "foo",
					Endpoint: "anOutPut",
				})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with an unrecognized type to be dropped")
				Expect(HasOutputStatus(request.ForwardingRequest.Status, "aName", loggingv1alpha1.OutputStateDropped, loggingv1alpha1.OutputConditionReasonUnrecognizedType, false)).To(BeTrue(), "Exp. the status to be updated")
			})
			It("should drop outputs that have empty endpoints", func() {
				request.ForwardingSpec.Outputs = append(request.ForwardingSpec.Outputs, loggingv1alpha1.OutputSpec{
					Name:     "aName",
					Type:     "fluentForward",
					Endpoint: "",
				})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with an empty endpoint to be dropped")
				Expect(HasOutputStatus(request.ForwardingRequest.Status, "aName", loggingv1alpha1.OutputStateDropped, loggingv1alpha1.OutputConditionReasonMissingEndpoint, false)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should drop outputs that have secrets with no names", func() {
				request.ForwardingSpec.Outputs = append(request.ForwardingSpec.Outputs, loggingv1alpha1.OutputSpec{
					Name:     "aName",
					Type:     loggingv1alpha1.OutputTypeElasticsearch,
					Endpoint: "an output",
					Secret:   &loggingv1alpha1.OutputSecretSpec{},
				})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with empty secrets to be dropped")
				Expect(HasOutputStatus(request.ForwardingRequest.Status, "aName", loggingv1alpha1.OutputStateDropped, loggingv1alpha1.OutputConditionReasonMissingSecretName, false)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should drop outputs that have secrets which don't exist", func() {
				request.ForwardingSpec.Outputs = append(request.ForwardingSpec.Outputs, loggingv1alpha1.OutputSpec{
					Name:     "aName",
					Type:     loggingv1alpha1.OutputTypeElasticsearch,
					Endpoint: "an output",
					Secret: &loggingv1alpha1.OutputSecretSpec{
						Name: "mysecret",
					},
				})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with non-existent secrets to be dropped")
				Expect(HasOutputStatus(request.ForwardingRequest.Status, "aName", loggingv1alpha1.OutputStateDropped, loggingv1alpha1.OutputConditionReasonSecretDoesNotExist, false)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should drop forward outputs that have secrets and is missing shared_key", func() {
				secret := &core.Secret{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: core.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mysecret",
						Namespace: "openshift-loggingv1alpha1",
					},
					Type: "Opaque",
					Data: map[string][]byte{
						"foo": []byte("bar"),
					},
				}
				request.Client = fake.NewFakeClient(secret)
				request.ForwardingSpec.Outputs = append(request.ForwardingSpec.Outputs, loggingv1alpha1.OutputSpec{
					Name:     "aName",
					Type:     loggingv1alpha1.OutputTypeForward,
					Endpoint: "an output",
					Secret: &loggingv1alpha1.OutputSecretSpec{
						Name: "mysecret",
					},
				})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with missing shared_key to be dropped")
				Expect(HasOutputStatus(request.ForwardingRequest.Status, "aName", loggingv1alpha1.OutputStateDropped, loggingv1alpha1.OutputConditionReasonSecretMissingSharedKey, false)).To(BeTrue(), "Exp. the status to be updated")
			})

			It("should accept well formed outputs", func() {
				request.Client = fake.NewFakeClient(&core.Secret{
					TypeMeta: metav1.TypeMeta{
						Kind:       "Secret",
						APIVersion: core.SchemeGroupVersion.String(),
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      "mysecret",
						Namespace: "openshift-loggingv1alpha1",
					},
				})
				request.ForwardingSpec.Outputs = append(request.ForwardingSpec.Outputs, loggingv1alpha1.OutputSpec{
					Name:     "aName",
					Type:     loggingv1alpha1.OutputTypeElasticsearch,
					Endpoint: "an output",
					Secret: &loggingv1alpha1.OutputSecretSpec{
						Name: "mysecret",
					},
				})
				normalizedForwardingSpec := request.normalizeLogForwarding(namespace, cluster)
				Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(3))
				Expect(HasOutputStatus(request.ForwardingRequest.Status, "aName", loggingv1alpha1.OutputStateAccepted, loggingv1alpha1.OutputConditionReasonSecretDoesNotExist, true)).To(BeTrue(), "Exp. the status to be updated")
			})

		})
	})

	Context("and a logstore is not spec'd", func() {
		It("should return an empty forwarding spec", func() {
			spec := request.normalizeLogForwarding(namespace, cluster)
			Expect(spec).To(Equal(loggingv1alpha1.ForwardingSpec{Outputs: []loggingv1alpha1.OutputSpec{}, Pipelines: nil}))
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
				Expect(helpers.StringsContain(sources, string(loggingv1alpha1.LogSourceTypeApp))).To(BeTrue(), "Exp. the internal pipeline to include app logs")
				Expect(helpers.StringsContain(sources, string(loggingv1alpha1.LogSourceTypeInfra))).To(BeTrue(), "Exp. the internal pipeline to include infa logs")
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
					Expect(helpers.StringsContain(sources, string(loggingv1alpha1.LogSourceTypeApp))).To(BeTrue(), "Exp. the internal pipeline to include app logs")
					Expect(helpers.StringsContain(sources, string(loggingv1alpha1.LogSourceTypeInfra))).To(BeTrue(), "Exp. the internal pipeline to include infa logs")
				})
			})
			Context("and disableDefaultForwarding is true", func() {

				Context("and a pipline spec'd for undefined outputs", func() {
					BeforeEach(func() {
						request.ForwardingSpec = loggingv1alpha1.ForwardingSpec{
							Pipelines: []loggingv1alpha1.PipelineSpec{
								{
									Name:       "mypipeline",
									OutputRefs: []string{output.Name},
									SourceType: loggingv1alpha1.LogSourceTypeApp,
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
						request.ForwardingSpec = loggingv1alpha1.ForwardingSpec{
							Outputs: []loggingv1alpha1.OutputSpec{output},
							Pipelines: []loggingv1alpha1.PipelineSpec{
								{
									Name:       "mypipeline",
									OutputRefs: []string{otherOutput.Name, output.Name},
									SourceType: loggingv1alpha1.LogSourceTypeApp,
								},
							},
						}
						normalizedForwardingSpec = request.normalizeLogForwarding(namespace, cluster)
					})
					It("should drop the undefined outputs", func() {
						Expect(normalizedForwardingSpec.Outputs).To(Equal(request.ForwardingSpec.Outputs), "Exp. outputs to be included in the normalized forwarding")
						Expect(normalizedForwardingSpec.Pipelines[0].OutputRefs).To(Equal([]string{output.Name}))
						Expect(helpers.StringsContain(normalizedForwardingSpec.Pipelines[0].OutputRefs, theInternalOutputName)).To(Not(BeTrue()), "Exp. the internal log store to not be part of the pipelines")
					})
				})
				Context("and a pipline spec'd for defined outputs", func() {
					BeforeEach(func() {
						request.ForwardingSpec = loggingv1alpha1.ForwardingSpec{
							Outputs: []loggingv1alpha1.OutputSpec{
								output,
								otherOutput,
							},
							Pipelines: []loggingv1alpha1.PipelineSpec{
								{
									Name:       "mypipeline",
									OutputRefs: []string{output.Name, otherOutput.Name},
									SourceType: loggingv1alpha1.LogSourceTypeApp,
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
						Expect(helpers.StringsContain(pipeline.OutputRefs, theInternalOutputName)).To(Not(BeTrue()), "Exp. the internal log store to not be part of the pipelines")
					})
				})
			})
		})
	})
})

func TestClusterLoggingRequest_generateCollectorConfig(t *testing.T) {
	_ = loggingv1.SchemeBuilder.AddToScheme(scheme.Scheme)
	_ = loggingv1alpha1.SchemeBuilder.AddToScheme(scheme.Scheme)

	type fields struct {
		client            client.Client
		cluster           *loggingv1.ClusterLogging
		ForwardingRequest *loggingv1alpha1.LogForwarding
		ForwardingSpec    loggingv1alpha1.ForwardingSpec
		Collector         *loggingv1alpha1.CollectorSpec
	}
	tests := []struct {
		name       string
		fields     fields
		wantConfig string
		wantErr    bool
	}{
		{
			name: "Valid collector config",
			fields: fields{
				cluster: &loggingv1.ClusterLogging{
					Spec: loggingv1.ClusterLoggingSpec{
						LogStore: nil,
						Collection: &loggingv1.CollectionSpec{
							Logs: loggingv1.LogCollectionSpec{
								Type: "fluentd",
								FluentdSpec: loggingv1.FluentdSpec{
									Resources: &core.ResourceRequirements{
										Limits: core.ResourceList{
											"Memory": defaultFluentdMemory,
										},
										Requests: core.ResourceList{
											"Memory": defaultFluentdMemory,
										},
									},
									NodeSelector: map[string]string{"123": "123"},
								},
							},
						},
					},
				},
				ForwardingRequest: nil,
				ForwardingSpec:    loggingv1alpha1.ForwardingSpec{},
			},
		},
		{
			name: "Collection not specified. Shouldn't crash",
			fields: fields{
				cluster: &loggingv1.ClusterLogging{
					Spec: loggingv1.ClusterLoggingSpec{
						LogStore: nil,
					},
				},
				ForwardingRequest: nil,
				ForwardingSpec:    loggingv1alpha1.ForwardingSpec{},
				Collector:         nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clusterRequest := &ClusterLoggingRequest{
				Client:            tt.fields.client,
				Cluster:           tt.fields.cluster,
				ForwardingRequest: tt.fields.ForwardingRequest,
				ForwardingSpec:    tt.fields.ForwardingSpec,
				Collector:         tt.fields.Collector,
			}

			config := &core.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: "secure-forward",
				},
				Data:       map[string]string{},
				BinaryData: nil,
			}

			clusterRequest.Client = fake.NewFakeClient(tt.fields.cluster, config)

			gotConfig, err := clusterRequest.generateCollectorConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("generateCollectorConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotConfig != tt.wantConfig {
				t.Errorf("generateCollectorConfig() gotConfig = %v, want %v", gotConfig, tt.wantConfig)
			}
		})
	}
}
