package transforms

import (
	"sort"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

type Filter struct {
	Type types.TransformType `json:"type" yaml:"type" toml:"type"`

	// Inputs is the IDs of the components feeding into this component
	Inputs []string `json:"inputs" yaml:"inputs" toml:"inputs"`

	// Source is the VRL script used for the remap transformation
	Condition Condition `json:"condition,omitempty" yaml:"condition,omitempty" toml:"condition,omitempty" multiline:"true" literal:"true"`
}

type Condition string

func NewFilter(source string, inputs ...string) *Filter {
	sort.Strings(inputs)
	return &Filter{
		Type:      types.TransformTypeFilter,
		Inputs:    inputs,
		Condition: Condition(source),
	}
}

func (t *Filter) TransformType() types.TransformType {
	return t.Type
}
