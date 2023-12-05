package helpers

import "strings"

// Component is a vector sink, transformation, source
type Component interface {
	// InputIDs are the ids of config elemements to use as input to other components
	InputIDs() []string
}

// ComponentReceiver is a vector component that receives input from another component (e.g. transform, sink)
type ComponentReceiver interface {
	AddInputFrom(n Component)
}

// MakeID given a list of components
func MakeID(parts ...string) string {
	return FormatComponentID(strings.Join(parts, "_"))
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
