package input

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/openshift/viaq"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"strings"
)

const (
	fmtLogType = `
._internal.log_source = %q
._internal.log_type = %q
`
	parseStructured = `
._internal.structured = parse_json!(string!(._internal.message))
._internal = merge!(._internal,._internal.structured)
`

	setClusterID            = `._internal.openshift = { "cluster_id": "${OPENSHIFT_CLUSTER_ID:-}"}`
	setEnvelope             = `. = {"_internal": .}`
	setEnvelopeToStructured = `. = {"_internal": {"structured": .}}`
	setHostName             = `._internal.hostname = get_env_var("VECTOR_SELF_NODE_NAME") ?? ""`
	setTimestampField       = `ts = del(._internal.timestamp); if !exists(._internal."@timestamp") {._internal."@timestamp" = ts}`
)

// NewAuditInternalNormalization returns configuration elements to normalize audit log entries to an internal, common data model
func NewAuditInternalNormalization(id string, logSource obs.AuditSource, inputs string, parseIntoStructured bool, addVRLs ...string) framework.Element {
	var vrls []string
	if parseIntoStructured {
		vrls = append(vrls, setEnvelope, parseStructured)
	}
	vrls = append(vrls,
		fmt.Sprintf(fmtLogType, logSource, obs.InputTypeAudit),
		setHostName,
		setClusterID,
		setTimestampField,
	)
	vrls = append(vrls, addVRLs...)
	return elements.Remap{
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs),
		VRL:         strings.Join(vrls, "\n"),
	}
}

// NewInternalNormalization returns configuration elements to normalize log entries to an internal, common data model
func NewInternalNormalization(id string, logSource, logType interface{}, inputs string, addVRLs ...string) framework.Element {
	vrls := []string{
		setEnvelope,
		fmt.Sprintf(fmtLogType, logSource, logType),
		setHostName,
		setClusterID,
		setTimestampField,
		viaq.SetLogLevel,
	}
	vrls = append(vrls, addVRLs...)
	return elements.Remap{
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs),
		VRL:         strings.Join(vrls, "\n"),
	}
}

// NewJournalInternalNormalization returns configuration elements to normalize journal log entries to an internal, common data model
func NewJournalInternalNormalization(id string, logSource interface{}, envelopeVrl, inputs string, addVRLs ...string) framework.Element {
	vrls := []string{
		envelopeVrl,
		fmt.Sprintf(fmtLogType, logSource, obs.InputTypeInfrastructure),
		setTimestampField,
		setClusterID,
	}
	vrls = append(vrls, addVRLs...)
	return elements.Remap{
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs),
		VRL:         strings.Join(vrls, "\n"),
	}
}
