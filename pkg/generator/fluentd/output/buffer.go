package output

import (
	"fmt"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/pkg/generator"
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
	defaultRetryMaxInterval = "60s"
	defaultRetryTimeout     = "60m"
	defaultBufferQueueLimit = "32"
	defaultTotalLimitSize   = "8589934592" // 0x200000000, 8GB
	defaultBufferSizeLimit  = "8m"

	// Output fluentdForward default
	fluentdForwardOverflowAction = "block"
	fluentdForwardFlushInterval  = "5s"
)

var NOKEYS = []string{}

func Buffer(bufkeys []string, bufspec *logging.FluentdBufferSpec, bufpath string, os *logging.OutputSpec) []Element {
	return []Element{
		MakeBuffer(bufkeys, bufspec, bufpath, os),
	}
}

func MakeBuffer(bufkeys []string, bufspec *logging.FluentdBufferSpec, bufpath string, os *logging.OutputSpec) BufferConfig {
	return BufferConfig{
		BufferKeys:           bufkeys,
		BufferPath:           BufferPath(bufpath),
		FlushMode:            FlushMode(bufspec),
		FlushThreadCount:     FlushThreadCount(bufspec),
		FlushInterval:        FlushInterval(os, bufspec),
		RetryType:            RetryType(bufspec),
		RetryWait:            RetryWait(bufspec),
		RetryMaxInterval:     RetryMaxInterval(bufspec),
		RetryTimeout:         RetryTimeout(bufspec),
		QueuedChunkLimitSize: QueuedChunkLimitSize(bufspec),
		TotalLimitSize:       TotalLimitSize(bufspec),
		ChunkLimitSize:       ChunkLimitSize(bufspec),
		OverflowAction:       OverflowAction(os, bufspec),
	}
}

func BufferPath(bufpath string) string {
	return fmt.Sprintf("/var/lib/fluentd/%s", bufpath)
}

func ChunkLimitSize(bufspec *logging.FluentdBufferSpec) string {
	if bufspec != nil && bufspec.ChunkLimitSize != "" {
		return string(bufspec.ChunkLimitSize)
	}
	return FromEnv("BUFFER_SIZE_LIMIT", defaultBufferSizeLimit)
}

func QueuedChunkLimitSize(bufspec *logging.FluentdBufferSpec) string {
	return FromEnv("BUFFER_QUEUE_LIMIT", defaultBufferQueueLimit)
}

func TotalLimitSize(bufspec *logging.FluentdBufferSpec) string {
	if bufspec != nil && bufspec.TotalLimitSize != "" {
		return string(bufspec.TotalLimitSize)
	}
	return FromEnv("TOTAL_LIMIT_SIZE", defaultTotalLimitSize)
}

func OverflowAction(os *logging.OutputSpec, bufspec *logging.FluentdBufferSpec) string {
	if bufspec != nil {
		oa := string(bufspec.OverflowAction)

		if oa != "" {
			return oa
		}
	}

	switch os.Type {
	case logging.OutputTypeFluentdForward:
		return fluentdForwardOverflowAction
	default:
		return defaultOverflowAction
	}
}

func FlushThreadCount(bufspec *logging.FluentdBufferSpec) string {
	if bufspec != nil {
		ftc := bufspec.FlushThreadCount

		if ftc > 0 {
			return fmt.Sprintf("%d", ftc)
		}
	}
	return defaultFlushThreadCount
}

func FlushMode(bufspec *logging.FluentdBufferSpec) string {
	if bufspec != nil {
		fm := string(bufspec.FlushMode)

		if fm != "" {
			return fm
		}
	}
	return defaultFlushMode
}

func FlushInterval(os *logging.OutputSpec, bufspec *logging.FluentdBufferSpec) string {
	if bufspec != nil {
		fi := string(bufspec.FlushInterval)

		if fi != "" {
			return fi
		}
	}

	switch os.Type {
	case logging.OutputTypeFluentdForward:
		return fluentdForwardFlushInterval
	default:
		return defaultFlushInterval
	}
}

func RetryWait(bufspec *logging.FluentdBufferSpec) string {
	if bufspec != nil {
		rw := string(bufspec.RetryWait)

		if rw != "" {
			return rw
		}
	}
	return defaultRetryWait
}

func RetryType(bufspec *logging.FluentdBufferSpec) string {
	if bufspec != nil {
		rt := string(bufspec.RetryType)

		if rt != "" {
			return rt
		}
	}
	return defaultRetryType
}

func RetryMaxInterval(bufspec *logging.FluentdBufferSpec) string {
	if bufspec != nil {
		rmi := string(bufspec.RetryMaxInterval)

		if rmi != "" {
			return rmi
		}
	}
	return defaultRetryMaxInterval
}

func RetryTimeout(bufspec *logging.FluentdBufferSpec) string {
	value := defaultRetryTimeout
	if bufspec != nil {
		if string(bufspec.RetryTimeout) != "" {
			value = string(bufspec.RetryTimeout)
		}
	}
	return value
}

func FromEnv(env string, defaultVal string) string {
	return fmt.Sprintf("\"#{ENV['%s'] || '%s'}\"", env, defaultVal)
}
