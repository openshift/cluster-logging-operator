package apiaudit

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/next/filters"
)

const (
	Name = "kubeAPIAudit"
)

type Filter struct {
	*logging.FilterSpec
}

func New(spec *logging.FilterSpec) filters.Filter {
	return &Filter{
		spec,
	}
}

func (f *Filter) Elements(inputs []string, pipeline logging.PipelineSpec, spec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
	if vrl, err := filter.RemapVRL(f.FilterSpec); err == nil {
		return []generator.Element{
			elements.Remap{
				ComponentID: filters.MakeID(pipeline.Name, Name),
				Inputs:      helpers.MakeInputs(inputs...),
				VRL:         vrl,
			},
		}
	} else {
		log.V(0).Error(err, "Unable to configure filter", "name", f.Name, "spec", f.FilterSpec)
	}
	return nil
}

func (f *Filter) TranformsNames(pipeline logging.PipelineSpec) []string {
	return []string{filters.MakeID(pipeline.Name, Name)}
}
