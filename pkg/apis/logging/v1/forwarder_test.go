package v1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
)

var _ = Describe("ClusterLogForwarderSpec", func() {

	It("calculates routes", func() {
		spec := ClusterLogForwarderSpec{
			Pipelines: []PipelineSpec{
				{
					InputRefs:  []string{InputNameApplication},
					OutputRefs: []string{"X", "Y"},
				},
				{
					InputRefs:  []string{InputNameInfrastructure, InputNameAudit},
					OutputRefs: []string{"Y", "Z"},
				},
				{
					InputRefs:  []string{InputNameAudit},
					OutputRefs: []string{"X", "Z"},
				},
			},
		}
		routes := NewRoutes(spec.Pipelines)
		Expect(routes.ByInput).To(Equal(RouteMap{
			InputNameAudit:          {"X": {}, "Y": {}, "Z": {}},
			InputNameApplication:    {"X": {}, "Y": {}},
			InputNameInfrastructure: {"Y": {}, "Z": {}},
		}))
		Expect(routes.ByOutput).To(Equal(RouteMap{
			"X": {InputNameApplication: {}, InputNameAudit: {}},
			"Y": {InputNameApplication: {}, InputNameInfrastructure: {}, InputNameAudit: {}},
			"Z": {InputNameInfrastructure: {}, InputNameAudit: {}},
		}))
	})
})

var _ = Describe("inputs", func() {
	It("has built-in input types", func() {
		Expect(ReservedInputNames.List()).To(ConsistOf("infrastructure", "application", "audit"))
	})
})
