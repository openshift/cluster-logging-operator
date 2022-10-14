package v1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test"
)

var _ = Describe("ClusterLogging", func() {
	var _ = DescribeTable("Spec Serialization", func(yamlSpec string) {
		spec := &logging.ClusterLoggingSpec{}
		test.MustUnmarshal(yamlSpec, spec)
		Expect(test.YAMLString(spec)).To(MatchYAML(yamlSpec))
	},
		Entry("with loki and vector", `
  managementState: "Managed"
  logStore:
    type: "lokistack"
    lokistack:
      name: lokistack-instance
  collection:
    type: "vector" 
`),
	)
})
