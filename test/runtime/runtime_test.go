package runtime

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
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
