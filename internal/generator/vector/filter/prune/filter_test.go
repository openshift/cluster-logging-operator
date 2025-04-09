package prune

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("prune functions", func() {
	DescribeTable("#splitPath generates correct array of path segments", func(path string, expectedArray []string) {
		Expect(splitPath(path)).To(Equal(expectedArray))
	},
		Entry("with single segment", `.foo`, []string{"foo"}),
		Entry("with 2 segments", `.foo.bar`, []string{"foo", "bar"}),
		Entry("with first segment in quotes", `."@foobar"`, []string{`"@foobar"`}),
		Entry("with 1 quoted segment and one with quotes", `.foo."bar111-22/333"`, []string{"foo", `"bar111-22/333"`}),
		Entry("with 2 non quoted segments and one quoted segment ", `.foo.bar."baz111-22/333"`, []string{"foo", "bar", `"baz111-22/333"`}),
		Entry("with multiple quoted and unquoted segments", `.foo."@some"."d.f.g.o111-22/333".foo_bar`, []string{"foo", `"@some"`, `"d.f.g.o111-22/333"`, "foo_bar"}))

	DescribeTable("#quotePathSegments generates array with path segments quoted", func(pathSegments []string, expectedArray []string) {
		Expect(quotePathSegments(pathSegments)).To(Equal(expectedArray))
	},
		Entry("", []string{"foo"}, []string{`"foo"`}),
		Entry("", []string{"foo", "bar", `"foo-bar"`}, []string{`"foo"`, `"bar"`, `"foo-bar"`}),
	)

	It("#generateQuotedPathSegmentArrayStr should generate string of an array of quoted path segments from dot-delimited path expressions", func() {
		pathExpression := []obs.FieldPath{`.foo.bar."foo.bar.baz-ok".foo123."bar/baz0-9.test"`, `.foo.bar`}
		expectedString := `[["_internal","foo","bar","foo.bar.baz-ok","foo123","bar/baz0-9.test"],["_internal","foo","bar"]]`
		Expect(generateQuotedPathSegmentArrayStr(pathExpression)).To(Equal(expectedString))
	})
	Context("#generateQuotedPathSegmentArrayStr for explicitly dedoted fields", func() {

		It("should do nothing special for path segments where the dedotted labels dont have dots", func() {
			pathExpression := []obs.FieldPath{`.kubernetes.labels.foo`}
			expectedString := `[["_internal","kubernetes","labels","foo"]]`
			Expect(generateQuotedPathSegmentArrayStr(pathExpression)).To(Equal(expectedString))
		})
		It("should generate path segments for the original and dedotted labels", func() {
			pathExpression := []obs.FieldPath{`.kubernetes.labels."bar/baz0-9.test"`}
			expectedString := `[["_internal","kubernetes","labels","bar/baz0-9.test"],["_internal","kubernetes","labels","bar_baz0-9_test"]]`
			Expect(generateQuotedPathSegmentArrayStr(pathExpression)).To(Equal(expectedString), generateQuotedPathSegmentArrayStr(pathExpression))
		})
		It("should generate path segments for the original and dedotted namespace labels", func() {
			pathExpression := []obs.FieldPath{`.kubernetes.namespace_labels."bar/baz0-9.test"`}
			expectedString := `[["_internal","kubernetes","namespace_labels","bar/baz0-9.test"],["_internal","kubernetes","namespace_labels","bar_baz0-9_test"]]`
			Expect(generateQuotedPathSegmentArrayStr(pathExpression)).To(Equal(expectedString))
		})

	})

	Context("#VRL", func() {
		It("should generate valid VRL for pruning", func() {
			spec := &obs.PruneFilterSpec{
				In:    []obs.FieldPath{".log_type", ".message", ".kubernetes.container_name"},
				NotIn: []obs.FieldPath{`.kubernetes.labels."foo-bar/baz"`, ".level"},
			}
			Expect(NewFilter(spec).VRL()).To(matchers.EqualTrimLines(`
notIn = [["_internal","kubernetes","labels","foo-bar/baz"],["_internal","kubernetes","labels","foo-bar_baz"],["_internal","level"]]

# Prune keys not in notIn list
new_object = {}
for_each(notIn) -> |_index, pathSeg| {
    val = get(., pathSeg) ?? null
    if !is_null(val) {
        new_object = set!(new_object, pathSeg, val)
    }
}
. = new_object
in = [["_internal","log_type"],["_internal","message"],["_internal","kubernetes","container_name"]]

# Remove keys from in list
for_each(in) -> |_index, val| {
    . = remove!(., val)
}
`))
		})
	})

})
