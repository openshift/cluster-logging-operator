package input

import (
	"fmt"

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
	ParseStructured = `._internal.structured = parse_json!(string!(._internal.message))`
	
	setEnvelope       = `. = {"_internal": .}`
	setHostName       = `._internal.hostname = get_env_var("VECTOR_SELF_NODE_NAME") ?? ""`
	setClusterID      = `._internal.openshift.cluster_id = "${OPENSHIFT_CLUSTER_ID:-}"`
	setTimestampField = `ts = del(._internal.timestamp); if !exists(._internal."@timestamp") {._internal."@timestamp" = ts}`
)

func NewInternalNormalization(id string, logSource, logType interface{}, inputs string, addVRLs ...string) framework.Element {
	vrls := []string{
		setEnvelope,
		fmt.Sprintf(fmtLogType, logSource, logType),
		setHostName,
		setClusterID,
		setTimestampField,
		viaq.SetLogLevel,
	}
	for _, vrl := range addVRLs {
		vrls = append(vrls, vrl)
	}
	return elements.Remap{
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs),
		VRL:         strings.Join(vrls, "\n"),
	}
	if visit != nil {
		visit(&ele)
	}
	return ele
}
