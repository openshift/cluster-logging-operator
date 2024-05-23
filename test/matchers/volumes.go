package matchers

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/test"
	v1 "k8s.io/api/core/v1"
)

type VolumeMatcher struct {
	expected interface{}
}

func IncludeVolume(expected interface{}) *VolumeMatcher {
	return &VolumeMatcher{
		expected: expected,
	}
}

func (m *VolumeMatcher) Match(actual interface{}) (success bool, err error) {
	expVolume, ok := m.expected.(v1.Volume)
	if !ok {
		return false, fmt.Errorf("Matcher expects v1.Volume")
	}
	actualVolumes, ok := actual.([]v1.Volume)
	if !ok {
		return false, fmt.Errorf("Matcher expects []v1.Volume")
	}
	var found *v1.Volume
	for i := range actualVolumes {
		if actualVolumes[i].Name == expVolume.Name {
			found = &actualVolumes[i]
			break
		}
	}
	if found == nil {
		return false, nil
	}
	return test.JSONString(found) == test.JSONString(expVolume), nil
}

func (m *VolumeMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%s\nto contain \n\t%s", test.JSONString(actual), test.JSONString(m.expected))
}

func (m *VolumeMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto not contain \n\t%#v", actual, m.expected)
}

type VolumeMountMatcher struct {
	expected interface{}
}

func IncludeVolumeMount(expected interface{}) *VolumeMountMatcher {
	return &VolumeMountMatcher{
		expected: expected,
	}
}

func (m *VolumeMountMatcher) Match(actual interface{}) (success bool, err error) {
	expVolume, ok := m.expected.(v1.VolumeMount)
	if !ok {
		return false, fmt.Errorf("Matcher expects v1.VolumeMount")
	}
	actualVolumes, ok := actual.([]v1.VolumeMount)
	if !ok {
		return false, fmt.Errorf("Matcher expects []v1.VolumeMount")
	}
	var found *v1.VolumeMount
	for i := range actualVolumes {
		if actualVolumes[i].Name == expVolume.Name {
			found = &actualVolumes[i]
			break
		}
	}
	if found == nil {
		return false, nil
	}
	return test.JSONString(found) == test.JSONString(expVolume), nil
}

func (m *VolumeMountMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%s\nto contain \n\t%s", test.JSONString(actual), test.JSONString(m.expected))
}

func (m *VolumeMountMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto not contain \n\t%#v", actual, m.expected)
}
