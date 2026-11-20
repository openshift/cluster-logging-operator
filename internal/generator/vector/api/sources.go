package api

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

// Sources is a set of sources by their id.
type Sources map[string]types.Source

func (sourceMap *Sources) Add(id string, source types.Source) {
	(*sourceMap)[id] = source
}
