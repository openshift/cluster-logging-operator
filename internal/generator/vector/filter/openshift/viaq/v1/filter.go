package v1

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"strings"
)

const (
	Viaq               = "viaq"
	logSourceContainer = string(obs.ApplicationSourceContainer)
)

func New(id string, inputs []string, inputSpecs []obs.InputSpec) framework.Element {

	vrls := []string{
		SetOpenShiftSequence,
		SetHostnameOnRoot,
		SetLogTypeOnRoot,
		SetLogSourceOnRoot,
		SetOpenShiftOnRoot,
	}
	vrls = auditHost(vrls, inputSpecs)
	vrls = auditKube(vrls, inputSpecs)
	vrls = auditOpenShift(vrls, inputSpecs)
	vrls = auditOVN(vrls, inputSpecs)
	vrls = containerSource(vrls, inputSpecs)
	vrls = journalSource(vrls, inputSpecs)
	vrls = receiverSource(vrls, inputSpecs)
	vrls = append(vrls,
		MergeStructuredIntoRoot,
		`.timestamp = ._internal.timestamp`,
		`."@timestamp" = ._internal.timestamp`,
		SetLogLevelOnRoot,
	)
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
`, logSourceContainer, strings.Join(helpers.TrimSpaces([]string{
		HandleEventRouterLog,
		VRLDedotLabels,
		SetKubernetesOnRoot,
		SetMessageOnRoot,
	}), "\n"))
}

func auditKube(vrls []string, inputs internalobs.Inputs) []string {
	if inputs.HasAuditSource(obs.AuditSourceKube) {
		vrls = append(vrls, auditK8sLogs())
	}
	return vrls
}

func auditOpenShift(vrls []string, inputs internalobs.Inputs) []string {
	if inputs.HasAuditSource(obs.AuditSourceOpenShift) {
		vrls = append(vrls, auditOpenshiftLogs())
	}
	return vrls
}

func auditOVN(vrls []string, inputs internalobs.Inputs) []string {
	if inputs.HasAuditSource(obs.AuditSourceOVN) {
		vrls = append(vrls, auditOVNLogs())
	}
	return vrls
}

func auditHost(vrls []string, inputs internalobs.Inputs) []string {
	if inputs.HasAuditSource(obs.AuditSourceAuditd) {
		vrls = append(vrls, auditHostLogs())
	}
	return vrls
}

func containerSource(vrls []string, inputs internalobs.Inputs) []string {
	if inputs.HasContainerSource() {
		vrls = append(vrls, containerLogs())
	}
	return vrls
}

func journalSource(vrls []string, inputs internalobs.Inputs) []string {
	if inputs.HasJournalSource() {
		vrls = append(vrls, journalLogs())
	}
	return vrls
}

func receiverSource(vrls []string, inputs internalobs.Inputs) []string {
	if inputs.HasReceiverSource() {
		vrls = append(vrls, receiverLogs())
	}
	return vrls
}
