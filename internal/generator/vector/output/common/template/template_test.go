package template

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("Vector Output Template", func() {
	DescribeTable("transforms template syntax to VRL compatible string", func(expVRL, template string) {
		Expect(TransformUserTemplateToVRL(template)).To(EqualTrimLines(expVRL))
	},
		FEntry("should transform template with static and dynamic values into VRL compatible syntax",
			`"foo-" + to_string!(._internal.log_type||"none") + "." + to_string!(._internal.bar.foo.test||"missing) + "_" + to_string!(._internal.log_type||"none")`,
			`foo-{.log_type||"none"}.{.bar.foo.test||"missing}_{.log_type||"none"}`),

		Entry("should only add quotes and not transform template if using only a static value", `"foobar-myindex"`, `foobar-myindex`),
		Entry("should transform template if only a dynamic value is defined", `to_string!(._internal.foo.bar||"missing")`, `{.foo.bar||"missing"}`),
	)
})
