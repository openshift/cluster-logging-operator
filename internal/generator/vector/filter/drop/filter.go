package drop

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"strings"
)

// MakeDropFilter returns a concatenated vrl string of tests and their conditions
func MakeDropFilter(dropTestsSpec []obs.DropTest) (string, error) {
	vrlTests := []string{}
	for _, test := range dropTestsSpec {
		condList := []string{}
		for _, cond := range test.DropConditions {
			if cond.Matches != "" {
				condList = append(condList, fmt.Sprintf(`match(%s, r'%s')`, cond.Field, cond.Matches))
			} else {
				condList = append(condList, fmt.Sprintf(`!match(%s, r'%s')`, cond.Field, cond.NotMatches))
			}
		}
		// Concatenate the conditions with ANDs and add Vector's error coalescing.
		// If any errors arise from the match such as, `cond.Field` not being a string or a field
		// is not present in the record, then it will automatically evaluate to false for the condition and specific test.
		vrlCondition := "(" + strings.Join(condList, " && ") + " ?? false)"
		vrlTests = append(vrlTests, vrlCondition)
	}

	// Vector's transform.Filter keeps logs that match the condition
	// Need `!()` to negate the whole expression if any condition evaluates to TRUE to drop logs
	return "!(" + strings.Join(vrlTests, " || ") + ")", nil
}
