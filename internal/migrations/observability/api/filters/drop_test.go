package filters

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

var _ = Describe("#MapDropFilter", func() {
	It("should map logging.DropTest to observability.DropTest", func() {
		loggingDropTest := []logging.DropTest{
			{
				DropConditions: []logging.DropCondition{
					{
						Field:   ".foo",
						Matches: "bar",
					},
					{
						Field:      ".foo2",
						NotMatches: "baz",
					},
				},
			},
			{
				DropConditions: []logging.DropCondition{
					{
						Field:   ".test",
						Matches: "foo",
					},
					{
						Field:   ".laz",
						Matches: "paz",
					},
				},
			},
		}
		expObsDropTest := []obs.DropTest{
			{
				DropConditions: []obs.DropCondition{
					{
						Field:   ".foo",
						Matches: "bar",
					},
					{
						Field:      ".foo2",
						NotMatches: "baz",
					},
				},
			},
			{
				DropConditions: []obs.DropCondition{
					{
						Field:   ".test",
						Matches: "foo",
					},
					{
						Field:   ".laz",
						Matches: "paz",
					},
				},
			},
		}
		Expect(MapDropFilter(loggingDropTest)).To(Equal(expObsDropTest))
	})
})
