package matchers

import (
	"fmt"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/openshift/cluster-logging-operator/test"
)

type MetricsEndpointMatcher struct {
	expected interface{}
}

// EqualDiff is like Equal but gives cmp.Diff style output.
func IncludeMetricsEndpoint(expected interface{}) *MetricsEndpointMatcher {
	return &MetricsEndpointMatcher{
		expected: expected,
	}
}

func (m *MetricsEndpointMatcher) Match(actual interface{}) (success bool, err error) {
	exp, ok := m.expected.(monitoringv1.Endpoint)
	if !ok {
		return false, fmt.Errorf("Matcher expects monitoringv1.Endpoint")
	}
	actuals, ok := actual.([]monitoringv1.Endpoint)
	if !ok {
		return false, fmt.Errorf("Matcher expects []monitoringv1.Endpoint")
	}
	var found *monitoringv1.Endpoint
	for i := range actuals {
		if actuals[i].Port == exp.Port {
			found = &actuals[i]
			break
		}
	}
	if found == nil {
		return false, nil
	}
	return test.JSONString(found) == test.JSONString(exp), nil
}

func (m *MetricsEndpointMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%s\nto contain \n\t%s", test.JSONString(actual), test.JSONString(m.expected))
}

func (m *MetricsEndpointMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected\n\t%#v\nto not contain \n\t%#v", actual, m.expected)
}
