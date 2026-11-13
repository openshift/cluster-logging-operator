package v1

import (
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// attemptRe extracts a parse_regex pattern (with its named capture group) from the
// generated VRL, in evaluation order.
var attemptRe = regexp.MustCompile(`parse_regex\(message, r'([^']+)'\)`)

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

// klogLevels maps the klog severity letters to the level values assigned by attempt 2.
var klogLevels = map[string]string{"I": "info", "W": "warning", "E": "error", "F": "fatal"}

// firstLevel simulates attempts 1 (logfmt token), 2 (klog header) and 4 (structured
// patterns) of SetLogLevel in evaluation order. Attempts 3 (grok) and 5 (keyword) are
// not simulated.
func firstLevel(message string) string {
	// attempt 1: logfmt-style level token
	for _, match := range attemptRe.FindAllStringSubmatch(SetLogLevel, -1) {
		re, err := regexp.Compile(match[1])
		Expect(err).NotTo(HaveOccurred())
		if m := re.FindStringSubmatch(message); m != nil {
			names := re.SubexpNames()
			for i, name := range names {
				if name == "level" && m[i] != "" {
					return strings.ToLower(m[i])
				}
			}
		}
		break // only the first parse_regex in SetLogLevel is attempt 1
	}

	// attempt 2: klog header
	klogRe := regexp.MustCompile(`(?s)klog header.*?parse_regex\(message, r'([^']+)'\)`)
	if match := klogRe.FindStringSubmatch(SetLogLevel); match != nil {
		re, err := regexp.Compile(match[1])
		Expect(err).NotTo(HaveOccurred())
		if m := re.FindStringSubmatch(message); m != nil {
			names := re.SubexpNames()
			for i, name := range names {
				if name == "level" && m[i] != "" {
					return klogLevels[m[i]]
				}
			}
		}
	}

	// attempt 4: structured patterns
	return firstStructuredLevel(message)
}

var _ = Describe("#SetLogLevel structured severity", func() {
	DescribeTable("matches structured level tokens case-insensitively", func(message, want string) {
		Expect(firstStructuredLevel(message)).To(Equal(want))
	},
		Entry("uppercase JSON level is not overridden by an earlier severity word", `{"message":"There were no error","level":"INFO"}`, "info"),
		Entry("lowercase JSON level still matches", `{"message":"There were no error","level":"info"}`, "info"),
		Entry("uppercase error token wins over the word error in the message", `{"message":"There were no error","level":"ERROR"}`, "error"),
		Entry("logfmt and Value tokens are case-insensitive", `ts=2026-10-02 level=WARN Value:DEBUG`, "debug"),
		Entry("klog error prefix stays case-sensitive", `E1020 controller.go:10] "level":"info"`, "error"),
		Entry("lowercase klog-like prefix is not a structured error", `e1020 not a klog line`, ""),
	)
})

var _ = Describe("#SetLogLevel level cascade", func() {
	DescribeTable("classifies via attempts 1, 2 and 4 in order", func(message, want string) {
		Expect(firstLevel(message)).To(Equal(want))
	},
		// attempt 1: logfmt-style token
		Entry("logfmt level token", `ts=2026-10-02T00:00:00Z level=error msg="Failed to connect"`, "error"),
		Entry("logfmt token wins over a later Value token", `ts=2026-10-02 level=WARN Value:DEBUG`, "warn"),
		Entry("severity key is accepted", `time=12:00:00 severity=CRITICAL code=500`, "critical"),
		Entry("lvl key with quoted value is accepted", `lvl="info" component=api`, "info"),
		Entry("token match is case-insensitive", `LEVEL=Debug caller=main.go:42`, "debug"),
		Entry("token must be delimited", `alevel=error is not a level token`, ""),
		Entry("non-vocabulary value is declined", `level=3 retries=4`, ""),
		Entry("level token embedded in prose is accepted", `failed: level=debug ignored by caller`, "debug"),

		// attempt 2: klog header
		Entry("klog info header", `I0920 14:22:00.089385       1 scheduler.go:592] "Successfully bound pod to node"`, "info"),
		Entry("klog warning header", `W0920 14:22:00.089385 1 scheduler.go:592] warning message`, "warning"),
		Entry("klog fatal header", `F0920 14:22:00.089385 1 scheduler.go:592] fatal message`, "fatal"),
		Entry("klog EM marker falls through to attempt 4", `EM0920 14:22:00.089385 1 scheduler.go:592] message`, "emergency"),
		Entry("klog letter without a full header is not klog", `F0920 14:22 controller.go:592] truncated header`, ""),

		// pathological input (LOG-8640): a multi-MB line with no '=' must not pin a
		// worker; attempts are single linear-time regex scans.
		Entry("multi-MB line with no key-value tokens is unclassified", strings.Repeat("x", 4<<20)+": a very long stack trace line", ""),
	)
})
