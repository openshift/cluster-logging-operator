package openshift

type ParseFilter struct{}

func NewParseFilter() ParseFilter {
	return ParseFilter{}
}

func (f ParseFilter) VRL() (string, error) {
	return `
	if .log_type == "application" {
		parsed, err = parse_json(.message)
		if err == null {
			.structured = parsed
			del(.message)
		}
	}
	`, nil
}
