package matchers

import (
	"fmt"
	"reflect"

	log "github.com/ViaQ/logerr/v2/log/static"
	gformat "github.com/onsi/gomega/format"
	"github.com/onsi/gomega/matchers"
	"github.com/onsi/gomega/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/utils/toml"
	"github.com/openshift/cluster-logging-operator/test"
)

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
	rawExp, expCastable := expected.([]byte)
	if !expCastable {
		expectedType := reflect.TypeOf(expected)
		return false, fmt.Errorf("'expected' can not be converted to a string but is: %v", expectedType.Kind())
	}
	expConfig := &api.Config{}
	if err = toml.Unmarshal(string(rawExp), expConfig); err != nil {
		log.V(1).Info("expected config can not be unmarshalled as toml. trying another...", "err", err.Error(), "raw", string(rawExp))
		if err = test.Unmarshal(string(rawExp), expConfig); err != nil {
			return false, fmt.Errorf("'expected' can not be unmarshalled as yaml: %v", err)
		}
	}

	actualConfig, castable := m.actual.(*api.Config)
	if !castable {
		return false, fmt.Errorf("actual can not be converted to an api.Config: %v", m.actual)
	}
	m.matcher = matchers.MatchYAMLMatcher{
		YAMLToMatch: test.YAMLString(expConfig),
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
