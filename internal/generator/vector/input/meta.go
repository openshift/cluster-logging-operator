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
	setEnvelope = `. = {"_internal": .}`

	fmtLogType = `
._internal.log_source = %q
._internal.log_type = %q
`
	ParseStructured = `._internal.structured = parse_json!(string!(._internal.message))`
)

func NewInternalNormalization(id string, logSource, logType interface{}, inputs string, addVRLs ...string) framework.Element {
	vrls := []string{
		setEnvelope,
		fmt.Sprintf(fmtLogType, logSource, logType),
		`._internal.hostname = get_env_var("VECTOR_SELF_NODE_NAME") ?? ""`,
		viaq.SetClusterID,
		viaq.SetTimestampField,
		viaq.FixLogLevel,
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
