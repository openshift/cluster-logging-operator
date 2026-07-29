package drop

import (
	"fmt"
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/transforms"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

type Filter struct {
	tests []obs.DropTest
}

// NewFilter returns a drop filter
func NewFilter(dropTestsSpec []obs.DropTest) *Filter {
	return &Filter{dropTestsSpec}
}

func New(spec []obs.DropTest, inputs ...string) types.Transform {
	pf := NewFilter(spec)
	vrl, err := pf.VRL()
	if err != nil {
		log.Error(err, "bad filter", "dropSpec", spec)
		return nil
	}
	return transforms.NewFilter(vrl, inputs...)
}

func buildMatchCondition(field, pattern string, negate bool) (string, error) {
	if strings.ContainsAny(pattern, "'\n\r") {
		return "", fmt.Errorf("match pattern must not contain single quotes, newlines, or carriage returns: %q", pattern)
	}
	prefix := ""
	if negate {
		prefix = "!"
	}
	return fmt.Sprintf(`%smatch(to_string(%s) ?? "", r'%s')`, prefix, field, pattern), nil
}

func (f *Filter) VRL() (string, error) {
	vrlTests := []string{}
	for _, test := range f.tests {
		condList := []string{}
		for _, cond := range test.DropConditions {
			field := fmt.Sprintf("._internal%s", cond.Field)
			var matchExpr string
			var err error
			if cond.Matches != "" {
				matchExpr, err = buildMatchCondition(field, cond.Matches, false)
			} else {
				matchExpr, err = buildMatchCondition(field, cond.NotMatches, true)
			}
			if err != nil {
				return "", err
			}
			condList = append(condList, matchExpr)
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
