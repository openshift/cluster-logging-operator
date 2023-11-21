package matchers

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/onsi/gomega/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"reflect"
	"strings"
)

type GeneratorElementMatcher struct {
	expected interface{}
	diff     string
	err      error
}

func EqualConfigFrom(expected interface{}) types.GomegaMatcher {
	return &GeneratorElementMatcher{
		expected: expected,
	}
}

func (m *GeneratorElementMatcher) Match(actual interface{}) (bool, error) {
	if actType := reflect.TypeOf(actual); actType.Kind() != reflect.String {
		return false, fmt.Errorf("The 'actual' type is expected to be a string but is: %s", actType.Name())
	}
	conf, err := generateConf(m.expected)
	m.err = err
	if err != nil {
		return false, err
	}
	nactual := strings.Join(normalize(actual.(string), true), `\n`)
	nexpected := strings.Join(normalize(conf, true), `\n`)
	m.diff = cmp.Diff(nexpected, nactual)
	return m.diff == "", nil
}

func (m *GeneratorElementMatcher) FailureMessage(actual interface{}) (message string) {
	if m.err != nil {
		return fmt.Sprintf("Error generating 'expected' conf: %v", m.err)
	}
	return fmt.Sprintf("Expected element to produce a config from 'elements' but got this diff: %s", m.diff)
}

func (m *GeneratorElementMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	if m.err != nil {
		return fmt.Sprintf("Error generating 'expected' conf: %v", m.err)
	}
	return fmt.Sprintf("Expected element to not produce config from 'elements' but got this diff: %s", m.diff)
}

func generateConf(expected interface{}) (string, error) {
	var els []framework.Element
	expType := reflect.TypeOf(expected)
	switch {
	case expType.Kind() == reflect.Slice && expType.Elem().Name() == "Section":
		sections := expected.([]framework.Section)
		for _, v := range sections {
			els = append(els, v.Elements...)
		}
	case expType.Kind() == reflect.Slice && expType.Elem().Name() == "Element":
		elements := expected.([]framework.Element)
		els = elements
	case expType.Implements(generatorElementType):
		if el, ok := expected.(framework.Element); ok {
			els = []framework.Element{el}
		} else {
			return "", fmt.Errorf("Matcher unable to cast 'expected' type %q to a generator.Element", expType.Name())
		}
	default:
		return "", fmt.Errorf("Matcher does not support 'expected' kind %q or element type: %q", expType.Kind(), expType.Name())
	}
	g := framework.MakeGenerator()
	return g.GenerateConf(els...)
}

var (
	generatorElementType = reflect.TypeOf((*framework.Element)(nil)).Elem()
)
