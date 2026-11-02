package operators

// TimestampConfig defines how to parse timestamps
type TimestampConfig struct {
	ParseFrom  string `yaml:"parse_from"`            // Field containing the timestamp
	Layout     string `yaml:"layout,omitempty"`      // strptime layout
	LayoutType string `yaml:"layout_type,omitempty"` // "strptime", "gotime", "epoch"
	Location   string `yaml:"location,omitempty"`    // Timezone location (e.g., "UTC")
}

// SeverityConfig defines how to parse log severity/level
type SeverityConfig struct {
	ParseFrom string            `yaml:"parse_from"`        // Field containing severity
	Mapping   map[string]string `yaml:"mapping,omitempty"` // Map values to standard severities
}

// RouteEntry defines a routing rule
type RouteEntry struct {
	Output string `yaml:"output"` // Operator ID to route to
	Expr   string `yaml:"expr"`   // Boolean expression for routing
}
