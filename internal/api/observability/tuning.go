package observability

import (
	"time"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
)

const (
	minBufferSize = 268435488
)

type Tuning struct {
	obs.BaseOutputTuningSpec
	Compression string
}

func InitBuffer(t obs.BaseOutputTuningSpec, b *sinks.Buffer) *sinks.Buffer {
	switch t.DeliveryMode {
	case obs.DeliveryModeAtLeastOnce:
		b.WhenFull = sinks.BufferWhenFullBlock
		b.Type = sinks.BufferTypeDisk
		b.MaxSize = minBufferSize
	case obs.DeliveryModeAtMostOnce:
		b.WhenFull = sinks.BufferWhenFullDropNewest
	}
	return b
}

func InitRequest(t obs.BaseOutputTuningSpec, r *sinks.Request) *sinks.Request {
	var duration time.Duration
	if t.MinRetryDuration != nil && t.MinRetryDuration.Seconds() > 0 {
		// time.Duration is default nanosecond. Convert to seconds first.
		duration = *t.MinRetryDuration * time.Second
		r.RetryInitialBackoffSecs = uint(duration.Seconds())
	}
	if t.MaxRetryDuration != nil && t.MaxRetryDuration.Seconds() > 0 {
		duration = *t.MaxRetryDuration * time.Second
		r.RetryMaxDurationSec = uint(duration.Seconds())
	}

	return r
}

func NewTuning(spec obs.OutputSpec) Tuning {
	t := Tuning{}
	switch spec.Type {
	case obs.OutputTypeAzureMonitor:
		if spec.AzureMonitor != nil && spec.AzureMonitor.Tuning != nil {
			t.BaseOutputTuningSpec = *spec.AzureMonitor.Tuning
		}
	case obs.OutputTypeGoogleCloudLogging:
		if spec.GoogleCloudLogging != nil && spec.GoogleCloudLogging.Tuning != nil {
			t.BaseOutputTuningSpec = spec.GoogleCloudLogging.Tuning.BaseOutputTuningSpec
		}
	case obs.OutputTypeCloudwatch:
		if spec.Cloudwatch != nil && spec.Cloudwatch.Tuning != nil {
			t.BaseOutputTuningSpec = spec.Cloudwatch.Tuning.BaseOutputTuningSpec
			t.Compression = spec.Cloudwatch.Tuning.Compression
		}
	case obs.OutputTypeS3:
		if spec.S3 != nil && spec.S3.Tuning != nil {
			t.BaseOutputTuningSpec = spec.S3.Tuning.BaseOutputTuningSpec
			t.Compression = spec.S3.Tuning.Compression
		}
	case obs.OutputTypeElasticsearch:
		if spec.Elasticsearch != nil && spec.Elasticsearch.Tuning != nil {
			t.BaseOutputTuningSpec = spec.Elasticsearch.Tuning.BaseOutputTuningSpec
			t.Compression = spec.Elasticsearch.Tuning.Compression
		}
	case obs.OutputTypeHTTP:
		if spec.HTTP != nil && spec.HTTP.Tuning != nil {
			t.BaseOutputTuningSpec = spec.HTTP.Tuning.BaseOutputTuningSpec
			t.Compression = spec.HTTP.Tuning.Compression
		}
	case obs.OutputTypeOTLP:
		if spec.OTLP != nil && spec.OTLP.Tuning != nil {
			t.BaseOutputTuningSpec = spec.OTLP.Tuning.BaseOutputTuningSpec // TODO: test
			t.Compression = spec.OTLP.Tuning.Compression
		}
	case obs.OutputTypeKafka:
		if spec.Kafka != nil && spec.Kafka.Tuning != nil {
			t.DeliveryMode = spec.Kafka.Tuning.DeliveryMode
			t.MaxWrite = spec.Kafka.Tuning.MaxWrite
			t.Compression = spec.Kafka.Tuning.Compression
		}
	case obs.OutputTypeLoki:
		if spec.Loki != nil && spec.Loki.Tuning != nil {
			t.BaseOutputTuningSpec = spec.Loki.Tuning.BaseOutputTuningSpec
			t.Compression = spec.Loki.Tuning.Compression
		}
	case obs.OutputTypeLokiStack:
		if spec.LokiStack != nil && spec.LokiStack.Tuning != nil {
			t.BaseOutputTuningSpec = spec.LokiStack.Tuning.BaseOutputTuningSpec
			t.Compression = spec.LokiStack.Tuning.Compression
		}
	case obs.OutputTypeSplunk:
		if spec.Splunk != nil && spec.Splunk.Tuning != nil {
			t.BaseOutputTuningSpec = spec.Splunk.Tuning.BaseOutputTuningSpec
			t.Compression = spec.Splunk.Tuning.Compression
		}
	case obs.OutputTypeSyslog:
		if spec.Syslog != nil && spec.Syslog.Tuning != nil {
			t.DeliveryMode = spec.Syslog.Tuning.DeliveryMode
		}
	}
	return t
}
