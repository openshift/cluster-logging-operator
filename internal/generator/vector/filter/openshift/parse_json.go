package openshift

const (
	ParseTypeJson = "json"
	ParseJson     = "parseJSON"
)

func NewParseJSON() (string, error) {
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
