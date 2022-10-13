package runtime

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("Object", func() {
	var (
		clf = NewClusterLogForwarder()
	)

	DescribeTable("Decode",
		func(manifest string, o runtime.Object) {
			got := runtime.Decode(manifest)
			Expect(got).To(EqualDiff(o), "%#v", manifest)
		},
		Entry("YAML string clf", test.YAMLString(clf), clf),
	)
})

var _ = Describe("NewClusterLogging", func() {
	It("should stub a ClusterLogging", func() {
		Expect(NewClusterLogging().Spec).To(Equal(logging.ClusterLoggingSpec{
			Collection: &logging.CollectionSpec{
				Type: logging.LogCollectionTypeFluentd,
			},
			ManagementState: logging.ManagementStateManaged,
		}))
	})
})
