package viaq

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	common "github.com/openshift/cluster-logging-operator/internal/generator/vector"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/next/filters"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/next/source"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	"strings"
)

const (
	Name = "viaq"
)

type Filter struct {
	names []string
}

func New(spec *logging.FilterSpec) filters.Filter {
	return &Filter{}
}

//// IDs returns a set of ids produced by this component for a given pipeline
//func (f *Filter) IDs(spec logging.ClusterLogForwarderSpec, pipelineName string) []string {
//	inputs := spec.InputMap()
//}

func (f *Filter) Elements(inputs []string, pipeline logging.PipelineSpec, spec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {

	var el []generator.Element
	if elInputs, found := hasInputsForLogTypes(inputs, source.ContainerLogTypes...); found {
		el = append(el, NormalizeContainerLogs(filters.MakeID(pipeline.Name, Name), elInputs...)...)
	}
	if elInputs, found := hasInputsForLogTypes(inputs, source.JournalLogTypes...); found {
		el = append(el, JournalLogs(helpers.MakeInputs(elInputs...), filters.MakeID(pipeline.Name, Name, logging.InputNameNode))...)
	}
	if elInputs, found := hasInputsForLogTypes(inputs, source.AuditLogTypes...); found {
		el = append(el, common.NormalizeHostAuditLogs(makeID(pipeline, source.AuditHost), limitAuditInputsTo(elInputs, source.AuditHost)...)...)
		el = append(el, common.NormalizeK8sAuditLogs(makeID(pipeline, source.AuditKubernetes), limitAuditInputsTo(elInputs, source.AuditKubernetes)...)...)
		el = append(el, common.NormalizeOpenshiftAuditLogs(makeID(pipeline, source.AuditOpenShift), limitAuditInputsTo(elInputs, source.AuditOpenShift)...)...)
		el = append(el, common.NormalizeOVNAuditLogs(makeID(pipeline, source.AuditOVN), limitAuditInputsTo(elInputs, source.AuditOVN)...)...)
	}

	return el
}

func hasInputsForLogTypes(inputs []string, logType ...string) ([]string, bool) {
	result := sets.NewString()
	for _, lt := range logType {
		for _, i := range inputs {
			if strings.HasSuffix(i, lt) {
				result.Insert(i)
			}
		}
	}
	return result.List(), result.Len() > 0
}

func limitAuditInputsTo(inputs []string, auditType string) []string {
	result := sets.NewString()
	for _, i := range inputs {
		if strings.HasSuffix(i, auditType) {
			result.Insert(i)
		}
	}
	if result.Len() == 0 {
		return inputs
	}
	return result.List()
}

func makeID(p logging.PipelineSpec, part string) string {
	return filters.MakeID(p.Name, part)
}

func (f *Filter) TranformsNames(pipeline logging.PipelineSpec) []string {
	//types := logTypesFor(pipeline, spec.InputMap())
	//log.V(4).Info("init viaq filter", "pipeline", pipeline, "types", types)
	return []string{
		filters.MakeID(pipeline.Name, Name),
		filters.MakeID(pipeline.Name, Name, logging.InputNameNode),
		filters.MakeID(pipeline.Name, Name, source.AuditHost),
		filters.MakeID(pipeline.Name, Name, source.AuditKubernetes),
		filters.MakeID(pipeline.Name, Name, source.AuditOpenShift),
		filters.MakeID(pipeline.Name, Name, source.AuditOVN),
	}
}

func logTypesFor(pipeline logging.PipelineSpec, inputs map[string]*logging.InputSpec) sets.String {
	types := sets.NewString()
	for _, inputName := range pipeline.InputRefs {
		if logging.ReservedInputNames.Has(inputName) {
			types.Insert(inputName) // Use name as type.
		} else {
			if spec, found := inputs[inputName]; found {
				if spec.Application != nil {
					types.Insert(logging.InputNameApplication)
				}
				if spec.Infrastructure != nil {
					types.Insert(logging.InputNameInfrastructure)
				}
				if spec.Audit != nil {
					types.Insert(logging.InputNameAudit)
				}
			}
		}
	}

	return *types
}
