package types

import "strings"

// ParseComponentID extracts the type and optional name from a component ID
// Component IDs follow the format "type" or "type/name"
// Examples:
//   - "file_log" returns ("file_log", "")
//   - "file_log/containers" returns ("file_log", "containers")
//   - "otlphttp/loki" returns ("otlphttp", "loki")
func ParseComponentID(id string) (componentType, name string) {
	if idx := strings.Index(id, "/"); idx != -1 {
		return id[:idx], id[idx+1:]
	}
	return id, ""
}

// MakeComponentID constructs a component ID from a type and optional name
// If name is empty, returns just the type
// If name is provided, returns "type/name"
func MakeComponentID(componentType, name string) string {
	if name == "" {
		return componentType
	}
	return componentType + "/" + name
}
