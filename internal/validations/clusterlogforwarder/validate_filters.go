package clusterlogforwarder

import (
	"regexp"
	"strings"

	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/validations/clusterlogforwarder/conditions"
	"github.com/openshift/cluster-logging-operator/internal/validations/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Matches dot delimited paths with alphanumeric & `_`. Any other characters added in a segment will require quotes.
// Matches `.kubernetes.namespace_name` and `kubernetes."test-label/with slashes"`
var pathExpRegex = regexp.MustCompile(`^\.[a-zA-Z0-9_-]+(\.[a-zA-Z0-9_-]+|\."[^"]+")*$`)

// ValidateFilters validates the defined filters.
func ValidateFilters(clf loggingv1.ClusterLogForwarder, k8sClient client.Client, extras map[string]bool) (error, *loggingv1.ClusterLogForwarderStatus) {
	if len(clf.Spec.Filters) == 0 {
		return nil, nil
	}

	clf.Status.Filters = loggingv1.NamedConditions{}
	var err error
	for _, filterSpec := range clf.Spec.Filters {
		// Validate the drop filter
		if filterSpec.Type == loggingv1.FilterDrop {
			for _, dropFilter := range *filterSpec.DropTestsSpec {
				testErrors := []string{}
				for _, test := range dropFilter.DropConditions {
					// Validate field starts with a '.'
					if !strings.HasPrefix(test.Field, ".") {
						testErrors = append(testErrors, "field must start with a '.'")
					}
					// Validate field path
					if !pathExpRegex.MatchString(test.Field) {
						testErrors = append(testErrors, "field must be a valid dot delimited path expression (.kubernetes.container_name or .kubernetes.\"test-foo\")")
					}
					// Validate only one of matches/notMatches is defined
					if test.Matches != "" && test.NotMatches != "" {
						testErrors = append(testErrors, "only one of matches or notMatches can be defined at once")
					}
					// Validate provided regex
					if test.Matches != "" {
						_, err = regexp.Compile(test.Matches)
					} else if test.NotMatches != "" {
						_, err = regexp.Compile(test.Matches)
					}
					if err != nil {
						testErrors = append(testErrors, "matches/notMatches must be a valid regular expression.")
					}
				}
				if len(testErrors) != 0 {
					clf.Status.Filters.Set(filterSpec.Name, conditions.CondInvalid("%v", testErrors))
				}
			}
		}
	}
	if len(clf.Status.Filters) != 0 {
		return errors.NewValidationError("One or more errors are present in defined filters."), &clf.Status
	}
	return nil, nil
}
