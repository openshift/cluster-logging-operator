package k8shandler

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	helpers "github.com/openshift/cluster-logging-operator/test"
)

const (
	namespace             = "aNamespace"
	otherTargetName       = "someothername"
	theInternalOutputName = "clo-default-output-es"
)

var _ = Describe("Default secure-forward.conf hash", func() {
	It("should remain unchanged so we can determine how to upgrade", func() {
		//sanity check to ensure it does not change without intention
		Expect("8163d9a59a20ada8ab58c2535a3a4924").To(Equal(secureForwardConfHash))
		file := string(utils.GetFileContents(utils.GetShareDir() + "/fluentd/secure-forward.conf"))
		Expect(utils.CalculateMD5Hash(file)).To(Equal(secureForwardConfHash))
	})

})

func HasOutputStatus(status *logging.ForwardingStatus, outputName string, state logging.OutputState, reason logging.OutputStateReason) bool {
	logger.Debugf("Status: %v", status.Outputs)
	for _, output := range status.Outputs {
		if output.Name == outputName && output.State == state {
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
	)
	BeforeEach(func() {
		cluster = &logging.ClusterLogging{}
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
						OutputRefs: []string{output.Name, otherOutput.Name},
						SourceType: logging.LogSourceTypeApp,
					},
				},
			}
		})

		It("should drop outputs that do not have unique names", func() {
			cluster.Spec.Forwarding.Outputs = append(cluster.Spec.Forwarding.Outputs, logging.OutputSpec{
				Name:     "myOutput",
				Type:     logging.OutputTypeElasticsearch,
				Endpoint: "anOutPut",
			})
			//sanity check
			Expect(len(cluster.Spec.Forwarding.Outputs)).To(Equal(3))
			normalizedForwardingSpec := normalizeLogForwarding(namespace, cluster)
			Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. non-unique outputs to be dropped")
			Expect(HasOutputStatus(cluster.Status.Forwarding, "output[2]", logging.OutputStateDropped, logging.OutputStateNonUniqueName)).To(BeTrue(), "Exp. the status to be updated")
		})

		It("should drop outputs that have empty names", func() {
			cluster.Spec.Forwarding.Outputs = append(cluster.Spec.Forwarding.Outputs, logging.OutputSpec{
				Name:     "",
				Type:     logging.OutputTypeElasticsearch,
				Endpoint: "anOutPut",
			})
			normalizedForwardingSpec := normalizeLogForwarding(namespace, cluster)
			Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with an empty name to be dropped")
			Expect(HasOutputStatus(cluster.Status.Forwarding, "output[2]", logging.OutputStateDropped, logging.OutputStateReasonMissingName)).To(BeTrue(), "Exp. the status to be updated")
		})

		It("should drop outputs that conflict with the internally reserved name", func() {
			cluster.Spec.Forwarding.Outputs = append(cluster.Spec.Forwarding.Outputs, logging.OutputSpec{
				Name:     internalOutputName,
				Type:     logging.OutputTypeElasticsearch,
				Endpoint: "anOutPut",
			})
			normalizedForwardingSpec := normalizeLogForwarding(namespace, cluster)
			Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with an internal name conflict to be dropped")
			Expect(HasOutputStatus(cluster.Status.Forwarding, "output[2]", logging.OutputStateDropped, logging.OutputStateReservedNameConflict)).To(BeTrue(), "Exp. the status to be updated")
		})

		It("should drop outputs that have empty types", func() {
			cluster.Spec.Forwarding.Outputs = append(cluster.Spec.Forwarding.Outputs, logging.OutputSpec{
				Name:     "aName",
				Type:     "",
				Endpoint: "anOutPut",
			})
			normalizedForwardingSpec := normalizeLogForwarding(namespace, cluster)
			Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with an empty type to be dropped")
			Expect(HasOutputStatus(cluster.Status.Forwarding, "aName", logging.OutputStateDropped, logging.OutputStateReasonMissingType)).To(BeTrue(), "Exp. the status to be updated")
		})
		It("should drop outputs that have unrecognized types", func() {
			cluster.Spec.Forwarding.Outputs = append(cluster.Spec.Forwarding.Outputs, logging.OutputSpec{
				Name:     "aName",
				Type:     "foo",
				Endpoint: "anOutPut",
			})
			normalizedForwardingSpec := normalizeLogForwarding(namespace, cluster)
			Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with an unrecognized type to be dropped")
			Expect(HasOutputStatus(cluster.Status.Forwarding, "aName", logging.OutputStateDropped, logging.OutputStateReasonUnrecognizedType)).To(BeTrue(), "Exp. the status to be updated")
		})
		It("should drop outputs that have empty endpoints", func() {
			cluster.Spec.Forwarding.Outputs = append(cluster.Spec.Forwarding.Outputs, logging.OutputSpec{
				Name:     "aName",
				Type:     "foo",
				Endpoint: "",
			})
			normalizedForwardingSpec := normalizeLogForwarding(namespace, cluster)
			Expect(len(normalizedForwardingSpec.Outputs)).To(Equal(2), "Exp. outputs with an empty endpoint to be dropped")
			Expect(HasOutputStatus(cluster.Status.Forwarding, "aName", logging.OutputStateDropped, logging.OutputStateReasonMissingEndpoint)).To(BeTrue(), "Exp. the status to be updated")
		})

	})

	Context("and a logstore is not spec'd", func() {
		It("should return an empty forwarding spec", func() {
			spec := normalizeLogForwarding(namespace, cluster)
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
				normalizedForwardingSpec := normalizeLogForwarding(namespace, cluster)
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
					normalizedForwardingSpec := normalizeLogForwarding(namespace, cluster)
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
									OutputRefs: []string{output.Name},
									SourceType: logging.LogSourceTypeApp,
								},
							},
						}
						normalizedForwardingSpec = normalizeLogForwarding(namespace, cluster)
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
									OutputRefs: []string{otherOutput.Name, output.Name},
									SourceType: logging.LogSourceTypeApp,
								},
							},
						}
						normalizedForwardingSpec = normalizeLogForwarding(namespace, cluster)
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
									OutputRefs: []string{output.Name, otherOutput.Name},
									SourceType: logging.LogSourceTypeApp,
								},
							},
						}
						normalizedForwardingSpec = normalizeLogForwarding(namespace, cluster)
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
