package v1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

var _ = Describe("ClusterLogForwarderSpec", func() {

	It("calculates routes", func() {
		spec := v1.ClusterLogForwarderSpec{
			Pipelines: []v1.PipelineSpec{
				{
					InputRefs:  []string{v1.InputNameApplication},
					OutputRefs: []string{"X", "Y"},
				},
				{
					InputRefs:  []string{v1.InputNameInfrastructure, v1.InputNameAudit},
					OutputRefs: []string{"Y", "Z"},
				},
				{
					InputRefs:  []string{v1.InputNameAudit},
					OutputRefs: []string{"X", "Z"},
				},
			},
		}
		routes := v1.NewRoutes(spec.Pipelines)
		Expect(routes.ByInput).To(Equal(v1.RouteMap{
			v1.InputNameAudit:          {"X": {}, "Y": {}, "Z": {}},
			v1.InputNameApplication:    {"X": {}, "Y": {}},
			v1.InputNameInfrastructure: {"Y": {}, "Z": {}},
		}))
		Expect(routes.ByOutput).To(Equal(v1.RouteMap{
			"X": {v1.InputNameApplication: {}, v1.InputNameAudit: {}},
			"Y": {v1.InputNameApplication: {}, v1.InputNameInfrastructure: {}, v1.InputNameAudit: {}},
			"Z": {v1.InputNameInfrastructure: {}, v1.InputNameAudit: {}},
		}))
	})
})

var _ = Describe("inputs", func() {
	It("has built-in input types", func() {
		Expect(v1.ReservedInputNames.List()).To(ConsistOf("infrastructure", "application", "audit"))
	})
})
