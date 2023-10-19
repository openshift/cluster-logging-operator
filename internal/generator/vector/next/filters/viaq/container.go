package viaq

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	common "github.com/openshift/cluster-logging-operator/internal/generator/vector"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
	"strings"
)

const (
	RenameMeta = `
.kubernetes.labels = del(.kubernetes.pod_labels)
.kubernetes.namespace_id = del(.kubernetes.namespace_uid)
.kubernetes.namespace_name = del(.kubernetes.pod_namespace)
.kubernetes.annotations = del(.kubernetes.pod_annotations)
.kubernetes.pod_id = del(.kubernetes.pod_uid)
.hostname = del(.kubernetes.pod_node_name)
`
)

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
