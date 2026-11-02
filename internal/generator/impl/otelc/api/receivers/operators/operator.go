package operators

// Operator represents a generic log processing operator
// The actual operator configuration is flexible and depends on the operator type
// This struct provides maximum flexibility for defining custom operators
type Operator struct {
	Type   OperatorType           `yaml:"type"`             // Required: operator type
	ID     string                 `yaml:"id,omitempty"`     // Optional unique ID
	Output string                 `yaml:"output,omitempty"` // Optional operator ID to send output
	Config map[string]interface{} `yaml:",inline"`          // Operator-specific configuration fields
}
