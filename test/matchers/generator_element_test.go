package matchers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	"strings"
)

const FakeConfiguration = "fake-configuration"

type FakeElement struct{}

func (f FakeElement) Name() string {
	return "fake-element"
}
func (f FakeElement) Template() string {
	return `{{define "` + f.Name() + `" -}}
fake-configuration
{{- end}}
`
}

var _ = Describe("Generator Element matcher", func() {

	Context("when verifying an array of Sections", func() {

		It("should match when elements generate the same config", func() {
			Expect(FakeConfiguration).To(EqualConfigFrom([]framework.Section{
				{Elements: []framework.Element{&FakeElement{}}},
			}))
		})

	})
	Context("when verifying a slice of Elements", func() {
		It("should match the generated config from a slice of pointer elements", func() {
			exp := []string{FakeConfiguration, FakeConfiguration, FakeConfiguration}
			Expect(strings.Join(exp, "\n")).To(EqualConfigFrom([]framework.Element{&FakeElement{}, &FakeElement{}, &FakeElement{}}))
		})
		It("should match the generated config from a slice of struct elements", func() {
			exp := []string{FakeConfiguration, FakeConfiguration, FakeConfiguration}
			Expect(strings.Join(exp, "\n")).To(EqualConfigFrom([]framework.Element{FakeElement{}, FakeElement{}, FakeElement{}}))
		})
	})
	Context("when verifying an Element", func() {
		It("should match when a pointer element generates the same config", func() {
			Expect(FakeConfiguration).To(EqualConfigFrom(&FakeElement{}))
		})
		It("should match when a struct element generates the same config", func() {
			Expect(FakeConfiguration).To(EqualConfigFrom(FakeElement{}))
		})
	})

})
