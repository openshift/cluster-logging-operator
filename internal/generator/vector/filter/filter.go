package filter

import (
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/multilineexception"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/parse"

	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/drop"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/openshift"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/prune"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/apiaudit"
)

// InternalFilterSpec is a wrapper to allow separation of public and internal filters
type InternalFilterSpec struct {
	*obs.FilterSpec

	// SuppliesTransform identifies if the filter will provide its own transformation or rely upon the framework
	// to generate a remap transformation using the VRL it provides
	SuppliesTransform bool

	// RemapFilter is a filter that uses a remap transformation
	RemapFilter RemapFilter

	//TranformFactory takes an id, inputs and returns an Element
	TranformFactory func(id string, inputs ...string) framework.Element
}

// RemapFilter is a remap transform that provides VRL script
type RemapFilter interface {
	// VRL returns the VRL for filter or error
	VRL() (string, error)
}

func NewInternalFilterMap(filters map[string]*obs.FilterSpec) map[string]*InternalFilterSpec {
	internalFilters := map[string]*InternalFilterSpec{}
	for _, f := range filters {
		internalFilter := &InternalFilterSpec{FilterSpec: f}
		switch f.Type {
		case obs.FilterTypeOpenshiftLabels:
			internalFilter.RemapFilter = openshift.NewLabelsFilter(f.OpenshiftLabels)
		case obs.FilterTypeDrop:
			internalFilter.RemapFilter = drop.NewFilter(f.DropTestsSpec)
		case obs.FilterTypePrune:
			internalFilter.RemapFilter = prune.NewFilter(f.PruneFilterSpec)
		case obs.FilterTypeKubeApiAudit:
			internalFilter.RemapFilter = apiaudit.NewFilter(f.KubeApiAudit)
		case obs.FilterTypeParse:
			internalFilter.RemapFilter = parse.NewParseFilter()
		case obs.FilterTypeDetectMultiline:
			internalFilter.SuppliesTransform = true
			internalFilter.TranformFactory = multilineexception.NewDetectException
		default:
			log.V(0).Error(fmt.Errorf("unknown filter type: %v", f.Type), "This should have been caught by declarative API validation")
		}
		internalFilters[f.Name] = internalFilter
	}
	return internalFilters
}
