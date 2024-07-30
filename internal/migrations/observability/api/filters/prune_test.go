package filters

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

var _ = Describe("#MapPruneFilter", func() {
	It("should map logging.PruneFilterSpec to observability.PruneFilterSpec", func() {
		loggingPruneSpec := &logging.PruneFilterSpec{
			In:    []string{"foo", "bar", "baz"},
			NotIn: []string{"test", "something", "other"},
		}

		expObsPruneSpec := &obs.PruneFilterSpec{
			In:    []obs.FieldPath{"foo", "bar", "baz"},
			NotIn: []obs.FieldPath{"test", "something", "other"},
		}

		Expect(MapPruneFilter(*loggingPruneSpec)).To(Equal(expObsPruneSpec))
	})
})
