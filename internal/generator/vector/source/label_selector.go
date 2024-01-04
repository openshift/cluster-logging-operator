package source

import (
	"fmt"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
	"strings"
)

func LabelSelectorFrom(selector *logging.LabelSelector) string {
	if selector == nil {
		return ""
	}
	results := matchLabels(selector.MatchLabels)
	results = append(results, matchExpressions(selector.MatchExpressions)...)
	return strings.Join(results, ",")
}

func matchExpressions(expressions []metav1.LabelSelectorRequirement) (results []string) {
	for _, r := range expressions {
		switch r.Operator {
		case metav1.LabelSelectorOpExists:
			results = append(results, r.Key)
		case metav1.LabelSelectorOpDoesNotExist:
			results = append(results, fmt.Sprintf("!%s", r.Key))
		case metav1.LabelSelectorOpIn:
			sort.Strings(r.Values)
			results = append(results, fmt.Sprintf("%s in (%s)", r.Key, strings.Join(r.Values, ",")))
		case metav1.LabelSelectorOpNotIn:
			results = append(results, fmt.Sprintf("%s notin (%s)", r.Key, strings.Join(r.Values, ",")))
		}
	}
	sort.Strings(results)
	return results
}

func matchLabels(matchLabels map[string]string) (results []string) {
	for k, v := range matchLabels {
		if v == "" {
			results = append(results, k)
		} else {
			results = append(results, fmt.Sprintf("%s=%s", k, v))
		}
	}
	sort.Strings(results)
	return results
}
