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
	Type   types.TransformType `json:"type"`
	Inputs []string            `json:"inputs"`
	Routes map[string]string   `json:"route,omitempty"`
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
