package template

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("Vector Output Template", func() {
	Context("GroupNames", func() {
		DescribeTable("transforms cloudwatch group name to VRL compatible string", func(expVRL, groupNameString string) {
			Expect(TransformUserTemplateToVRL(groupNameString)).To(EqualTrimLines(expVRL))
		},
			Entry("should transform group name with static and dynamic values into VRL compatible syntax",
				`"foo-" + to_string!(.log_type||"none") + "." + to_string!(.bar.foo.test||"missing) + "_" + to_string!(.log_type||"none")`,
				`foo-{.log_type||"none"}.{.bar.foo.test||"missing}_{.log_type||"none"}`),

			Entry("should only add quotes and not transform group name if using only a static value", `"foobar-myindex"`, `foobar-myindex`),
			Entry("should transform group_name if only a dynamic value is defined", `to_string!(.foo.bar||"missing")`, `{.foo.bar||"missing"}`),
		)
	})
})
