package matchers

import (
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
	"github.com/onsi/gomega/types"
)

type TomlMatcher struct {
	actualToml string
	expToml    interface{}
	diff       string
	err        error
}

func MatchToml(expected interface{}) types.GomegaMatcher {
	return &TomlMatcher{
		expToml: expected,
	}
}

func (m *TomlMatcher) Match(actual interface{}) (bool, error) {
	m.actualToml, m.err = generateConf(actual)
	if m.err != nil {
		return false, m.err
	}
	actualConfig := strings.Join(normalize(m.actualToml, true), "\n")
	expConfig := strings.Join(normalize(m.expToml.(string), true), "\n")

	m.diff = cmp.Diff(expConfig, actualConfig)
	return m.diff == "", nil
}

func (m *TomlMatcher) FailureMessage(expected interface{}) (message string) {
	if m.err != nil {
		return fmt.Sprintf("Error generating 'expected' conf: %v", m.err)
	}
	return fmt.Sprintf("Expected element to produce a config from 'elements'\nexpected:\n>>>\n%s\n\n<<<\n\n\n\nactual\n>>>:\n\n%s\n\n<<<\ndiff: %s\n", expected, m.actualToml, m.diff)
}

func (m *TomlMatcher) NegatedFailureMessage(expected interface{}) (message string) {
	if m.err != nil {
		return fmt.Sprintf("Error generating 'expected' conf: %v", m.err)
	}
	return fmt.Sprintf("Expected element to not produce a config from 'elements'\nexpected:\n%s\n\nactual:\n\n%s\n\ndiff: %s\n", expected, m.actualToml, m.diff)
}
