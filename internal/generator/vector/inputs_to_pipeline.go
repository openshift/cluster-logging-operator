package vector

import (
	"encoding/json"
	"fmt"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

var (
	UserDefinedInput = fmt.Sprintf("%s.%%s", RouteApplicationLogs)
)

func InputsToPipelines(spec *logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
	el := []generator.Element{}
	userDefined := spec.InputMap()
	for _, p := range spec.Pipelines {
		vrl := SrcPassThrough
		if p.Labels != nil && len(p.Labels) != 0 {
			s, _ := json.Marshal(p.Labels)
			vrl = fmt.Sprintf(".openshift.labels = %s", s)
		}
		inputs := []string{}
		for _, i := range p.InputRefs {
			if _, ok := userDefined[i]; ok {
				inputs = append(inputs, fmt.Sprintf(UserDefinedInput, i))
			} else {
				inputs = append(inputs, i)
			}
		}
		r := Remap{
			ComponentID: p.Name,
			Inputs:      helpers.MakeInputs(inputs...),
			VRL:         vrl,
		}
		el = append(el, r)

	}
	return el
}
