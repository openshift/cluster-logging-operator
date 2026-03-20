package parse

import "github.com/openshift/cluster-logging-operator/internal/generator/vector/api/transforms"

type Filter struct{}

func NewParseFilter() Filter {
	return Filter{}
}

func (f Filter) VRL() (string, error) {
	return `
	if ._internal.log_source == "container" {
		parsed, err = parse_json(._internal.message)
		if err == null {
			._internal.structured = parsed
		}
	}
	`, nil
}

func New(inputs ...string) *transforms.Remap {
	return transforms.NewRemap(`
	if ._internal.log_source == "container" {
		parsed, err = parse_json(._internal.message)
		if err == null {
			._internal.structured = parsed
		}
	}
`, inputs...)
}
