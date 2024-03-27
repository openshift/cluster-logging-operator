package viaq

import (
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"strings"
)

const (
	VRLDedotLabels = `if exists(.kubernetes.namespace_labels) {
    for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
      newkey = replace(key, r'[\./]', "_") 
      .kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
      if newkey != key {
        .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
      }
    }
}
if exists(.kubernetes.labels) {
    for_each(object!(.kubernetes.labels)) -> |key,value| { 
      newkey = replace(key, r'[\./]', "_") 
      .kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
      if newkey != key {
        .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
      }
    }
}`
	VRLOpenShiftSequence = `.openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")`
)

// DedotLabels replaces '[\./]' with '_' as well as adds the openshift.sequence id
func DedotLabels(id string, inputs []string) Element {
	return elements.Remap{
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
		VRL:         strings.Join([]string{VRLOpenShiftSequence, VRLDedotLabels}, "\n"),
	}
}
