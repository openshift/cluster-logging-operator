package transforms

import (
	"fmt"
	"sort"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

const (
	Unmatched = "unmatched"
)

type Route struct {
	Type   types.TransformType `json:"type" yaml:"type" toml:"type"`
	Inputs []string            `json:"inputs" yaml:"inputs" toml:"inputs"`
	Routes map[string]string   `json:"route,omitempty" yaml:"route,omitempty" toml:"route,omitempty"`
}

func NewRoute(init func(*Route), inputs ...string) *Route {
	sort.Strings(inputs)
	t := &Route{
		Type:   types.TransformTypeRoute,
		Inputs: inputs,
	}
	if init != nil {
		init(t)
	}
	return t
}

func (t *Route) TransformType() types.TransformType {
	return t.Type
}

func UnmatchedRoute(baseId string) string {
	return fmt.Sprintf("%s._%s", baseId, Unmatched)
}
