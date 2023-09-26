package filter

import (
	"fmt"

	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/apiaudit"
)

// RemapVRL returns a VRL expression to add to the remap program of a pipeline containing this filter.
// Can be used for validation as well as execution of the filter.
func RemapVRL(filterSpec *loggingv1.FilterSpec) (vrl string, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("filter %v: %w", filterSpec.Name, err)
		}
	}()
	switch filterSpec.Type {

	case loggingv1.FilterKubeAPIAudit:
		return apiaudit.PolicyToVRL(filterSpec.KubeAPIAudit)

	case "":
		return "", fmt.Errorf("missing filter type")
	default:
		return "", fmt.Errorf("unknown filter type: %v", filterSpec.Type)
	}
}
