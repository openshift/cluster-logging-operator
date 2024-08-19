package parse

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
