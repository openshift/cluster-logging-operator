package inputs

import (
	. "github.com/onsi/ginkgo"

	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

var _ = Describe("#ConvertInfrastructureInputs", func() {
	It("should map logging.Infrastructure to observability.Infrastructure", func() {
		loggingInfra := logging.Infrastructure{
			Sources: []string{"foo", "bar", "baz"},
		}

		expObsInfra := &obs.Infrastructure{
			Sources: []obs.InfrastructureSource{"foo", "bar", "baz"},
		}

		Expect(MapInfrastructureInput(&loggingInfra)).To(Equal(expObsInfra))
	})
})
