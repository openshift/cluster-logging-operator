package output

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
)

func (o Output) VisitSink(s common.SinkConfig) {
	if o.spec.Tuning != nil {
		comp := o.spec.Tuning.Compression
		if comp != "" && comp != "none" {
			s.SetCompression(comp)
		}
	}
}

// VisitAcknowledgements enables acknowledgements when an output is tuned for AtLeastOnce delivery
func (o Output) VisitAcknowledgements(a common.Acknowledgments) common.Acknowledgments {
	a.Enabled = o.spec.Tuning != nil && o.spec.Tuning.Delivery == logging.OutputDeliveryModeAtLeastOnce
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
		if o.spec.Tuning.MinRetryDuration != nil && o.spec.Tuning.MinRetryDuration.Seconds() > 0 {
			r.RetryInitialBackoffSec.Value = o.spec.Tuning.MinRetryDuration.Seconds()
		}
		if o.spec.Tuning.MaxRetryDuration != nil && o.spec.Tuning.MaxRetryDuration.Seconds() > 0 {
			r.RetryMaxDurationSec.Value = o.spec.Tuning.MaxRetryDuration.Seconds()
		}
	} else {
		r.RetryAttempts.Value = 17
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
		case logging.OutputDeliveryModeAtMostOnce:
			b.WhenFull.Value = common.BufferWhenFullDropNewest
		}
	} else {
		b.WhenFull.Value = common.BufferWhenFullDropNewest
	}
	return b
}
