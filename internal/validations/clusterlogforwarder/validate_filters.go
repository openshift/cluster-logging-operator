package clusterlogforwarder

import (
	"fmt"
	"regexp"
	"strings"

	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	"github.com/openshift/cluster-logging-operator/internal/validations/clusterlogforwarder/conditions"
	"github.com/openshift/cluster-logging-operator/internal/validations/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// Matches dot delimited paths with alphanumeric & `_`. Any other characters added in a segment will require quotes.
	// Matches `.kubernetes.namespace_name` & `kubernetes."test-label/with slashes"` & `."@timestamp"`
	pathExpRegex = regexp.MustCompile(`^(\.[a-zA-Z0-9_]+|\."[^"]+")(\.[a-zA-Z0-9_]+|\."[^"]+")*$`)
)

// ValidateFilters validates the defined filters.
func ValidateFilters(clf loggingv1.ClusterLogForwarder, k8sClient client.Client, extras map[string]bool) (error, *loggingv1.ClusterLogForwarderStatus) {
	if len(clf.Spec.Filters) == 0 {
		return nil, nil
	}

	clf.Status.Filters = loggingv1.NamedConditions{}
	for _, filterSpec := range clf.Spec.Filters {
		switch filterSpec.Type {
		case loggingv1.FilterDrop:
			validateDropFilter(&filterSpec, &clf.Status.Filters)
		case loggingv1.FilterPrune:
			validatePruneFilter(&filterSpec, &clf.Status.Filters)
		}
	}
	if len(clf.Status.Filters) != 0 {
		return errors.NewValidationError("One or more errors are present in defined filters."), &clf.Status
	}
	return nil, nil
}

// validateDropFilter validates each test and their associated conditions in a drop filter.
// It sets the filter status for the specific drop test index to better diagnose problems
func validateDropFilter(filterSpec *loggingv1.FilterSpec, clfStatus *loggingv1.NamedConditions) {
	var err error
	// Validate each test
	for i, dropTest := range *filterSpec.DropTestsSpec {
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
			clfStatus.Set(fmt.Sprintf("%s: test[%d]", filterSpec.Name, i), conditions.CondInvalid("%v", testErrors))
		}
	}
}

func validatePruneFilter(filterSpec *loggingv1.FilterSpec, clfStatus *loggingv1.NamedConditions) {
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
		clfStatus.Set(filterSpec.Name, conditions.CondInvalid("%v", errList))
	}
}

// validateFieldPath validates a field path for correctness
func validateFieldPath(fieldPath string) string {
	// Validate field starts with a '.'
	if !strings.HasPrefix(fieldPath, ".") {
		return fmt.Sprintf("%q must start with a '.'", fieldPath)
		// Validate field path
	} else if !pathExpRegex.MatchString(fieldPath) {
		return fmt.Sprintf("%q must be a valid dot delimited path expression (.kubernetes.container_name or .kubernetes.\"test-foo\")", fieldPath)
	}
	return ""
}

func validateRequiredFields(fieldList []string, pruneType string) string {
	requiredFields := sets.NewString(".log_type", ".message")

	if pruneType == "in" {
		foundInList := []string{}
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
				requiredFields.Remove(field)
			}
		}
		if requiredFields.Len() != 0 {
			return fmt.Sprintf("%q is/are required fields and must be included in the `notIn` list.", requiredFields.List())
		}
	}

	return ""
}
