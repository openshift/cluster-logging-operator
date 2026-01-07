package transforms

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

type Remap struct {
	Type types.TransformType `json:"type" yaml:"type" toml:"type"`

	// Inputs is the IDs of the components feeding into this component
	Inputs []string `json:"inputs" yaml:"inputs" toml:"inputs"`

	// Source is the VRL script used for the remap transformation
	Source VrlString `json:"source" yaml:"source" toml:"source" multiline:"true" literal:"true"`
}

func NewRemap(source string, inputs ...string) *Remap {
	sort.Strings(inputs)
	return &Remap{
		Type:   types.TransformTypeRemap,
		Inputs: inputs,
		Source: VrlString(normalize(source)),
	}
}

func (t *Remap) TransformType() types.TransformType {
	return t.Type
}

type VrlString string

func (s *VrlString) UnmarshalTOML(data interface{}) (err error) {
	raw, castable := data.(string)
	if !castable {
		return fmt.Errorf("data can not be converted to a string: %v", data)
	}
	*s = VrlString(normalize(raw))
	return nil
}

// normalize sanitizes the VRL string for beginning tabs, spaces and new line characters
func normalize(s string) string {
	chunks := strings.Split(strings.TrimSpace(s), "\n")
	indentCount := 0
	totChunks := len(chunks)
	buffer := bytes.NewBufferString("")
	for i, chunk := range chunks {
		raw := strings.TrimSpace(chunk)
		if strings.HasPrefix(raw, "}") || strings.HasPrefix(raw, ")") || strings.HasPrefix(raw, "]") {
			indentCount = indentCount - 2
			if indentCount < 0 {
				indentCount = 0
			}
		}
		writeLine := strings.TrimSpace(raw) != ""
		if writeLine {
			buffer.WriteString(indent(indentCount, raw))
		}
		if strings.HasSuffix(raw, "{") || strings.HasSuffix(raw, "(") || strings.HasSuffix(raw, "[") {
			indentCount = indentCount + 2
		}
		if writeLine && totChunks > 1 && i < totChunks-1 {
			buffer.WriteString("\n")
		}
	}
	return buffer.String()
}

func indent(count int, content string) string {
	return strings.Repeat(" ", count) + content
}
