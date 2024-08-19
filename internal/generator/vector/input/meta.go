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
	setEnvelope     = ". = ._internal.message = ."
	ParseStructured = `._internal.structured = merge(._internal.structured, parse_json!(string!(._internal.message))) ?? ._internal.structured`
)

func NewInternalNormalization(id string, logSource, logType interface{}, inputs string, addVRLs ...string) framework.Element {
	vrls := []string{
		fmt.Sprintf("._internal.log_source = %q\n._internal.log_type = %q", logSource, logType),
		viaq.FixHostname,
		viaq.FixLogLevel,
		setEnvelope,
		fmt.Sprintf(".log_source = %q\n.log_type = %q", logSource, logType),
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
