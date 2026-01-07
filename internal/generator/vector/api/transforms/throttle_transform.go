package transforms

import (
	"sort"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

type Throttle struct {
	Type types.TransformType `json:"type" yaml:"type" toml:"type"`

	// Inputs is the IDs of the components feeding into this component
	Inputs []string `json:"inputs" yaml:"inputs" toml:"inputs"`

	WindowSecs uint64 `json:"window_secs,omitempty" yaml:"window_secs,omitempty" toml:"window_secs,omitempty"`
	Threshold  uint64 `json:"threshold,omitempty" yaml:"threshold,omitempty" toml:"threshold,omitempty"`
	KeyField   string `json:"key_field,omitempty" yaml:"key_field,omitempty" toml:"key_field,omitempty"`
}

func NewThrottle(init func(*Throttle), inputs ...string) *Throttle {
	sort.Strings(inputs)
	t := &Throttle{
		Type:   types.TransformTypeThrottle,
		Inputs: inputs,
	}
	if init != nil {
		init(t)
	}
	return t
}

func (t *Throttle) TransformType() types.TransformType {
	return t.Type
}
