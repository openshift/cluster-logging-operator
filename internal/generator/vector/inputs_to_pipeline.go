package vector

import (
	"encoding/json"
	"fmt"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

func InputsToPipelines(spec *logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
	el := []generator.Element{}
	for _, p := range spec.Pipelines {
		vrl := SrcPassThrough
		if p.Labels != nil && len(p.Labels) != 0 {
			s, _ := json.Marshal(p.Labels)
			vrl = fmt.Sprintf(".openshift.labels = %s", s)
		}
		r := Remap{
			ComponentID: p.Name,
			Inputs:      helpers.MakeInputs(p.InputRefs...),
			VRL:         vrl,
		}
		el = append(el, r)

	}
	return el
}
