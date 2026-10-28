package v1

import (
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// structuredLevelRe extracts attempt-4 patterns from the generated VRL, in evaluation order.
var structuredLevelRe = regexp.MustCompile(`match!\(message, r'([^']+)'\) \{\s+level = "([a-z]+)"`)

func firstStructuredLevel(message string) string {
	for _, match := range structuredLevelRe.FindAllStringSubmatch(SetLogLevel, -1) {
		pattern, level := match[1], match[2]
		ok, err := regexp.MatchString(pattern, message)
		Expect(err).NotTo(HaveOccurred())
		if ok {
			return level
		}
	}
	return ""
}

var _ = Describe("#SetLogLevel structured severity", func() {
	DescribeTable("matches structured level tokens case-insensitively", func(message, want string) {
		Expect(firstStructuredLevel(message)).To(Equal(want))
	},
		Entry("uppercase JSON level is not overridden by an earlier severity word", `{"message":"There were no error","level":"INFO"}`, "info"),
		Entry("lowercase JSON level still matches", `{"message":"There were no error","level":"info"}`, "info"),
		Entry("uppercase error token wins over the word error in the message", `{"message":"There were no error","level":"ERROR"}`, "error"),
		Entry("logfmt and Value tokens are case-insensitive", `ts=2026-10-02 level=WARN Value:DEBUG`, "warn"),
		Entry("klog error prefix stays case-sensitive", `E1020 controller.go:10] "level":"info"`, "error"),
		Entry("lowercase klog-like prefix is not a structured error", `e1020 not a klog line`, ""),
	)
})
