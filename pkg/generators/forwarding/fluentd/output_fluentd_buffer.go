package fluentd

import (
	"fmt"

	loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
)

const (
	// Buffer size defaults
	defaultOverflowAction = "block"

	// Flush buffer defaults
	defaultFlushThreadCount = "2"
	defaultFlushMode        = "interval"
	defaultFlushInterval    = "1s"

	// Retry buffer to output defaults
	defaultRetryWait        = "1s"
	defaultRetryType        = "exponential_backoff"
	defaultRetryMaxInterval = "300s"

	// Output fluentdForward default
	fluentdForwardChunkLimitSize = "8m"
	fluentdForwardOverflowAction = "block"
	fluentdForwardFlushInterval  = "5s"
)

func (olc *outputLabelConf) ChunkLimitSize() string {
	if hasBufferConfig(olc.forwarder) {
		return string(olc.forwarder.Fluentd.Buffer.ChunkLimitSize)
	}

	return ""
}

func (olc *outputLabelConf) TotalLimitSize() string {
	if hasBufferConfig(olc.forwarder) {
		return string(olc.forwarder.Fluentd.Buffer.TotalLimitSize)
	}

	return ""
}

func (olc *outputLabelConf) OverflowAction() string {
	if hasBufferConfig(olc.forwarder) {
		return string(olc.forwarder.Fluentd.Buffer.OverflowAction)
	}

	switch olc.Target.Type {
	case loggingv1.OutputTypeFluentdForward:
		return fluentdForwardOverflowAction
	default:
		return defaultOverflowAction
	}
}

func (olc *outputLabelConf) FlushThreadCount() string {
	if hasBufferConfig(olc.forwarder) {
		return fmt.Sprintf("%d", olc.forwarder.Fluentd.Buffer.FlushThreadCount)
	}

	return defaultFlushThreadCount
}

func (olc *outputLabelConf) FlushMode() string {
	if hasBufferConfig(olc.forwarder) {
		return string(olc.forwarder.Fluentd.Buffer.FlushMode)
	}

	return defaultFlushMode
}

func (olc *outputLabelConf) FlushInterval() string {
	if hasBufferConfig(olc.forwarder) {
		return string(olc.forwarder.Fluentd.Buffer.FlushInterval)
	}

	switch olc.Target.Type {
	case loggingv1.OutputTypeFluentdForward:
		return fluentdForwardFlushInterval
	default:
		return defaultFlushInterval
	}
}

func (olc *outputLabelConf) RetryWait() string {
	if hasBufferConfig(olc.forwarder) {
		return string(olc.forwarder.Fluentd.Buffer.RetryWait)
	}

	return defaultRetryWait
}

func (olc *outputLabelConf) RetryType() string {
	if hasBufferConfig(olc.forwarder) {
		return string(olc.forwarder.Fluentd.Buffer.RetryType)
	}

	return defaultRetryType
}

func (olc *outputLabelConf) RetryMaxInterval() string {
	if hasBufferConfig(olc.forwarder) {
		return string(olc.forwarder.Fluentd.Buffer.RetryMaxInterval)
	}

	return defaultRetryMaxInterval
}

func hasBufferConfig(config *loggingv1.ForwarderSpec) bool {
	return config != nil && config.Fluentd != nil && config.Fluentd.Buffer != nil
}
