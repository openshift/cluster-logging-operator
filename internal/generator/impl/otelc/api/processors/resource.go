package processors

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/types"
)

// Resource is the resource processor. It sets or upserts resource attributes
// on every record passing through it. It is used to carry static values that
// originate in the forwarder spec rather than the cluster, such as the
// openshift labels configured by a labels filter.
// See: https://opentelemetry.io/docs/collector/configuration/#resource
type Resource struct {
	// id is the processor identifier in the format "type" or "type/name"
	// It is not serialized to YAML
	id string

	// Attributes are the resource attributes to set
	Attributes []Attribute `yaml:"attributes"`
}

// Attribute is a resource attribute action
type Attribute struct {
	// Key is the attribute name
	Key string `yaml:"key"`

	// Value is the attribute value
	Value string `yaml:"value"`

	// Action defaults to upsert when empty
	Action string `yaml:"action,omitempty"`
}

// NewResource creates a new resource processor with the given name
func NewResource(name string) *Resource {
	return &Resource{
		id: types.MakeComponentID(string(types.ProcessorTypeResource), name),
	}
}

// ID returns the processor identifier in the format "type" or "type/name"
func (p *Resource) ID() string {
	return p.id
}

// ProcessorType extracts the processor type from the ID
func (p *Resource) ProcessorType() types.ProcessorType {
	componentType, _ := types.ParseComponentID(p.id)
	return types.ProcessorType(componentType)
}

// SetAttribute upserts a resource attribute
func (p *Resource) SetAttribute(key, value string) *Resource {
	p.Attributes = append(p.Attributes, Attribute{Key: key, Value: value, Action: "upsert"})
	return p
}
