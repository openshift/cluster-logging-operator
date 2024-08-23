package viaq

import (
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"strings"
)

const (
	VRLDedotLabels = `
if .log_source == "Container" {
  if exists(.kubernetes.namespace_labels) {
    ._internal.kubernetes.namespace_labels = .kubernetes.namespace_labels
    for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
      newkey = replace(key, r'[\./]', "_") 
      .kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
      if newkey != key {.kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)}
    }
  }
  if exists(.kubernetes.labels) {
    ._internal.kubernetes.labels = .kubernetes.labels
    for_each(object!(.kubernetes.labels)) -> |key,value| { 
      newkey = replace(key, r'[\./]', "_") 
      .kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
      if newkey != key {.kubernetes.labels = remove!(.kubernetes.labels,[key],true)}
    }
  }
}
if exists(.openshift.labels) {for_each(object!(.openshift.labels)) -> |key,value| {
  newkey = replace(key, r'[\./]', "_") 
  .openshift.labels = set!(.openshift.labels,[newkey],value)
  if newkey != key {.openshift.labels = remove!(.openshift.labels,[key],true)}
}}
`
)

func NewDedot(id string, inputs ...string) Element {
	return DedotLabels(id, inputs...)
}

// DedotLabels replaces dots and forward slashes with underscores
func DedotLabels(id string, inputs ...string) Element {
	return elements.Remap{
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
		VRL:         strings.Join([]string{VRLDedotLabels}, "\n"),
	}
}
