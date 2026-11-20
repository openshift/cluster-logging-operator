package transforms

import (
	"sort"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

type Reduce struct {
	Type            types.TransformType `json:"type"`
	Inputs          []string            `json:"inputs"`
	ExpireAfterMs   uint64              `json:"expire_after_ms,omitempty"`
	MaxEvents       uint64              `json:"max_events,omitempty"`
	GroupBy         []string            `json:"group_by,omitempty"`
	MergeStrategies *MergeStrategies    `json:"merge_strategies,omitempty"`
}

type MergeStrategiesResourceType string
type MergeStrategiesLogRecordsType string

const (
	MergeStrategiesResourceRetain  MergeStrategiesResourceType   = "retain"
	MergeStrategiesLogRecordsArray MergeStrategiesLogRecordsType = "array"
)

type MergeStrategies struct {
	Resource   MergeStrategiesResourceType   `json:"resource,omitempty"`
	LogRecords MergeStrategiesLogRecordsType `json:"logRecords,omitempty"`
}

func NewReduce(init func(*Reduce), inputs ...string) *Reduce {
	sort.Strings(inputs)
	t := &Reduce{
		Type:   types.TransformTypeReduce,
		Inputs: inputs,
	}
	if init != nil {
		init(t)
	}
	return t
}

func (t *Reduce) TransformType() types.TransformType {
	return t.Type
}
