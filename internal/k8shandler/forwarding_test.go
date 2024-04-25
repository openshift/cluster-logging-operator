package k8shandler

import (
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/test"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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

			clf := logging.ClusterLogForwarder{
				ObjectMeta: metav1.ObjectMeta{
					Name:      constants.SingletonName,
					Namespace: constants.OpenshiftNS,
				},
			}
			clusterRequest := &ClusterLoggingRequest{
				Cluster: &logging.ClusterLogging{
					Spec: logging.ClusterLoggingSpec{
						Collection: &logging.CollectionSpec{
							Type: logging.LogCollectionTypeVector,
						},
					},
				},
				Forwarder:     &clf,
				ResourceNames: factory.GenerateResourceNames(clf),
			}
			clusterRequest.Forwarder.Spec = forwarderSpec
			clusterRequest.Client = fake.NewFakeClient() //nolint

			_, err := clusterRequest.generateCollectorConfig()
			Expect(err).To(BeNil(), "Generating the collector config should not produce an error: %s=%s %s=%s", "clusterRequest", test.YAMLString(clusterRequest))
		})
	})
})
