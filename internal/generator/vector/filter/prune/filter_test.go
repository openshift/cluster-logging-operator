package prune

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("prune functions", func() {
	DescribeTable("generates correct array of path segments", func(path string, expectedArray []string) {
		Expect(splitPath(path)).To(Equal(expectedArray))
	},
		Entry("with single segment", `.foo`, []string{"foo"}),
		Entry("with 2 segments", `.foo.bar`, []string{"foo", "bar"}),
		Entry("with first segment in quotes", `."@foobar"`, []string{`"@foobar"`}),
		Entry("with 1 quoted segment and one with quotes", `.foo."bar111-22/333"`, []string{"foo", `"bar111-22/333"`}),
		Entry("with 2 non quoted segments and one quoted segment ", `.foo.bar."baz111-22/333"`, []string{"foo", "bar", `"baz111-22/333"`}),
		Entry("with multiple quoted and unquoted segments", `.foo."@some"."d.f.g.o111-22/333".foo_bar`, []string{"foo", `"@some"`, `"d.f.g.o111-22/333"`, "foo_bar"}))

	DescribeTable("generates array with path segments quoted", func(pathSegments []string, expectedArray []string) {
		Expect(quotePathSegments(pathSegments)).To(Equal(expectedArray))
	},
		Entry("", []string{"foo"}, []string{`"foo"`}),
		Entry("", []string{"foo", "bar", `"foo-bar"`}, []string{`"foo"`, `"bar"`, `"foo-bar"`}),
	)

	It("should generate string of an array of quoted path segments from dot-delimited path expressions", func() {
		pathExpression := []string{`.foo.bar."foo.bar.baz-ok".foo123."bar/baz0-9.test"`, `.foo.bar`}
		expectedString := `[["foo","bar","foo.bar.baz-ok","foo123","bar/baz0-9.test"],["foo","bar"]]`
		Expect(generateQuotedPathSegmentArrayStr(pathExpression)).To(Equal(expectedString))
	})

})
