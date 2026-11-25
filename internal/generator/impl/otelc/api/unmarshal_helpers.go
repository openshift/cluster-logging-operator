package api

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// componentDecoder is a function that decodes a YAML node into a component of the specified type
// It returns the decoded component or an error
type componentDecoder[T any] func(componentType, componentID string, node *yaml.Node) (T, error)

// componentSetter is a function that stores a decoded component with the given ID
type componentSetter[T any] func(id string, component T)

// unmarshalComponentMap is a generic helper for unmarshaling YAML maps of OpenTelemetry Collector components
// (receivers, exporters, processors, etc.) which follow the pattern "type" or "type/name"
func unmarshalComponentMap[T any](
	value *yaml.Node,
	componentName string,
	decoder componentDecoder[T],
	setter componentSetter[T],
) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("%s data must be a mapping, got: %v", componentName, value.Kind)
	}

	// Parse the mapping to extract entries
	entries := make(map[string]*yaml.Node)
	for i := 0; i < len(value.Content); i += 2 {
		key := value.Content[i].Value
		val := value.Content[i+1]
		entries[key] = val
	}

	// Decode each component
	for id, entry := range entries {
		// Extract component type from the ID (e.g., "file_log" or "file_log/my-instance")
		componentType := id
		if slashIdx := strings.Index(id, "/"); slashIdx != -1 {
			componentType = id[:slashIdx]
		}

		// Use the decoder to create the specific component type
		component, err := decoder(componentType, id, entry)
		if err != nil {
			return err
		}

		// Use the setter to store the component
		setter(id, component)
	}

	return nil
}
