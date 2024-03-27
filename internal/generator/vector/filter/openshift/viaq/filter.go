package viaq

import (
	"encoding/json"
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"strings"
)

const (
	Viaq = "viaq"
)

func New(id, inputs string, labels map[string]string) framework.Element {
	return elements.Remap{
		ComponentID: id,
		Inputs:      inputs,
		VRL: strings.Join(helpers.TrimSpaces([]string{
			containerLogs(labels),
		}), "\n"),
	}
}

func containerLogs(labels map[string]string) string {
	labelsVRL := ""
	if len(labels) != 0 {
		s, _ := json.Marshal(labels)
		labelsVRL = fmt.Sprintf(".openshift.labels = %s", s)
	}
	return fmt.Sprintf(`
if .log_source == "container" {
  %s
}
`, strings.Join(helpers.TrimSpaces([]string{
		ClusterID,
		FixLogLevel,
		HandleEventRouterLog,
		RemoveSourceType,
		RemoveStream,
		RemovePodIPs,
		RemoveNodeLabels,
		RemoveTimestampEnd,
		FixTimestampField,
		labelsVRL,
		VRLOpenShiftSequence,
		VRLDedotLabels,
	}), "\n"))
}
