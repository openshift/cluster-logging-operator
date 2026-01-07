package filter

import (
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/adapters"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/multilineexception"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/parse"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/drop"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/openshift"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/prune"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/apiaudit"
)

func NewInternalFilterMap(filters map[string]*obs.FilterSpec) map[string]*adapters.InternalFilterSpec {
	internalFilters := map[string]*adapters.InternalFilterSpec{}
	for _, f := range filters {
		internalFilter := &adapters.InternalFilterSpec{FilterSpec: f}
		switch f.Type {
		case obs.FilterTypeOpenshiftLabels:
			internalFilter.Factory = func(inputs ...string) types.Transform {
				return openshift.New(f.OpenshiftLabels, inputs...)
			}
		case obs.FilterTypeDrop:
			internalFilter.Factory = func(inputs ...string) types.Transform {
				return drop.New(f.DropTestsSpec, inputs...)
			}
		case obs.FilterTypePrune:
			internalFilter.Factory = func(inputs ...string) types.Transform {
				return prune.New(f.PruneFilterSpec, inputs...)
			}
		case obs.FilterTypeKubeAPIAudit:
			internalFilter.Factory = func(inputs ...string) types.Transform {
				return apiaudit.New(f.KubeAPIAudit, inputs...)
			}
		case obs.FilterTypeParse:
			internalFilter.Factory = func(inputs ...string) types.Transform {
				return parse.New(inputs...)
			}
		case obs.FilterTypeDetectMultiline:
			internalFilter.Factory = multilineexception.New
		default:
			log.V(0).Error(fmt.Errorf("unknown filter type: %v", f.Type), "This should have been caught by declarative API validation")
		}
		internalFilters[f.Name] = internalFilter
	}
	return internalFilters
}
