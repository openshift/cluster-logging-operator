package operators

// RegexParserOperator parses logs using regular expressions
type RegexParserOperator struct {
	Type   OperatorType `yaml:"type"`             // Must be OperatorTypeRegexParser
	ID     string       `yaml:"id,omitempty"`     // Optional unique ID
	Output string       `yaml:"output,omitempty"` // Next operator ID
	Regex  string       `yaml:"regex"`            // Regex pattern with named capture groups

	// Timestamp configuration
	Timestamp *TimestampConfig `yaml:"timestamp,omitempty"`

	// Severity configuration
	Severity *SeverityConfig `yaml:"severity,omitempty"`

	// Parse from field
	ParseFrom string `yaml:"parse_from,omitempty"` // Default: "body"

	// Parse to field
	ParseTo string `yaml:"parse_to,omitempty"` // Default: "attributes"
}

// NewRegexParser creates a regex parser operator
func NewRegexParser(id, regex string) *RegexParserOperator {
	return &RegexParserOperator{
		Type:  OperatorTypeRegexParser,
		ID:    id,
		Regex: regex,
	}
}
