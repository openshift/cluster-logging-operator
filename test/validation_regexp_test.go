package test_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"regexp"
)

var _ = Describe("Test validation regexp from kubebuilder annotation", func() {
	It("Elasticsearch index matching", func() {
		regex := `^([a-zA-Z\d\-_.\/]*)(\{(\.\w+(\.\w+)*(\|\|(\.\w+(\.\w+)*|"[^"]*"))*)\})?([a-zA-Z\d\-_.\/]*)(\{(\.\w+(\.\w+)*(\|\|(\.\w+(\.\w+)*|"[^"]*"))*)\})?([a-zA-Z\d\-_.\/]*)$`
		testStrings := []string{
			"main",
			"{.log_type}-write",
			"foo-{.bar||\"none\"}",
			"{.foo||.bar||\"missing\"}",
			"foo.{.bar.baz||.qux.quux.corge||.grault||\"nil\"}-waldo.fred{.plugh||\"none\"}",
		}

		for _, testString := range testStrings {
			matched, _ := regexp.MatchString(regex, testString)
			Expect(matched).To(BeTrue())
		}
	})
})
