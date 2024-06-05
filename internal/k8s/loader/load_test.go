package loader

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
)

var _ = Describe("#FetchClusterLogForwarder", func() {

	var (
		clf      *logging.ClusterLogForwarder
		k8Client client.Client
	)

	Context("when retrieving a legacy CLF without existing CL", func() {

		BeforeEach(func() {
			k8Client = fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
			clf = runtime.NewClusterLogForwarder(constants.OpenshiftNS, constants.SingletonName)
			clf.Spec = logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						Name:       "all-logs-to-default",
						InputRefs:  logging.ReservedInputNames.List(),
						OutputRefs: []string{logging.OutputNameDefault},
					},
				},
			}
		})

		// https://issues.redhat.com/browse/LOG-4564
		It("should return with validation errors", func() {
			_, err := FetchClusterLogForwarder(k8Client, clf.Namespace, clf.Name)
			Expect(err).ToNot(BeNil(), "legacy CLF without CL should fail with validation errors")
		})
	})
})
