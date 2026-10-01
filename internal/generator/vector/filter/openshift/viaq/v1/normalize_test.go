package v1

import (
	"regexp"
	"testing"
)

// structuredLevelRe extracts attempt-4 patterns from the generated VRL, in evaluation order.
var structuredLevelRe = regexp.MustCompile(`match!\(message, r'([^']+)'\) \{\s+level = "([a-z]+)"`)

func firstStructuredLevel(message string) string {
	for _, match := range structuredLevelRe.FindAllStringSubmatch(SetLogLevel, -1) {
		pattern, level := match[1], match[2]
		ok, err := regexp.MatchString(pattern, message)
		if err != nil {
			panic(err)
		}
		if ok {
			return level
		}
	}
	return ""
}

func TestStructuredLevelIsCaseInsensitive(t *testing.T) {
	cases := []struct {
		name    string
		message string
		want    string
	}{
		{
			name:    "uppercase JSON level is not overridden by an earlier severity word",
			message: `{"message":"There were no error","level":"INFO"}`,
			want:    "info",
		},
		{
			name:    "lowercase JSON level still matches",
			message: `{"message":"There were no error","level":"info"}`,
			want:    "info",
		},
		{
			name:    "uppercase error token wins over the word error in the message",
			message: `{"message":"There were no error","level":"ERROR"}`,
			want:    "error",
		},
		{
			name:    "logfmt and Value tokens are case-insensitive",
			message: `ts=2026-10-02 level=WARN Value:DEBUG`,
			want:    "warn",
		},
		{
			name:    "klog error prefix stays case-sensitive",
			message: `E1020 controller.go:10] "level":"info"`,
			want:    "error",
		},
		{
			name:    "lowercase klog-like prefix is not a structured error",
			message: `e1020 not a klog line`,
			want:    "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := firstStructuredLevel(tc.message); got != tc.want {
				t.Fatalf("firstStructuredLevel(%q) = %q, want %q", tc.message, got, tc.want)
			}
		})
	}
}
