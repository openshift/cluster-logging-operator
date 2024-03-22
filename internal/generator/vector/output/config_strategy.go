package output

import (
	"time"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
)

const (
	minBufferSize   = 268435488
	buffertTypeDisk = "disk"
)

func (o Output) VisitSink(s common.SinkConfig) {
	if o.spec.Tuning != nil {
		comp := o.spec.Tuning.Compression
		if comp != "" && comp != "none" {
			s.SetCompression(comp)
		}
	}
}

func (o Output) VisitAcknowledgements(a common.Acknowledgments) common.Acknowledgments {
	return a
}

func (o Output) VisitBatch(b common.Batch) common.Batch {
	if o.spec.Tuning != nil && o.spec.Tuning.MaxWrite != nil && !o.spec.Tuning.MaxWrite.IsZero() {
		b.MaxBytes.Value = o.spec.Tuning.MaxWrite.Value()
	}
	return b
}

func (o Output) VisitRequest(r common.Request) common.Request {
	if o.spec.Tuning != nil {
		var duration time.Duration
		if o.spec.Tuning.MinRetryDuration != nil && o.spec.Tuning.MinRetryDuration.Seconds() > 0 {
			// time.Duration is default nanosecond. Convert to seconds first.
			duration = *o.spec.Tuning.MinRetryDuration * time.Second
			r.RetryInitialBackoffSec.Value = duration.Seconds()
		}
		if o.spec.Tuning.MaxRetryDuration != nil && o.spec.Tuning.MaxRetryDuration.Seconds() > 0 {
			duration = *o.spec.Tuning.MaxRetryDuration * time.Second
			r.RetryMaxDurationSec.Value = duration.Seconds()
		}
	}

	return r
}

// VisitBuffer modifies the buffer behavior depending upon the value
// of the tuning.Delivery mode
func (o Output) VisitBuffer(b common.Buffer) common.Buffer {
	if o.spec.Tuning != nil {
		switch o.spec.Tuning.Delivery {
		case logging.OutputDeliveryModeAtLeastOnce:
			b.WhenFull.Value = common.BufferWhenFullBlock
			b.Type.Value = buffertTypeDisk
			b.MaxSize.Value = minBufferSize
		case logging.OutputDeliveryModeAtMostOnce:
			b.WhenFull.Value = common.BufferWhenFullDropNewest
		}
	}
	return b
}
