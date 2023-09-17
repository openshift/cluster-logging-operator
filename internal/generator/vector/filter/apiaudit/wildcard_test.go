package apiaudit

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("wildcard matching", func() {
	DescribeTable("generates regexp from wildcard",
		func(wildcard, regexp string) {
			b := &strings.Builder{}
			wildcardRegexp(b, wildcard)
			Expect(regexp).To(Equal(b.String()))
		},
		Entry("", `abc`, `abc`),
		Entry("", `abc*`, `abc.*`),
		Entry("", `a*bc`, `a\*bc`),
		Entry("", `*abc`, `.*abc`),
		Entry("", `*abc*`, `.*abc.*`),
		Entry("", `*`, `.*`))

	DescribeTable("generates regexp from list of wildcards",
		func(wildcards []string, regexp string, match, nomatch []string) {
			Expect(regexp).To(Equal(matchAny(wildcards)))
			re := regexp[2 : len(regexp)-1] //  Strip VRL r'' quotes
			for _, s := range match {
				Expect(s).To(MatchRegexp(re))
			}
			for _, s := range nomatch {
				Expect(s).NotTo(MatchRegexp(re))
			}
		},
		Entry("", []string{"*a*", "b*", "*c", "d"}, `r'^(.*a.*|b.*|.*c|d)$'`, []string{"1a2", "b", "b/foo", "foo/c"}, []string{"/b", "x/", "foo", ""}),
		Entry("", []string{"*/status"}, `r'^(.*/status)$'`, []string{"foo/status", "/status"}, []string{"status", "/foo/status/x"}))
})
