package k8shandler

import (
	"github.com/openshift/cluster-logging-operator/internal/collector/fluentd"
	"github.com/openshift/cluster-logging-operator/internal/constants"

	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test"

	core "k8s.io/api/core/v1"
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
					Resources: &core.ResourceRequirements{
						Limits: core.ResourceList{
							"Memory": fluentd.DefaultMemory,
						},
						Requests: core.ResourceList{
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
