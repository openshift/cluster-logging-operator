package input

import (
	"fmt"

	"strings"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/transforms/remap"
	v1 "github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/openshift/viaq/v1"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
)

const (
	fmtLogSource     = `._internal.log_source = %q`
	fmtLogType       = `._internal.log_type = %q`
	logTypeContainer = `
  # If namespace is infra, label log_type as infra
  if match_any(string!(._internal.kubernetes.namespace_name), [r'^default$', r'^openshift(-.+)?$', r'^kube(-.+)?$']) {
      ._internal.log_type = "infrastructure"
  } else {
      ._internal.log_type = "application"
  }
`
	parseStructured = `
._internal.structured = parse_json!(string!(._internal.message))
._internal = merge!(._internal,._internal.structured)
`

	setClusterID                   = `._internal.openshift = { "cluster_id": "${OPENSHIFT_CLUSTER_ID:-}"}`
	setOpenshiftSequence           = `._internal.openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")`
	setEnvelope                    = `. = {"_internal": .}`
	setKubernetesContainerIOStream = `if exists(._internal.stream) {._internal.kubernetes.container_iostream = ._internal.stream}`
	setEnvelopeToStructured        = `. = {"_internal": {"structured": .}}`
	setHostName                    = `._internal.hostname = get_env_var("VECTOR_SELF_NODE_NAME") ?? ""`
)

// NewAuditInternalNormalization returns configuration elements to normalize audit log entries to an internal, common data model
func NewAuditInternalNormalization(id string, logSource obs.AuditSource, inputs string, parseIntoStructured bool, addVRLs ...string) framework.Element {
	vrls := []string{setEnvelope}
	if parseIntoStructured {
		vrls = append(vrls, parseStructured)
	}
	vrls = append(vrls,
		fmt.Sprintf(fmtLogSource, logSource),
		fmt.Sprintf(fmtLogType, obs.InputTypeAudit),
		setHostName,
		setClusterID,
		setOpenshiftSequence,
	)
	vrls = append(vrls, addVRLs...)
	return remap.New(id, strings.Join(vrls, "\n"), inputs)
}

// NewInternalNormalization returns configuration elements to normalize log entries to an internal, common data model
func NewInternalNormalization(id string, logSource, logType interface{}, inputs string, addVRLs ...string) framework.Element {
	logTypeVRL := fmt.Sprintf(fmtLogType, logType)
	vrls := []string{setEnvelope}

	if logSource == obs.InfrastructureSourceContainer {
		logTypeVRL = logTypeContainer
		// Add kubernetes container iostream for all container sources
		vrls = append(vrls, setKubernetesContainerIOStream)
	}
	vrls = append(vrls, fmt.Sprintf(fmtLogSource, logSource),
		logTypeVRL,
		setHostName,
		setClusterID,
		setOpenshiftSequence,
		v1.SetLogLevel,
	)
	vrls = append(vrls, addVRLs...)
	vrl := strings.Join(vrls, "\n")
	return remap.New(id, vrl, inputs)
}

// NewJournalInternalNormalization returns configuration elements to normalize journal log entries to an internal, common data model
func NewJournalInternalNormalization(id string, logSource interface{}, envelopeVrl, inputs string, addVRLs ...string) framework.Element {
	vrls := []string{
		envelopeVrl,
		fmt.Sprintf(fmtLogSource, logSource),
		fmt.Sprintf(fmtLogType, obs.InputTypeInfrastructure),
		setClusterID,
		setOpenshiftSequence,
	}
	vrls = append(vrls, addVRLs...)
	return remap.New(id, strings.Join(vrls, "\n"), inputs)
}

// NewJournalInternalNormalization returns configuration elements to normalize journal log entries to an internal, common data model
func NewReceiverInternalNormalization(id string, logSource interface{}, envelopeVrl, inputs string, addVRLs ...string) framework.Element {
	vrls := []string{
		envelopeVrl,
		fmt.Sprintf(fmtLogSource, logSource),
		fmt.Sprintf(fmtLogType, obs.InputTypeReceiver),
		`._internal.timestamp = del(._internal.structured.timestamp)`,
		`._internal.message = del(._internal.structured.message)`,
	}
	vrls = append(vrls, addVRLs...)
	return remap.New(id, strings.Join(vrls, "\n"), inputs)
}
