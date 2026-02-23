package common

import (
	"github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
)

const (
	BufferWhenFullBlock      = "block"
	BufferWhenFullDropNewest = "drop_newest"
	minBufferSize            = 268435488
)

// NewApiBuffer returns the buffer tuning for an output or nil when nothing varies
// from the defaults
func NewApiBuffer(t observability.TunableOutput) *sinks.Buffer {
	switch t.GetTuning().DeliveryMode {
	case v1.DeliveryModeAtLeastOnce:
		return &sinks.Buffer{
			WhenFull: sinks.BufferWhenFullBlock,
			Type:     sinks.BufferTypeDisk,
			MaxSize:  minBufferSize,
		}
	case v1.DeliveryModeAtMostOnce:
		return &sinks.Buffer{
			WhenFull: sinks.BufferWhenFullDropNewest,
		}
	}
	return nil
}
