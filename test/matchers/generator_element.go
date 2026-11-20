package matchers

import (
	"fmt"
	"reflect"

	gformat "github.com/onsi/gomega/format"
	"github.com/onsi/gomega/matchers"
	"github.com/onsi/gomega/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/test"
)

// GeneratorElementMatcher asserts a generator element marshals to the expected YAML.
//
// It is deliberately marshal-only: the element is serialized through sigs.k8s.io/yaml
// (encoding/json semantics, honoring `json:` tags) and compared against the expected
// YAML. The historical TOML round-trip guard — which existed because go-toml's
// tree-based marshaling could silently misplace fields (LOG-9444) — is unnecessary
// here: encoding/json always nests struct fields under their parent key, so that
// failure mode cannot occur.
type GeneratorElementMatcher struct {
	actual       interface{}
	matcher      matchers.MatchYAMLMatcher
	truncateDiff bool
	maxLength    int
}

func EqualConfigFrom(actual interface{}) types.GomegaMatcher {

	return &GeneratorElementMatcher{
		actual: actual,
	}
}

func (m *GeneratorElementMatcher) Match(expected interface{}) (_ bool, err error) {
	defer func() {
		gformat.TruncatedDiff = m.truncateDiff
		gformat.MaxLength = m.maxLength
	}()
	gformat.TruncatedDiff = true
	gformat.MaxLength = 0
	var expString string
	expectedType := reflect.TypeOf(expected)
	if expectedType.Kind() == reflect.String {
		expString = expected.(string)
	} else {
		rawExp, expCastable := expected.([]byte)
		if !expCastable {
			return false, fmt.Errorf("'expected' can not be converted to a string but is: %v", expectedType.Kind())
		}
		expString = string(rawExp)
	}

	actualConfig, castable := m.actual.(*api.Config)
	if !castable {
		return false, fmt.Errorf("actual can not be converted to an api.Config: %v", m.actual)
	}

	m.matcher = matchers.MatchYAMLMatcher{
		YAMLToMatch: expString,
	}
	m.actual = test.YAMLString(actualConfig)
	return m.matcher.Match(m.actual)
}

func (m *GeneratorElementMatcher) FailureMessage(_ interface{}) (message string) {
	return m.matcher.FailureMessage(m.actual)
}

func (m *GeneratorElementMatcher) NegatedFailureMessage(_ interface{}) (message string) {
	return m.matcher.NegatedFailureMessage(m.actual)
}
