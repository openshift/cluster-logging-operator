package drop

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"strings"
)

type Filter struct {
	tests []obs.DropTest
}

// NewFilter returns a drop filter
func NewFilter(dropTestsSpec []obs.DropTest) *Filter {
	return &Filter{dropTestsSpec}
}

func (f *Filter) VRL() (string, error) {
	vrlTests := []string{}
	for _, test := range f.tests {
		condList := []string{}
		for _, cond := range test.DropConditions {
			field := fmt.Sprintf("_.internal%s", cond.Field)
			if cond.Matches != "" {
				condList = append(condList, fmt.Sprintf(`match(to_string(%s) ?? "", r'%s')`, field, cond.Matches))
			} else {
				condList = append(condList, fmt.Sprintf(`!match(to_string(%s) ?? "", r'%s')`, field, cond.NotMatches))
			}
		}
		// Concatenate the conditions with ANDs and add Vector's error coalescing.
		// If any errors arise from the match such as, `cond.Field` not being a string or a field
		// is not present in the record, then it will automatically evaluate to false for the condition and specific test.
		vrlCondition := "(" + strings.Join(condList, " && ") + ")"
		vrlTests = append(vrlTests, vrlCondition)
	}

	// Vector's transform.Filter keeps logs that match the condition
	// Need `!()` to negate the whole expression if any condition evaluates to TRUE to drop logs
	return "!(" + strings.Join(vrlTests, " || ") + ")", nil
}
