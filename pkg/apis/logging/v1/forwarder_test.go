package v1_test

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1/outputs"
)

var _ = Describe("ClusterLogForwarderSpec", func() {

	It("calculates routes", func() {
		spec := ClusterLogForwarderSpec{
			Pipelines: []PipelineSpec{
				{
					InputRefs:  []string{"Application"},
					OutputRefs: []string{"X", "Y"},
				},
				{
					InputRefs:  []string{"Infrastructure", "Audit"},
					OutputRefs: []string{"Y", "Z"},
				},
				{
					InputRefs:  []string{"Audit"},
					OutputRefs: []string{"X", "Z"},
				},
			},
		}
		routes := NewRoutes(spec.Pipelines)
		Expect(routes.ByInput).To(Equal(RouteMap{
			"Audit":          {"X": {}, "Y": {}, "Z": {}},
			"Application":    {"X": {}, "Y": {}},
			"Infrastructure": {"Y": {}, "Z": {}},
		}))
		Expect(routes.ByOutput).To(Equal(RouteMap{
			"X": {"Application": {}, "Audit": {}},
			"Y": {"Application": {}, "Infrastructure": {}, "Audit": {}},
			"Z": {"Infrastructure": {}, "Audit": {}},
		}))
	})
})

var _ = Describe("outputs", func() {
	It("finds output type specs by name", func() {
		Expect(FindOutputType("syslog")).To(Equal(reflect.TypeOf(outputs.Syslog{})))
		Expect(FindOutputType("fluentForward")).To(Equal(reflect.TypeOf(outputs.FluentForward{})))
		Expect(FindOutputType("elasticsearch")).To(Equal(reflect.TypeOf(outputs.ElasticSearch{})))
		Expect(FindOutputType("nosuch")).To(BeNil())
		Expect(FindOutputType("")).To(BeNil())
		Expect(FindOutputType("type")).To(BeNil())
	})
	It("gets JSON field name for output type specs", func() {
		Expect(OutputTypeName(outputs.Syslog{})).To(Equal("syslog"))
		Expect(OutputTypeName(reflect.TypeOf(outputs.Syslog{}))).To(Equal("syslog"))
		Expect(OutputTypeName(outputs.ElasticSearch{})).To(Equal("elasticsearch"))
		Expect(OutputTypeName(outputs.FluentForward{})).To(Equal("fluentForward"))
		Expect(OutputTypeName(reflect.TypeOf(OutputSpec{}))).To(Equal(""))
	})
})

var _ = Describe("inputs", func() {
	It("has built-in input types", func() {
		Expect(BuiltInInputs.List()).To(ConsistOf("Infrastructure", "Application", "Audit"))
	})
})
