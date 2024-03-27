package viaq

import (
	"encoding/json"
	"fmt"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	"strings"
)

const (
	Viaq        = "viaq"
	ViaqJournal = "viaqjournal"
)

func New(id string, inputs []string, labels map[string]string, inputSpecs []logging.InputSpec) framework.Element {

	vrls := auditHost([]string{}, inputSpecs)
	vrls = auditKube(vrls, inputSpecs)
	vrls = auditOpenShift(vrls, inputSpecs)
	vrls = auditOVN(vrls, inputSpecs)
	vrls = containerSource(vrls, inputSpecs, labels)
	vrls = journalSource(vrls, inputSpecs)
	return elements.Remap{
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
		VRL:         strings.Join(vrls, "\n"),
	}
}

func containerLogs(labels map[string]string) string {
	labelsVRL := ""
	if len(labels) != 0 {
		s, _ := json.Marshal(labels)
		labelsVRL = fmt.Sprintf(".openshift.labels = %s", s)
	}
	return fmt.Sprintf(`
if .log_source == "%s" {
  %s
}
`, logging.InfrastructureSourceContainer, strings.Join(helpers.TrimSpaces([]string{
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

func auditKube(vrls []string, inputs []logging.InputSpec) []string {
	if hasSource(inputs, logging.InputNameAudit, logging.AuditSourceKube) {
		vrls = append(vrls, auditK8sLogs())
	}
	return vrls
}

func auditOpenShift(vrls []string, inputs []logging.InputSpec) []string {
	if hasSource(inputs, logging.InputNameAudit, logging.AuditSourceOpenShift) {
		vrls = append(vrls, auditOpenshiftLogs())
	}
	return vrls
}
func auditOVN(vrls []string, inputs []logging.InputSpec) []string {
	if hasSource(inputs, logging.InputNameAudit, logging.AuditSourceOVN) {
		vrls = append(vrls, auditOVNLogs())
	}
	return vrls
}
func auditHost(vrls []string, inputs []logging.InputSpec) []string {
	if hasSource(inputs, logging.InputNameAudit, logging.AuditSourceAuditd) {
		vrls = append(vrls, auditHostLogs())
	}
	return vrls
}
func containerSource(vrls []string, inputs []logging.InputSpec, labels map[string]string) []string {
	if hasSource(inputs, logging.InputNameApplication, "") || hasSource(inputs, logging.InputNameInfrastructure, logging.InfrastructureSourceContainer) {
		vrls = append(vrls, containerLogs(labels))
	}
	return vrls
}

func journalSource(vrls []string, inputs []logging.InputSpec) []string {
	if HasJournalSource(inputs) {
		vrls = append(vrls, journalLogs())
	}
	return vrls
}

func HasJournalSource(inputs []logging.InputSpec) bool {
	return hasSource(inputs, logging.InputNameInfrastructure, logging.InfrastructureSourceNode)
}

func hasSource(inputSpecs []logging.InputSpec, logType, logSource string) bool {
	for _, i := range inputSpecs {
		switch logType {
		case logging.InputNameApplication:
			if i.Application != nil {
				return true
			}
		case logging.InputNameAudit:
			if i.Audit != nil && (sets.NewString(i.Audit.Sources...).Has(logSource) || len(i.Audit.Sources) == 0) {
				return true
			}
		case logging.InputNameInfrastructure:
			if i.Infrastructure != nil && (sets.NewString(i.Infrastructure.Sources...).Has(logSource) || len(i.Infrastructure.Sources) == 0) {
				return true
			}
		}

	}
	return false
}
