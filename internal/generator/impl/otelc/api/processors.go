package api

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/processors"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/types"
	"gopkg.in/yaml.v3"
)

// Processors is a set of processors by their id.
// In OpenTelemetry Collector configuration, processor IDs follow the pattern "type" or "type/name"
// where the type determines the processor implementation.
type Processors map[string]types.Processor

func (processorMap *Processors) Add(id string, processor types.Processor) {
	(*processorMap)[id] = processor
}

func (processorMap *Processors) Merge(processors Processors) {
	for _, p := range processors {
		(*processorMap)[p.ID()] = p
	}
}

func (processorMap *Processors) UnmarshalYAML(value *yaml.Node) error {
	if *processorMap == nil {
		*processorMap = make(Processors)
	}
	return unmarshalComponentMap(value, "processors", decodeProcessor, func(id string, processor types.Processor) {
		(*processorMap)[id] = processor
	})
}

// decodeProcessor decodes a YAML node into a specific processor type
func decodeProcessor(processorType, processorID string, node *yaml.Node) (types.Processor, error) {
	switch types.ProcessorType(processorType) {
	case types.ProcessorTypeK8sAttributes:
		var p processors.K8sAttributes
		if err := node.Decode(&p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal k8sattributes processor %s: %w", processorID, err)
		}
		return &p, nil
	case types.ProcessorTypeResource:
		var p processors.Resource
		if err := node.Decode(&p); err != nil {
			return nil, fmt.Errorf("failed to unmarshal resource processor %s: %w", processorID, err)
		}
		return &p, nil
	default:
		return nil, fmt.Errorf("unknown processor type %s for processor %s", processorType, processorID)
	}
}
