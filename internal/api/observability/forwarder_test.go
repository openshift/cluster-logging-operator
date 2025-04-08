package observability

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("[internal][api][observability]", func() {

	Context("#IsValidSpec", func() {

		var (
			forwarder obs.ClusterLogForwarder
		)
		BeforeEach(func() {
			forwarder = obs.ClusterLogForwarder{
				Spec: obs.ClusterLogForwarderSpec{
					Inputs:    []obs.InputSpec{{}},
					Outputs:   []obs.OutputSpec{{}},
					Filters:   []obs.FilterSpec{{}},
					Pipelines: []obs.PipelineSpec{{}},
				},
				Status: obs.ClusterLogForwarderStatus{
					Conditions: []metav1.Condition{
						NewCondition(obs.ConditionTypeAuthorized, obs.ConditionTrue, "", ""),
					},
					InputConditions: []metav1.Condition{
						NewConditionFromPrefix(obs.ConditionTypeValidInputPrefix, "foo", true, "", ""),
					},
					PipelineConditions: []metav1.Condition{
						NewConditionFromPrefix(obs.ConditionTypeValidPipelinePrefix, "foo", true, "", ""),
					},
					FilterConditions: []metav1.Condition{
						NewConditionFromPrefix(obs.ConditionTypeValidFilterPrefix, "foo", true, "", ""),
					},
					OutputConditions: []metav1.Condition{
						NewConditionFromPrefix(obs.ConditionTypeValidOutputPrefix, "foo", true, "", ""),
					},
				},
			}
		})

		It("should be true when the forwarder is authorized and all input, outputs, pipelines and filters are valid", func() {
			Expect(IsValidSpec(forwarder)).To(BeTrue())
		})
		It("should be false when inputs are invalid", func() {
			forwarder.Status.Conditions = []metav1.Condition{
				NewConditionFromPrefix(obs.ConditionTypeValidInputPrefix, "foo", false, "", ""),
			}
			Expect(IsValidSpec(forwarder)).To(BeFalse())
		})
		It("should be false when outputs are invalid", func() {
			forwarder.Status.Conditions = []metav1.Condition{
				NewConditionFromPrefix(obs.ConditionTypeValidOutputPrefix, "foo", false, "", ""),
			}
			Expect(IsValidSpec(forwarder)).To(BeFalse())
		})
		It("should be false when filters are invalid", func() {
			forwarder.Status.Conditions = []metav1.Condition{
				NewConditionFromPrefix(obs.ConditionTypeValidFilterPrefix, "foo", false, "", ""),
			}
			Expect(IsValidSpec(forwarder)).To(BeFalse())
		})
		It("should be false when pipelines are invalid", func() {
			forwarder.Status.Conditions = []metav1.Condition{
				NewConditionFromPrefix(obs.ConditionTypeValidPipelinePrefix, "foo", false, "", ""),
			}
			Expect(IsValidSpec(forwarder)).To(BeFalse())
		})
		It("should be false when the forwarder is not authorized", func() {
			forwarder.Status.Conditions = []metav1.Condition{
				NewCondition(obs.ConditionTypeAuthorized, obs.ConditionFalse, "", ""),
			}
			Expect(IsValidSpec(forwarder)).To(BeFalse())
		})
	})

	Context("#DeployAsDeployment", func() {
		var (
			forwarder obs.ClusterLogForwarder
		)
		BeforeEach(func() {
			forwarder = *obsruntime.NewClusterLogForwarder(constants.OpenshiftNS, constants.SingletonName, runtime.Initialize)
			forwarder.Spec.Inputs = []obs.InputSpec{
				{Type: obs.InputTypeApplication},
				{Type: obs.InputTypeReceiver},
				{Type: obs.InputTypeReceiver},
			}
		})

		It("should not be a deployment if the annotation does not enable the feature", func() {
			Expect(DeployAsDeployment(forwarder)).To(BeFalse())
		})

		Context("when the forwarder is annotated to enable the feature", func() {
			BeforeEach(func() {
				forwarder.Annotations = map[string]string{constants.AnnotationEnableCollectorAsDeployment: "true"}
			})
			It("should be true when there are only receiver inputs", func() {
				forwarder.Spec.Inputs = []obs.InputSpec{
					{Type: obs.InputTypeReceiver},
					{Type: obs.InputTypeReceiver},
				}
				Expect(DeployAsDeployment(forwarder)).To(BeTrue())
			})
			It("should be false when there are more then just receiver inputs", func() {
				Expect(DeployAsDeployment(forwarder)).To(BeFalse())
			})
		})
	})
})
