package transforms

import (
	"sort"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

// DetectExceptions is a transform to detect exception stack traces that is only provided by Red Hat's distribution of
// vector
type DetectExceptions struct {
	Type                     types.TransformType `json:"type"`
	Inputs                   []string            `json:"inputs"`
	Languages                []LanguageType      `json:"languages,omitempty"`
	GroupBy                  []string            `json:"group_by,omitempty"`
	ExpireAfterMs            uint                `json:"expire_after_ms,omitempty"`
	MultilineFlushIntervalMs uint                `json:"multiline_flush_interval_ms,omitempty"`
	MessageKey               string              `json:"message_key,omitempty"`
}

type LanguageType string

const (
	LanguageTypeAll LanguageType = "All"
)

func NewDetectExceptions(languages []LanguageType, inputs ...string) *DetectExceptions {
	sort.Strings(inputs)
	return &DetectExceptions{
		Type:      types.TransformTypeDetectExceptions,
		Inputs:    inputs,
		Languages: languages,
	}
}

func (t *DetectExceptions) TransformType() types.TransformType {
	return t.Type
}
