package matchers

import (
	"fmt"
	"time"

	"github.com/onsi/gomega/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LastTransitionTimeMatcher struct {
	expected metav1.Time
}

func MatchLastTransitionTime(expected metav1.Time) types.GomegaMatcher {
	return &LastTransitionTimeMatcher{
		expected: expected,
	}
}

// Match returns success either if both timestamps are identical. If expected.Time is time 0 UTC, it will check if
// actual.Time has a time stamp within the last 5 minutes.
func (m *LastTransitionTimeMatcher) Match(a interface{}) (success bool, err error) {
	expected := m.expected
	actual, ok := a.(metav1.Time)
	if !ok {
		return false, fmt.Errorf("Matcher expects metav1.Time")
	}
	if expected.Time == time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC) {
		if actual.Time.Before(time.Now().Add(time.Minute * time.Duration(-5))) {
			return false, nil
		}
		return true, nil
	}
	return actual == expected, nil
}

func (m *LastTransitionTimeMatcher) FailureMessage(actual interface{}) (message string) {
	if m.expected.Time == time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC) {
		return fmt.Sprintf("Expected LastTransitionTime not be within the last 5 minutes, time now: %s, got: %s",
			time.Now(), actual)
	}
	return fmt.Sprintf("Expected\n\t%s\nto equal \n\t%s", actual, m.expected)
}

func (m *LastTransitionTimeMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	if m.expected.Time == time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC) {
		return fmt.Sprintf("Expected LastTransitionTime to be within the last 5 minutes, time now: %s, got: %s",
			time.Now(), actual)
	}
	return fmt.Sprintf("Expected\n\t%#v\nnot to match \n\t%#v", actual, m.expected)
}
