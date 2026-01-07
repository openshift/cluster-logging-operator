package helpers

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = DescribeTable("#normalizeNamespace", func(ns, exp string) {
	Expect(exp).To(Equal(normalizeNamespace(ns)), fmt.Sprintf("Exp. %q to be formalized to %q", ns, exp))
},
	Entry("should format explict namespaces", "foo", "foo_*"),
	Entry("should format single wildcard namespaces", "*f*o*", "*f*o*_*"),
	Entry("should normalize wildcards at the beginning", "**foo", "*foo_*"),
	Entry("should normalize wildcards at the end", "foo**", "foo*_*"),
	Entry("should normalize wildcards in the middle", "f*o", "f*o_*"),
)

var _ = Describe("#ContainerPathGlobFrom", func() {
	It("should return an empty string when there are no paths", func() {
		Expect(ContainerPathGlobFrom([]string{}, []string{})).To(BeEmpty())
	})
})

var _ = Describe("#joinContainerPathsForVector", func() {
	It("should return an empty string when there are no paths", func() {
		Expect(joinContainerPathsForVector([]string{})).To(BeEmpty())
	})
	It("should join the paths when paths exist", func() {
		Expect(joinContainerPathsForVector([]string{"a", "b"})).To(Equal(`[a, b]`))
	})
})
