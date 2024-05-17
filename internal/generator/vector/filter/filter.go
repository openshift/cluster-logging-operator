package filter

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"

	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/drop"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/openshift"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/prune"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/apiaudit"
)

// RemapVRL returns a VRL expression to add to the remap program of a pipeline containing this filter.
// Can be used for validation as well as execution of the filter.
func RemapVRL(filterSpec *obs.FilterSpec) (vrl string, err error) {
	iSpec := &InternalFilterSpec{FilterSpec: filterSpec}
	return VRLFrom(iSpec)
}

func VRLFrom(filterSpec *InternalFilterSpec) (vrl string, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("filter %v: %w", filterSpec.Name, err)
		}
	}()
	switch filterSpec.Type {
	case obs.FilterTypeOpenshiftLabels:
		return openshift.NewLabels(filterSpec.OpenShiftLabels)
	case obs.FilterTypeDrop:
		return drop.MakeDropFilter(filterSpec.DropTestsSpec)
	case obs.FilterTypePrune:
		return prune.MakePruneFilter(filterSpec.PruneFilterSpec)
	case obs.FilterTypeKubeAPIAudit:
		return apiaudit.PolicyToVRL(filterSpec.KubeAPIAudit)
	case obs.FilterTypeParse:
		return openshift.NewParseJSON()
	case "":
		return "", fmt.Errorf("missing filter type")
	default:
		return "", fmt.Errorf("unknown filter type: %v", filterSpec.Type)
	}
}

// InternalFilterSpec is a wrapper to allow separation of public and internal filters
type InternalFilterSpec struct {
	*obs.FilterSpec
	SuppliesTransform bool

	//TranformFactory takes an id, inputs and returns an Element
	TranformFactory func(id string, inputs ...string) framework.Element
}

func NewInternalFilterMap(filters map[string]*obs.FilterSpec) map[string]*InternalFilterSpec {
	internalFilters := map[string]*InternalFilterSpec{}
	for _, f := range filters {
		internalFilters[f.Name] = &InternalFilterSpec{FilterSpec: f}
	}
	return internalFilters
}
