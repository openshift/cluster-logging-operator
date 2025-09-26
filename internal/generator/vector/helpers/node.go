package helpers

import "strings"

// InputComponent is a vector sink, transformation, source that is
// provided as input to other components
type InputComponent interface {
	// InputIDs are the ids of config elemements to use as input to other components
	InputIDs() []string
}

// ComponentReceiver is a vector component that receives input from another component (e.g. transform, sink)
type ComponentReceiver interface {
	AddInputFrom(n InputComponent)
}

// MakeID given a list of components
func MakeID(parts ...string) string {
	return FormatComponentID(strings.Join(parts, "_"))
}

// MakeIDList given a list of components and return as a single entry array
func MakeIDList(parts ...string) []string {
	return []string{MakeID(parts...)}
}

// MakeRouteInputID appends sourceType to rerouteId for input ids
func MakeRouteInputID(rerouteId, sourceType string) string {
	return strings.ToLower(strings.Join([]string{rerouteId, sourceType}, "."))
}

// MakeInputID for components that logically represent clf.input
func MakeInputID(parts ...string) string {
	parts = append([]string{"input"}, parts...)
	return MakeID(parts...)
}

// MakePipelineID for components that logically represent clf.pipeline (e.g. filters)
func MakePipelineID(parts ...string) string {
	parts = append([]string{"pipeline"}, parts...)
	return MakeID(parts...)
}

// MakeOutPutID for components that logically represent clf.output
func MakeOutputID(parts ...string) string {
	parts = append([]string{"output"}, parts...)
	return MakeID(parts...)
}
