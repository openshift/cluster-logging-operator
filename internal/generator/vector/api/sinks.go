package api

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

type Sinks map[string]types.Sink

func (sinkMap *Sinks) Add(id string, sink types.Sink) {
	(*sinkMap)[id] = sink
}
func (sinkMap *Sinks) Merge(sinks Sinks) {
	for id, s := range sinks {
		(*sinkMap)[id] = s
	}
}
