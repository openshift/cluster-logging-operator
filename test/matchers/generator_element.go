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
	expConfig := &api.Config{}
	if err = toml.Unmarshal(expString, expConfig); err != nil {
		log.V(1).Info("expected config can not be unmarshalled as toml. trying another...", "err", err.Error(), "raw", expString)
		if err = test.Unmarshal(expString, expConfig); err != nil {
			return false, fmt.Errorf("'expected' can not be unmarshalled as yaml: %v", err)
		}
	}

	actualConfig, castable := m.actual.(*api.Config)
	if !castable {
		return false, fmt.Errorf("actual can not be converted to an api.Config: %v", m.actual)
	}

	// Round-trip the actual config through TOML marshal/unmarshal to verify
	// that serialization preserves all fields correctly. This catches bugs
	// where struct field ordering causes fields to be nested under the wrong
	// TOML table (e.g. LOG-9444: compression inside [auth]).
	tomlStr, err := toml.Marshal(actualConfig)
	if err != nil {
		return false, fmt.Errorf("actual config can not be marshalled to TOML: %v", err)
	}
	roundTripped := &api.Config{}
	if err = toml.Unmarshal(tomlStr, roundTripped); err != nil {
		return false, fmt.Errorf("actual config TOML output can not be unmarshalled back: %v", err)
	}

	m.matcher = matchers.MatchYAMLMatcher{
		YAMLToMatch: test.YAMLString(expConfig),
	}
	m.actual = test.YAMLString(roundTripped)
	return m.matcher.Match(m.actual)
}

func (m *GeneratorElementMatcher) FailureMessage(_ interface{}) (message string) {
	return m.matcher.FailureMessage(m.actual)
}

func (m *GeneratorElementMatcher) NegatedFailureMessage(_ interface{}) (message string) {
	return m.matcher.NegatedFailureMessage(m.actual)
}
