package observability_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	. "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/test"
	"strings"
)

var _ = Describe("helpers for output types", func() {

	Context("#SecretKeys", func() {

		It("should return an empty set of keys when authentication is not defined for an output", func() {
			for _, t := range obsv1.OutputTypes {

				outputType := strings.TrimPrefix("OutputType", string(t))
				outputType = strings.ToLower(outputType[0:1]) + outputType[1:]
				yaml := test.JSONLine(map[string]interface{}{
					"type":     t,
					outputType: map[string]interface{}{},
				})
				spec := &obsv1.OutputSpec{}
				test.MustUnmarshal(yaml, spec)
				Expect(SecretKeys(*spec)).To(BeEmpty())
			}
		})

	})
})
