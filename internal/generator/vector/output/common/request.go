package common

import (
	"time"

	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
)

func NewApiRequest(o observability.TunableOutput) (r *sinks.Request) {
	var duration time.Duration
	t := o.GetTuning()
	if t.MinRetryDuration != nil && t.MinRetryDuration.Seconds() > 0 {
		r = &sinks.Request{}
		// time.Duration is default nanosecond. Convert to seconds first.
		duration = *t.MinRetryDuration * time.Second
		r.RetryInitialBackoffSecs = uint(duration.Seconds())
	}
	if t.MaxRetryDuration != nil && t.MaxRetryDuration.Seconds() > 0 {
		if r == nil {
			r = &sinks.Request{}
		}
		duration = *t.MaxRetryDuration * time.Second
		r.RetryMaxDurationSec = uint(duration.Seconds())
	}
	return r
}
