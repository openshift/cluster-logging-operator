package viaq

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"k8s.io/utils/set"
	"strings"
)

const (
	Viaq        = "viaq"
	ViaqJournal = "viaqjournal"
	ViaqDedot   = "viaqdedot"
)

func New(id string, inputs []string, inputSpecs []obs.InputSpec) framework.Element {

	vrls := auditHost([]string{}, inputSpecs)
	vrls = auditKube(vrls, inputSpecs)
	vrls = auditOpenShift(vrls, inputSpecs)
	vrls = auditOVN(vrls, inputSpecs)
	vrls = containerSource(vrls, inputSpecs)
	vrls = journalSource(vrls, inputSpecs)
	return elements.Remap{
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
		VRL:         strings.Join(vrls, "\n"),
	}
}

func containerLogs() string {
	return fmt.Sprintf(`
if .log_source == "%s" {
  %s
}
`, obs.InfrastructureSourceContainer, strings.Join(helpers.TrimSpaces([]string{
		ClusterID,
		FixLogLevel,
		HandleEventRouterLog,
		RemovePartial,
		RemoveFile,
		RemoveSourceType,
		HandleStream,
		RemovePodIPs,
		RemoveNodeLabels,
		RemoveTimestampEnd,
		FixTimestampField,
		VRLOpenShiftSequence,
	}), "\n"))
}

func auditKube(vrls []string, inputs []obs.InputSpec) []string {
	if hasAuditSource(inputs, obs.AuditSourceKube) {
		vrls = append(vrls, auditK8sLogs())
	}
	return vrls
}

func auditOpenShift(vrls []string, inputs []obs.InputSpec) []string {
	if hasAuditSource(inputs, obs.AuditSourceOpenShift) {
		vrls = append(vrls, auditOpenshiftLogs())
	}
	return vrls
}

func auditOVN(vrls []string, inputs []obs.InputSpec) []string {
	if hasAuditSource(inputs, obs.AuditSourceOVN) {
		vrls = append(vrls, auditOVNLogs())
	}
	return vrls
}

func auditHost(vrls []string, inputs []obs.InputSpec) []string {
	if hasAuditSource(inputs, obs.AuditSourceAuditd) {
		vrls = append(vrls, auditHostLogs())
	}
	return vrls
}

func containerSource(vrls []string, inputs []obs.InputSpec) []string {
	if hasContainerSource(inputs) {
		vrls = append(vrls, containerLogs())
	}
	return vrls
}

func journalSource(vrls []string, inputs []obs.InputSpec) []string {
	if HasJournalSource(inputs) {
		vrls = append(vrls, journalLogs())
	}
	return vrls
}

func HasJournalSource(inputs []obs.InputSpec) bool {
	for _, i := range inputs {
		if i.Type == obs.InputTypeInfrastructure && i.Infrastructure != nil && (len(i.Infrastructure.Sources) == 0 || set.New(i.Infrastructure.Sources...).Has(obs.InfrastructureSourceNode)) {
			return true
		}
	}
	return false
}

func hasContainerSource(inputSpecs []obs.InputSpec) bool {
	for _, i := range inputSpecs {
		if i.Type == obs.InputTypeApplication {
			return true
		}
		if i.Type == obs.InputTypeInfrastructure && i.Infrastructure != nil && (len(i.Infrastructure.Sources) == 0 || set.New(i.Infrastructure.Sources...).Has(obs.InfrastructureSourceContainer)) {
			return true
		}
	}
	return false
}

func hasAuditSource(inputSpecs []obs.InputSpec, logSource obs.AuditSource) bool {
	for _, i := range inputSpecs {
		if i.Type == obs.InputTypeAudit && i.Audit != nil && (set.New(i.Audit.Sources...).Has(logSource) || len(i.Audit.Sources) == 0) {
			return true
		}
	}
	return false
}
