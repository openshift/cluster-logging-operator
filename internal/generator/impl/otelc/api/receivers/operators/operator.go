package operators

import "fmt"

// Operator represents a generic log processing operator
// The actual operator configuration is flexible and depends on the operator type
// This struct provides maximum flexibility for defining custom operators
type Operator struct {
	Type   OperatorType           `yaml:"type"`             // Required: operator type
	ID     string                 `yaml:"id,omitempty"`     // Optional unique ID
	Output string                 `yaml:"output,omitempty"` // Optional operator ID to send output
	Config map[string]interface{} `yaml:",inline"`          // Operator-specific configuration fields
}

// ResourceAttribute returns the stanza field expression addressing a resource
// attribute, e.g. ResourceAttribute("k8s.pod.name") is `resource["k8s.pod.name"]`.
// Use it instead of formatting the expression by hand so the quoting stays
// consistent across operators
func ResourceAttribute(name string) string {
	return fmt.Sprintf("resource[%q]", name)
}
