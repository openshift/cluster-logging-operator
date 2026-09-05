package parse

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/transforms"
)

type Filter struct {
	spec *obs.ParseFilterSpec
}

func NewParseFilter(spec *obs.ParseFilterSpec) Filter {
	return Filter{spec: spec}
}

func (f Filter) VRL() (string, error) {
	if f.spec != nil && f.spec.AddToRoot {
		removeMessage := ""
		if !f.spec.PreserveMessage {
			removeMessage = "\n\t\t\t._internal.parse_remove_message = true"
		}
		return `
	if ._internal.log_source == "container" {
		parsed, err = parse_json(._internal.message)
		if err == null && is_object(parsed) {
			. = merge!(parsed, .)` + removeMessage + `
		}
	}
	`, nil
	}

	return `
	if ._internal.log_source == "container" {
		parsed, err = parse_json(._internal.message)
		if err == null {
			._internal.structured = parsed
		}
	}
	`, nil
}

func New(spec *obs.ParseFilterSpec, inputs ...string) *transforms.Remap {
	vrl, _ := NewParseFilter(spec).VRL()
	return transforms.NewRemap(vrl, inputs...)
}
