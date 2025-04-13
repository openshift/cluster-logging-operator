package helpers

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

var _ = Describe("helpers functions", func() {
	Context("#GenerateQuotedPathSegmentArrayStr", func() {
		It("should generate array of path segments and flat path for single value", func() {
			pathExpression := []obs.FieldPath{`.kubernetes.labels.foo`}
			expectedArrayPath := `[["kubernetes","labels","foo"]]`
			expectedFlatPath := `["kubernetes_labels_foo"]`
			arrayPath, flatPath := GenerateQuotedPathSegmentArrayStr(pathExpression)
			Expect(arrayPath).To(Equal(expectedArrayPath))
			Expect(flatPath).To(Equal(expectedFlatPath))
		})
		It("should generate array of path segments and flat path for multiple value", func() {
			pathExpression := []obs.FieldPath{`.kubernetes.labels.foo`, `.kubernetes.labels.bar`}
			expectedArrayPath := `[["kubernetes","labels","foo"],["kubernetes","labels","bar"]]`
			expectedFlatPath := `["kubernetes_labels_foo","kubernetes_labels_bar"]`
			arrayPath, flatPath := GenerateQuotedPathSegmentArrayStr(pathExpression)
			Expect(arrayPath).To(Equal(expectedArrayPath))
			Expect(flatPath).To(Equal(expectedFlatPath))
		})
		It("should generate array of path segments and escaped flat path", func() {
			pathExpression := []obs.FieldPath{`.kubernetes.labels."bar/baz0-9.test"`}
			expectedArrayPath := `[["kubernetes","labels","bar/baz0-9.test"]]`
			expectedFlatPath := `["kubernetes_labels_bar_baz0-9_test"]`
			arrayPath, flatPath := GenerateQuotedPathSegmentArrayStr(pathExpression)
			Expect(arrayPath).To(Equal(expectedArrayPath))
			Expect(flatPath).To(Equal(expectedFlatPath))
		})
		It("should generate array of path segments and escaped flat path", func() {
			pathExpression := []obs.FieldPath{`.foo.bar."foo.bar.baz-ok".foo123."bar/baz0-9.test"`, `.foo.bar`}
			expectedArrayPath := `[["foo","bar","foo.bar.baz-ok","foo123","bar/baz0-9.test"],["foo","bar"]]`
			expectedFlatPath := `["foo_bar_foo_bar_baz-ok_foo123_bar_baz0-9_test","foo_bar"]`
			arrayPath, flatPath := GenerateQuotedPathSegmentArrayStr(pathExpression)
			Expect(arrayPath).To(Equal(expectedArrayPath))
			Expect(flatPath).To(Equal(expectedFlatPath))
		})
	})

	DescribeTable("#SplitPath generates correct array of path segments", func(path string, expectedArray []string) {
		Expect(SplitPath(path)).To(Equal(expectedArray))
	},
		Entry("with single segment", `.foo`, []string{"foo"}),
		Entry("with 2 segments", `.foo.bar`, []string{"foo", "bar"}),
		Entry("with first segment in quotes", `."@foobar"`, []string{`"@foobar"`}),
		Entry("with 1 quoted segment and one with quotes", `.foo."bar111-22/333"`, []string{"foo", `"bar111-22/333"`}),
		Entry("with 2 non quoted segments and one quoted segment ", `.foo.bar."baz111-22/333"`, []string{"foo", "bar", `"baz111-22/333"`}),
		Entry("with multiple quoted and unquoted segments", `.foo."@some"."d.f.g.o111-22/333".foo_bar`, []string{"foo", `"@some"`, `"d.f.g.o111-22/333"`, "foo_bar"}))

	DescribeTable("#QuotePathSegments generates array with path segments quoted", func(pathSegments []string, expectedArray []string) {
		Expect(QuotePathSegments(pathSegments)).To(Equal(expectedArray))
	},
		Entry("single value", []string{"foo"}, []string{`"foo"`}),
		Entry("multiple value", []string{"foo", "bar.zip", `"foo-bar"`}, []string{`"foo"`, `"bar.zip"`, `"foo-bar"`}),
	)
})
