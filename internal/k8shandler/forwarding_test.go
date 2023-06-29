package k8shandler

import (
	"github.com/openshift/cluster-logging-operator/internal/collector/fluentd"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// Match condition by type, status and reason if reason != "".
// Also match messageRegex if it is not empty.
func matchCondition(t logging.ConditionType, s bool, r logging.ConditionReason, messageRegex string) types.GomegaMatcher {
	var status corev1.ConditionStatus
	if s {
		status = corev1.ConditionTrue
	} else {
		status = corev1.ConditionFalse
	}
	fields := Fields{"Type": Equal(t), "Status": Equal(status)}
	if r != "" {
		fields["Reason"] = Equal(r)
	}
	if messageRegex != "" {
		fields["Message"] = MatchRegexp(messageRegex)
	}
	return MatchFields(IgnoreExtras, fields)
}

func HaveCondition(t logging.ConditionType, s bool, r logging.ConditionReason, messageRegex string) types.GomegaMatcher {
	return ContainElement(matchCondition(t, s, r, messageRegex))
}

var _ = DescribeTable("#generateCollectorConfig",
	func(cluster logging.ClusterLogging, forwardSpec logging.ClusterLogForwarderSpec) {
		clusterRequest := &ClusterLoggingRequest{
			Cluster: &cluster,
			Forwarder: &logging.ClusterLogForwarder{
				ObjectMeta: metav1.ObjectMeta{
					Name:      constants.SingletonName,
					Namespace: constants.OpenshiftNS,
				},
				Spec: forwardSpec,
			},
			ResourceNames: factory.GenerateResourceNames(constants.SingletonName, constants.OpenshiftNS),
			CollectionSpec: &logging.CollectionSpec{
				Type: logging.LogCollectionTypeVector,
			},
		}

		clusterRequest.Client = fake.NewFakeClient(clusterRequest.Cluster) //nolint

		_, err := clusterRequest.generateCollectorConfig()
		Expect(err).To(BeNil(), "Generating the collector config should not produce an error: %s=%s %s=%s", "clusterRequest", test.YAMLString(clusterRequest))
	},
	Entry("Valid collector config", logging.ClusterLogging{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.SingletonName,
			Namespace: constants.OpenshiftNS,
		},
		Spec: logging.ClusterLoggingSpec{
			LogStore: nil,
			Collection: &logging.CollectionSpec{
				Type: "fluentd",
				CollectorSpec: logging.CollectorSpec{
					Resources: &corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							"Memory": fluentd.DefaultMemory,
						},
						Requests: corev1.ResourceList{
							"Memory": fluentd.DefaultMemory,
						},
					},
					NodeSelector: map[string]string{"123": "123"},
				},
			},
		},
	}, logging.ClusterLogForwarderSpec{
		Inputs: []logging.InputSpec{
			{Name: logging.InputNameInfrastructure},
		},
		Outputs: []logging.OutputSpec{
			{
				Name: "foo",
				Type: logging.OutputTypeFluentdForward,
				URL:  "tcp://someplace.domain.com",
			},
		},
		Pipelines: []logging.PipelineSpec{
			{
				InputRefs:  []string{logging.InputNameInfrastructure},
				OutputRefs: []string{"foo"},
			},
		},
	}),
	Entry("Collection not specified. Shouldn't crash", logging.ClusterLogging{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.SingletonName,
			Namespace: constants.OpenshiftNS,
		},
		Spec: logging.ClusterLoggingSpec{
			LogStore: nil,
		},
	}, logging.ClusterLogForwarderSpec{}),
)

var _ = Describe("#generateCollectorConfig", func() {
	Context("not dependent on ClusterLogging instance", func() {
		It("should generate appropriate vector config", func() {
			forwarderSpec := logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{Name: logging.InputNameInfrastructure},
				},
				Outputs: []logging.OutputSpec{
					{
						Name: "foo",
						Type: logging.OutputTypeFluentdForward,
						URL:  "tcp://someplace.domain.com",
					},
				},
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs:  []string{logging.InputNameInfrastructure},
						OutputRefs: []string{"foo"},
					},
				},
			}

			clusterRequest := &ClusterLoggingRequest{
				Forwarder: &logging.ClusterLogForwarder{
					ObjectMeta: metav1.ObjectMeta{
						Name:      constants.SingletonName,
						Namespace: constants.OpenshiftNS,
					},
				},
				ResourceNames: factory.GenerateResourceNames(constants.SingletonName, constants.OpenshiftNS),
				CollectionSpec: &logging.CollectionSpec{
					Type: logging.LogCollectionTypeVector,
				},
			}
			clusterRequest.Forwarder.Spec = forwarderSpec
			clusterRequest.Client = fake.NewFakeClient() //nolint

			_, err := clusterRequest.generateCollectorConfig()
			Expect(err).To(BeNil(), "Generating the collector config should not produce an error: %s=%s %s=%s", "clusterRequest", test.YAMLString(clusterRequest))
		})
	})
})
