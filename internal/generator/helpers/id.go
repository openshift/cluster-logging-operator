package helpers

import "strings"

var (
	idReplacer = strings.NewReplacer(" ", "_", "-", "_", ".", "_")
)

// MakeID given a list of components
func MakeID(parts ...string) string {
	return strings.ToLower(idReplacer.Replace(strings.Join(parts, "_")))
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
