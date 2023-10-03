package viaq

import (
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	common "github.com/openshift/cluster-logging-operator/internal/generator/vector"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/next/filters"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	"strings"
)

const (
	Name = "viaq"

	RenameMeta = `
.kubernetes.labels = del(.kubernetes.pod_labels)
.kubernetes.namespace_id = del(.kubernetes.namespace_uid)
.kubernetes.namespace_name = del(.kubernetes.pod_namespace)
.kubernetes.annotations = del(.kubernetes.pod_annotations)
.kubernetes.pod_id = del(.kubernetes.pod_uid)
.hostname = del(.kubernetes.pod_node_name)
`
)

type Filter struct{}

func New(spec *logging.FilterSpec) filters.Filter {
	return &Filter{}
}

func (f *Filter) Elements(inputs []string, pipeline logging.PipelineSpec, spec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
	types := logTypesFor(pipeline, spec.InputMap())
	log.V(4).Info("init viaq filter", "pipeline", pipeline, "types", types)
	var el []generator.Element
	if types.Has(logging.InputNameContainer) || types.Has(logging.InputNameApplication) || types.Has(logging.InputNameInfrastructure) {
		el = append(el, NormalizeContainerLogs(f.TranformsName(pipeline), inputs...)...)
	}

	// TODO - How to distinguish between container and journal logs which have different normalization
	//if types.Has(logging.InputNameNode) || types.Has(logging.InputNameInfrastructure) {
	//	el = append(el, normalize.JournalLogs(helpers.MakeInputs(inputs...), "")...)
	//}

	return el
}

func (f *Filter) TranformsName(pipeline logging.PipelineSpec) string {
	return fmt.Sprintf("%s_logs_%s", pipeline.Name, Name)
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

func NormalizeContainerLogs(id string, inputs ...string) []generator.Element {
	return []generator.Element{
		elements.Remap{
			ComponentID: id,
			Inputs:      helpers.MakeInputs(inputs...),
			VRL: strings.Join(helpers.TrimSpaces([]string{
				RenameMeta,
				normalize.ClusterID,
				common.FixLogLevel,
				common.HandleEventRouterLog,
				common.RemoveSourceType,
				common.RemoveStream,
				common.RemovePodIPs,
				common.RemoveNodeLabels,
				common.RemoveTimestampEnd,
				normalize.FixTimestampField,
				SetLogType,
			}), "\n"),
		},
	}
}

var (
	SetLogType = fmt.Sprintf(`
namespace_name = string!(.kubernetes.namespace_name)
if match_any(namespace_name, [r'^(default|kube|openshift)$',r'^(kube|openshift)-.*']) {
  %s
} else {
  %s
}
`, common.AddLogTypeInfra, common.AddLogTypeApp)
)
