package observability

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

type Tuning struct {
	obs.BaseOutputTuningSpec
	Compression string
}

func NewTuning(spec obs.OutputSpec) Tuning {
	t := Tuning{}
	switch spec.Type {
	case obs.OutputTypeAzureMonitor:
		if spec.AzureMonitor != nil && spec.AzureMonitor.Tuning != nil {
			t.BaseOutputTuningSpec = *spec.AzureMonitor.Tuning
		}
	case obs.OutputTypeCloudwatch:
		if spec.Cloudwatch != nil && spec.Cloudwatch.Tuning != nil {
			t.BaseOutputTuningSpec = spec.Cloudwatch.Tuning.BaseOutputTuningSpec
			t.Compression = spec.Cloudwatch.Tuning.Compression
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
	case obs.OutputTypeKafka:
		if spec.Kafka != nil && spec.Kafka.Tuning != nil {
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
		}
	}
	return t
}
