package output

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
)

const (
	minBufferSize   = 268435488
	buffertTypeDisk = "disk"
)

func (o Output) VisitSink(s common.SinkConfig) {
	if o.tuning.Compression != "" {
		s.SetCompression(o.tuning.Compression)
	}
}

func (o Output) VisitAcknowledgements(a common.Acknowledgments) common.Acknowledgments {
	return a
}

func (o Output) VisitBatch(b common.Batch) common.Batch {
	if o.tuning.MaxWrite != nil && !o.tuning.MaxWrite.IsZero() {
		b.MaxBytes.Value = o.tuning.MaxWrite.Value()
	}
	return b
}

func (o Output) VisitRequest(r common.Request) common.Request {
	if o.tuning.MinRetryDuration != nil {
		duration := o.tuning.MinRetryDuration.Duration
		if duration.Seconds() > 0 {
			r.RetryInitialBackoffSec.Value = duration.Seconds()
		}
	}
	if o.tuning.MaxRetryDuration != nil {
		duration := o.tuning.MaxRetryDuration.Duration
		if duration.Seconds() > 0 {
			r.RetryMaxDurationSec.Value = duration.Seconds()
		}
	}

	return r
}

// VisitBuffer modifies the buffer behavior depending upon the value
// of the tuning.Delivery mode
func (o Output) VisitBuffer(b common.Buffer) common.Buffer {
	switch o.tuning.DeliveryMode {
	case obs.DeliveryModeAtLeastOnce:
		b.WhenFull.Value = common.BufferWhenFullBlock
		b.Type.Value = buffertTypeDisk
		b.MaxSize.Value = minBufferSize
	case obs.DeliveryModeAtMostOnce:
		b.WhenFull.Value = common.BufferWhenFullDropNewest
	}
	return b
}
