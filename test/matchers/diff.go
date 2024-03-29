package matchers

import (
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/onsi/gomega/matchers"
	"github.com/onsi/gomega/types"
)

// EqualDiff is like Equal but gives cmp.Diff style output.
func EqualDiff(expect interface{}) types.GomegaMatcher {
	return &diffMatcher{matchers.EqualMatcher{Expected: expect}}
}

type diffMatcher struct{ matchers.EqualMatcher }

func (m *diffMatcher) FailureMessage(actual interface{}) (message string) {
	return "Unexpected diff (-expected, +actual):\n" + cmp.Diff(m.EqualMatcher.Expected, actual)
}

// EqualLines matches multi-line text ignoring blank lines (but not leading/trailing space)
// On failure gives a diff-style message useful for long strings.
func EqualLines(expected string) types.GomegaMatcher {
	return &lineMatcher{expected: expected, trim: false}
}

// EqualTrimLines matches multi-line text ignoring blank lines and leading/trailing space.
// On failure gives a diff-style message useful for long strings.
func EqualTrimLines(expected string) types.GomegaMatcher {
	return &lineMatcher{expected: expected, trim: true}
}

type lineMatcher struct {
	expected interface{}
	diff     string
	trim     bool
}

func (m *lineMatcher) Match(actual interface{}) (success bool, err error) {
	m.diff = cmp.Diff(normalize(actual.(string), m.trim), normalize(m.expected.(string), m.trim))
	return m.diff == "", nil
}

func (m *lineMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n%s\nto equal\n%s\nDiff\n%s", actual, m.expected, m.diff)
}

func (m *lineMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return ("Expected differences but none found.")
}

func normalize(in string, trim bool) []string {
	out := []string{}
	for _, line := range strings.Split(in, "\n") {
		if trim {
			line = strings.TrimSpace(line)
		}
		if line != "" {
			out = append(out, strings.TrimSpace(line))
		}
	}
	return out
}
