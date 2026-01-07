package api

import (
	"errors"
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	gotoml "github.com/pelletier/go-toml"
)

type Sinks map[string]types.Sink

func (sinkMap *Sinks) Add(id string, sink types.Sink) {
	(*sinkMap)[id] = sink
}
func (sinkMap *Sinks) Merge(sinks Sinks) {
	for id, s := range sinks {
		(*sinkMap)[id] = s
	}
}

func (sinkMap *Sinks) UnmarshalTOML(data interface{}) (err error) {
	entries, castable := data.(map[string]interface{})
	if !castable {
		return fmt.Errorf("sinks data can not be converted to a map: %v", data)
	}
	if *sinkMap == nil {
		*sinkMap = make(Sinks)
	}
	for id, entry := range entries {
		rawSource, ok := entry.(map[string]interface{})
		if !ok {
			return fmt.Errorf("raw entry for sink %q can not be converted to a map %v", entry, rawSource)
		}
		tree, mapErr := gotoml.TreeFromMap(rawSource)
		if mapErr != nil {
			return errors.Join(fmt.Errorf("unable to initialize tree from sink %q", id), mapErr)
		}
		var typeExtractor struct {
			Type types.SinkType `yaml:"type"`
		}
		var sink types.Sink
		if err = tree.Unmarshal(&typeExtractor); err != nil {
			return errors.Join(fmt.Errorf("unable to unmarshal sink %q from %v to determine type", id, rawSource), err)
		}
		switch typeExtractor.Type {
		case types.SinkTypeAwsCloudwatchLogs:
			var s sinks.AwsCloudwatchLogs
			if err = tree.Unmarshal(&s); err != nil {
				return fmt.Errorf("failed to unmarshal sink %s: %w", id, err)
			}
			sink = &s
		case types.SinkTypeAwsS3:
			var s sinks.AwsS3
			if err = tree.Unmarshal(&s); err != nil {
				return fmt.Errorf("failed to unmarshal sink %s: %w", id, err)
			}
			sink = &s
		case types.SinkTypeAzureMonitorLogs:
			var s sinks.AzureMonitorLogs
			if err = tree.Unmarshal(&s); err != nil {
				return fmt.Errorf("failed to unmarshal sink %s: %w", id, err)
			}
			sink = &s
		case types.SinkTypeElasticsearch:
			var s sinks.Elasticsearch
			if err = tree.Unmarshal(&s); err != nil {
				return fmt.Errorf("failed to unmarshal sink %s: %w", id, err)
			}
			sink = &s
		case types.SinkTypeGcpStackdriverLogs:
			var s sinks.GcpStackdriverLogs
			if err = tree.Unmarshal(&s); err != nil {
				return fmt.Errorf("failed to unmarshal sink %s: %w", id, err)
			}
			sink = &s
		case types.SinkTypeHttp:
			var s sinks.Http
			if err = tree.Unmarshal(&s); err != nil {
				return fmt.Errorf("failed to unmarshal sink %s: %w", id, err)
			}
			sink = &s
		case types.SinkTypeLoki:
			var s sinks.Loki
			if err = tree.Unmarshal(&s); err != nil {
				return fmt.Errorf("failed to unmarshal sink %s: %w", id, err)
			}
			sink = &s
		case types.SinkTypeKafka:
			var s sinks.Kafka
			if err = tree.Unmarshal(&s); err != nil {
				return fmt.Errorf("failed to unmarshal sink %s: %w", id, err)
			}
			sink = &s
		case types.SinkTypeOpenTelemetry:
			var s sinks.OpenTelemetry
			if err = tree.Unmarshal(&s); err != nil {
				return fmt.Errorf("failed to unmarshal sink %s: %w", id, err)
			}
			sink = &s
		case types.SinkTypePrometheusExporter:
			var s sinks.PrometheusExporter
			if err = tree.Unmarshal(&s); err != nil {
				return fmt.Errorf("failed to unmarshal sink %s: %w", id, err)
			}
			sink = &s
		case types.SinkTypeSocket:
			var s sinks.Socket
			if err = tree.Unmarshal(&s); err != nil {
				return fmt.Errorf("failed to unmarshal sink %s: %w", id, err)
			}
			sink = &s
		case types.SinkTypeSpunkHecLogs:
			var s sinks.SplunkHecLogs
			if err = tree.Unmarshal(&s); err != nil {
				return fmt.Errorf("failed to unmarshal sink %s: %w", id, err)
			}
			sink = &s
		default:
			return fmt.Errorf("unknown sink type %s for sink %s", typeExtractor.Type, id)
		}

		(*sinkMap)[id] = sink
	}
	return nil
}
