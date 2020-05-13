package helpers

import (
	"fmt"
	"strings"

	yaml "gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func NormalizeYaml(doc string) string {
	doc = strings.TrimSpace(doc)
	data := &map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(doc), data); err != nil {
		Fail(fmt.Sprintf("Unable to normalize document '%s': %v", doc, err))
	}
	response, err := yaml.Marshal(data)
	if err != nil {
		Fail(fmt.Sprintf("Unable to normalize document '%s': %v", doc, err))
	}
	return string(response)
}

type YamlExpectation struct {
	actual string
}

func ExpectYaml(doc string) *YamlExpectation {
	return &YamlExpectation{actual: doc}
}

func (exp *YamlExpectation) ToEqual(doc string) {
	actual := NormalizeYaml(exp.actual)
	expected := NormalizeYaml(doc)
	if actual != expected {
		fmt.Printf("Actual>:\n%s<\n", actual)
		fmt.Printf("Expected>:\n%s\n<", expected)
		Expect(actual).To(Equal(expected))
	}
}
