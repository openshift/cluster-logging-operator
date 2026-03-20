package common

import (
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
)

// NewApiBatch returns the batch tuning for an output or nil when nothing varies
// from the defaults
func NewApiBatch(t observability.TunableOutput) *sinks.Batch {
	maxBytes := t.GetTuning().MaxWrite
	if maxBytes != nil && !maxBytes.IsZero() {
		return &sinks.Batch{
			MaxBytes: uint64(maxBytes.Value()),
		}
	}
	return nil
}
