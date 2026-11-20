package api

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

type Transforms map[string]types.Transform

func (tMap *Transforms) Add(id string, o types.Transform) {
	(*tMap)[id] = o
}
func (tMap *Transforms) Merge(o Transforms) {
	for id, oTransform := range o {
		(*tMap)[id] = oTransform
	}
}
