package matchers

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/test"
	v1 "k8s.io/api/core/v1"
)

type ContainerPortMatcher struct {
	expected interface{}
}

// EqualDiff is like Equal but gives cmp.Diff style output.
func IncludeContainerPort(expected interface{}) *ContainerPortMatcher {
	return &ContainerPortMatcher{
		expected: expected,
	}
}

func (m *ContainerPortMatcher) Match(actual interface{}) (success bool, err error) {
	exp, ok := m.expected.(v1.ContainerPort)
	if !ok {
		return false, fmt.Errorf("Matcher expects v1.ContainerPort")
	}
	actuals, ok := actual.([]v1.ContainerPort)
	if !ok {
		return false, fmt.Errorf("Matcher expects []v1.ContainerPort")
	}
	var found *v1.ContainerPort
	for i := range actuals {
		if actuals[i].Name == exp.Name {
			found = &actuals[i]
			break
		}
	}
	if found == nil {
		return false, nil
	}
	return test.JSONString(found) == test.JSONString(exp), nil
}

func (m *ContainerPortMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%s\nto contain \n\t%s", test.JSONString(actual), test.JSONString(m.expected))
}

func (m *ContainerPortMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto not contain \n\t%#v", actual, m.expected)
}

type ServicePortMatcher struct {
	expected interface{}
}

// EqualDiff is like Equal but gives cmp.Diff style output.
func IncludeServicePort(expected interface{}) *ServicePortMatcher {
	return &ServicePortMatcher{
		expected: expected,
	}
}

func (m *ServicePortMatcher) Match(actual interface{}) (success bool, err error) {
	exp, ok := m.expected.(v1.ServicePort)
	if !ok {
		return false, fmt.Errorf("Matcher expects v1.ServicePort")
	}
	actuals, ok := actual.([]v1.ServicePort)
	if !ok {
		return false, fmt.Errorf("Matcher expects []v1.ServicePort")
	}
	var found *v1.ServicePort
	for i := range actuals {
		if actuals[i].Name == exp.Name {
			found = &actuals[i]
			break
		}
	}
	if found == nil {
		return false, nil
	}
	return test.JSONString(found) == test.JSONString(exp), nil
}

func (m *ServicePortMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%s\nto contain \n\t%s", test.JSONString(actual), test.JSONString(m.expected))
}

func (m *ServicePortMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto not contain \n\t%#v", actual, m.expected)
}
