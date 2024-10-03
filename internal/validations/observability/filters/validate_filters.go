package filters

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/set"
	"regexp"
	"strings"
)

var (
	// Matches dot delimited paths with alphanumeric & `_`. Any other characters added in a segment will require quotes.
	// Matches `.kubernetes.namespace_name` & `kubernetes."test-label/with slashes"` & `."@timestamp"`
	pathExpRegex = regexp.MustCompile(`^(\.[a-zA-Z0-9_]+|\."[^"]+")(\.[a-zA-Z0-9_]+|\."[^"]+")*$`)
)

func ValidateFilter(spec obs.FilterSpec) (condition metav1.Condition) {

	var results []string
	switch spec.Type {
	case obs.FilterTypeDrop:
		results = append(results, validateDropFilter(spec)...)
	case obs.FilterTypePrune:
		results = append(results, validatePruneFilter(spec)...)
	}
	condition = internalobs.NewConditionFromPrefix(obs.ConditionTypeValidFilterPrefix, spec.Name, true, obs.ReasonValidationSuccess, fmt.Sprintf("filter %q is valid", spec.Name))
	if len(results) > 0 {
		condition.Status = metav1.ConditionFalse
		condition.Reason = obs.ReasonValidationFailure
		condition.Message = strings.Join(results, ",")
	}
	return condition
}

// validateDropFilter validates each test and their associated conditions in a drop filter.
// It sets the filter status for the specific drop test index to better diagnose problems
func validateDropFilter(filterSpec obs.FilterSpec) (results []string) {
	if len(filterSpec.DropTestsSpec) == 0 {
		results = append(results, fmt.Sprintf("%q drop filter must have at least one test spec'd", filterSpec.Name))
	}

	var err error
	// Validate each test
	for i, dropTest := range filterSpec.DropTestsSpec {
		testErrors := []string{}
		// For each test, validate conditions
		for _, testCondition := range dropTest.DropConditions {
			if err := validateFieldPath(testCondition.Field); err != "" {
				testErrors = append(testErrors, err)
			}
			// Validate only one of matches/notMatches is defined
			if testCondition.Matches != "" && testCondition.NotMatches != "" {
				testErrors = append(testErrors, "only one of matches or notMatches can be defined at once")
			}
			// Validate provided regex
			if testCondition.Matches != "" {
				_, err = regexp.Compile(testCondition.Matches)
			} else if testCondition.NotMatches != "" {
				_, err = regexp.Compile(testCondition.Matches)
			}
			if err != nil {
				testErrors = append(testErrors, "matches/notMatches must be a valid regular expression.")
			}
		}
		if len(testErrors) != 0 {
			results = append(results, fmt.Sprintf("%s: test[%d] %v", filterSpec.Name, i, testErrors))
		}
	}
	return results
}

func validatePruneFilter(filterSpec obs.FilterSpec) (results []string) {
	if filterSpec.PruneFilterSpec == nil {
		results = append(results, fmt.Sprintf("%s prune filter must have one or both of `in`, `notIn`", filterSpec.Name))
		return results
	}
	errList := []string{}
	// Validate `in` paths
	if filterSpec.PruneFilterSpec.In != nil {
		for _, fieldPath := range filterSpec.PruneFilterSpec.In {
			if err := validateFieldPath(fieldPath); err != "" {
				errList = append(errList, err)
			}
		}
		// Ensure required fields are not in this list
		if valMsg := validateRequiredFields(filterSpec.PruneFilterSpec.In, "in"); valMsg != "" {
			errList = append(errList, valMsg)
		}
	}

	// Validate `notIn` paths
	if filterSpec.PruneFilterSpec.NotIn != nil {
		for _, fieldPath := range filterSpec.PruneFilterSpec.NotIn {
			if err := validateFieldPath(fieldPath); err != "" {
				errList = append(errList, err)
			}
		}
		// Ensure required fields are in this list
		if valMsg := validateRequiredFields(filterSpec.PruneFilterSpec.NotIn, "notIn"); valMsg != "" {
			errList = append(errList, valMsg)
		}
	}
	if len(errList) != 0 {
		results = append(results, fmt.Sprintf("%s: %v", filterSpec.Name, errList))
	}
	return results
}

// validateFieldPath validates a field path for correctness
func validateFieldPath(fieldPath obs.FieldPath) string {
	path := string(fieldPath)
	// Validate field starts with a '.'
	if !strings.HasPrefix(path, ".") {
		return fmt.Sprintf("%q must start with a '.'", fieldPath)
		// Validate field path
	} else if !pathExpRegex.MatchString(path) {
		return fmt.Sprintf("%q must be a valid dot delimited path expression (.kubernetes.container_name or .kubernetes.\"test-foo\")", fieldPath)
	}
	return ""
}

func validateRequiredFields(fieldList []obs.FieldPath, pruneType string) string {
	requiredFields := set.New[obs.FieldPath](".log_type", ".log_source", ".message")

	if pruneType == "in" {
		foundInList := []obs.FieldPath{}
		for _, field := range fieldList {
			if requiredFields.Has(field) {
				foundInList = append(foundInList, field)
			}
		}
		if len(foundInList) != 0 {
			return fmt.Sprintf("%q is/are required fields and must be removed from the `in` list.", foundInList)
		}
	} else {
		for _, field := range fieldList {
			if requiredFields.Has(field) {
				requiredFields.Delete(field)
			}
		}
		if requiredFields.Len() != 0 {
			return fmt.Sprintf("%q is/are required fields and must be included in the `notIn` list.", requiredFields.SortedList())
		}
	}

	return ""
}
