package filters

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

func MapDropFilter(loggingDropTest []logging.DropTest) []obs.DropTest {
	obsDropTests := []obs.DropTest{}
	for _, test := range loggingDropTest {
		obsDropConditions := []obs.DropCondition{}

		for _, cond := range test.DropConditions {
			obsCond := obs.DropCondition{
				Field:      obs.FieldPath(cond.Field),
				Matches:    cond.Matches,
				NotMatches: cond.NotMatches,
			}

			obsDropConditions = append(obsDropConditions, obsCond)
		}

		obsDropTests = append(obsDropTests, obs.DropTest{
			DropConditions: obsDropConditions,
		})
	}

	return obsDropTests
}
