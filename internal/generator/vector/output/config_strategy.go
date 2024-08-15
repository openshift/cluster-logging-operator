package output

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"time"
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
	var duration time.Duration
	if o.tuning.MinRetryDuration != nil && o.tuning.MinRetryDuration.Seconds() > 0 {
		// time.Duration is default nanosecond. Convert to seconds first.
		duration = *o.tuning.MinRetryDuration * time.Second
		r.RetryInitialBackoffSec.Value = duration.Seconds()
	}
	if o.tuning.MaxRetryDuration != nil && o.tuning.MaxRetryDuration.Seconds() > 0 {
		duration = *o.tuning.MaxRetryDuration * time.Second
		r.RetryMaxDurationSec.Value = duration.Seconds()
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
