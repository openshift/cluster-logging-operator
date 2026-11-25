package processors

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/types"
)

// K8sAttributes is the k8sattributes processor. It enriches records with
// Kubernetes pod metadata (namespace, pod, node, workload names) and with pod
// labels and annotations, correlating the record to a pod via the resource
// attributes set by the input receivers.
// See: https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/k8sattributesprocessor
type K8sAttributes struct {
	// id is the processor identifier in the format "type" or "type/name"
	// It is not serialized to YAML
	id string

	// Extract configures which metadata, labels and annotations to add
	Extract *ExtractConfig `yaml:"extract,omitempty"`

	// PodAssociation configures how records are associated with pods. Rules
	// are evaluated in order; the first matching rule wins. When empty the
	// incoming connection IP is used.
	PodAssociation []Association `yaml:"pod_association,omitempty"`

	// Filter restricts which pods the processor considers
	Filter *Filter `yaml:"filter,omitempty"`
}

// ExtractConfig defines the metadata to extract from the Kubernetes API
type ExtractConfig struct {
	// Metadata are the resource attributes to add. Only names listed here may
	// be used for resource_attribute based pod association.
	Metadata []string `yaml:"metadata,omitempty"`

	// Labels are extraction rules for pod, namespace or node labels
	Labels []FieldExtract `yaml:"labels,omitempty"`
}

// FieldExtract is a label/annotation extraction rule. Key selects a single
// field by exact name; KeyRegex selects every matching field, expanding the
// TagName template with the regex back references.
type FieldExtract struct {
	// TagName is the resource attribute name to add. When empty with Key set,
	// the default is k8s.<from>.labels.<key>. With KeyRegex set, TagName
	// supports regex back references (e.g. k8s.pod.label.$1)
	TagName string `yaml:"tag_name,omitempty"`

	// Key is the exact label or annotation name to read
	Key string `yaml:"key,omitempty"`

	// KeyRegex is a full-length match regex selecting fields by name
	KeyRegex string `yaml:"key_regex,omitempty"`

	// From selects the object to read: pod (default), namespace or node
	From string `yaml:"from,omitempty"`
}

// Association associates a record with a pod
type Association struct {
	// Sources must all match for the association to apply
	Sources []AssociationSource `yaml:"sources"`
}

// AssociationSource identifies where a pod association value comes from
type AssociationSource struct {
	// From is either "resource_attribute" or "connection"
	From string `yaml:"from"`

	// Name is the resource attribute to read when From is resource_attribute
	Name string `yaml:"name,omitempty"`
}

// Filter restricts the pods the processor watches
type Filter struct {
	// NodeFromEnvVar restricts to the node named by this environment variable
	NodeFromEnvVar string `yaml:"node_from_env_var,omitempty"`
}

// NewK8sAttributes creates a new k8sattributes processor with the given name
func NewK8sAttributes(name string) *K8sAttributes {
	return &K8sAttributes{
		id: types.MakeComponentID(string(types.ProcessorTypeK8sAttributes), name),
	}
}

// ID returns the processor identifier in the format "type" or "type/name"
func (p *K8sAttributes) ID() string {
	return p.id
}

// ProcessorType extracts the processor type from the ID
func (p *K8sAttributes) ProcessorType() types.ProcessorType {
	componentType, _ := types.ParseComponentID(p.id)
	return types.ProcessorType(componentType)
}
