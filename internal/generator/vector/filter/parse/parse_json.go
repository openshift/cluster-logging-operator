package parse

type Filter struct{}

func NewParseFilter() Filter {
	return Filter{}
}

func (f Filter) VRL() (string, error) {
	return `
	if .log_source == "Container" {
		parsed, err = parse_json(.message)
		if err == null {
			.structured = parsed
			del(.message)
		}
	}
	`, nil
}
