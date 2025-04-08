package source

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("source", func() {
	DescribeTable("#NewKubernetesLogs", func(includes, excludes string, expFile string) {
		exp, err := tomlContent.ReadFile(expFile)
		if err != nil {
			Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", expFile, err))
		}
		id := helpers.MakeID("source", "foo")
		conf := NewKubernetesLogs(id, includes, excludes) //, includeNS, excludes)
		Expect(string(exp)).To(EqualConfigFrom(conf), fmt.Sprintf("for exp. file %s", expFile))
	},
		Entry("should exclude includes/excludes globs from the config when they are empty",
			"",
			"",
			"kubernetes_logs_no_includes_excludes.toml",
		),
		Entry("should use includes/excludes globs from the config when they exist",
			`["/var/log/pods/foo"]`,
			`["/var/log/pods/bar"]`,
			"kubernetes_logs_with_includes.toml",
		),
	)

	DescribeTable("#normalizeNamespace", func(ns, exp string) {
		Expect(exp).To(Equal(normalizeNamespace(ns)), fmt.Sprintf("Exp. %q to be formalized to %q", ns, exp))
	},
		Entry("should format explict namespaces", "foo", "foo_*"),
		Entry("should format single wildcard namespaces", "*f*o*", "*f*o*_*"),
		Entry("should normalize wildcards at the beginning", "**foo", "*foo_*"),
		Entry("should normalize wildcards at the end", "foo**", "foo*_*"),
		Entry("should normalize wildcards in the middle", "f*o", "f*o_*"),
	)

	Describe("#ContainerPathGlobFrom", func() {
		It("should return an empty string when there are no paths", func() {
			Expect(ContainerPathGlobFrom([]string{}, []string{})).To(BeEmpty())
		})
	})
	Describe("#joinContainerPathsForVector", func() {
		It("should return an empty string when there are no paths", func() {
			Expect(joinContainerPathsForVector([]string{})).To(BeEmpty())
		})
		It("should join the paths when paths exist", func() {
			Expect(joinContainerPathsForVector([]string{"a", "b"})).To(Equal(`[a, b]`))
		})
	})
})
