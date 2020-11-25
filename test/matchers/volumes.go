package matchers

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/test"
	v1 "k8s.io/api/core/v1"
)

type VolumeMountMatcher struct {
	expected interface{}
}

// EqualDiff is like Equal but gives cmp.Diff style output.
func IncludeVolumeMount(expected interface{}) *VolumeMountMatcher {
	return &VolumeMountMatcher{
		expected: expected,
	}
}

func (m *VolumeMountMatcher) Match(actual interface{}) (success bool, err error) {
	exp, ok := m.expected.(v1.VolumeMount)
	if !ok {
		return false, fmt.Errorf("Matcher expects v1.VolumeMount")
	}
	actualMounts, ok := actual.([]v1.VolumeMount)
	if !ok {
		return false, fmt.Errorf("Matcher expects []v1.VolumeMount")
	}
	var found *v1.VolumeMount
	for i := range actualMounts {
		if actualMounts[i].Name == exp.Name {
			found = &actualMounts[i]
			break
		}
	}
	if found == nil {
		return false, nil
	}
	return test.JSONString(found) == test.JSONString(exp), nil
}

func (m *VolumeMountMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%s\nto contain \n\t%s", test.JSONString(actual), test.JSONString(m.expected))
}

func (m *VolumeMountMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto not contain \n\t%#v", actual, m.expected)
}
