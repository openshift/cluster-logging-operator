package api

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/exporters"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/types"
	"gopkg.in/yaml.v3"
)

// Exporters is a set of exporters by their id.
// In OpenTelemetry Collector configuration, exporter IDs follow the pattern "type" or "type/name"
// where the type determines the exporter implementation.
type Exporters map[string]types.Exporter

func (exporterMap *Exporters) Add(id string, exporter types.Exporter) {
	(*exporterMap)[id] = exporter
}

func (exporterMap *Exporters) Merge(exporters Exporters) {
	for _, e := range exporters {
		(*exporterMap)[e.ID()] = e
	}
}

func (exporterMap *Exporters) UnmarshalYAML(value *yaml.Node) error {
	if *exporterMap == nil {
		*exporterMap = make(Exporters)
	}
	return unmarshalComponentMap(value, "exporters", decodeExporter, func(id string, exporter types.Exporter) {
		(*exporterMap)[id] = exporter
	})
}

// decodeExporter decodes a YAML node into a specific exporter type
func decodeExporter(exporterType, exporterID string, node *yaml.Node) (types.Exporter, error) {
	switch types.ExporterType(exporterType) {
	case types.ExporterTypeOtlpHttp:
		var e exporters.OtlpHttp
		if err := node.Decode(&e); err != nil {
			return nil, fmt.Errorf("failed to unmarshal otlphttp exporter %s: %w", exporterID, err)
		}
		return &e, nil
	default:
		return nil, fmt.Errorf("unknown exporter type %s for exporter %s", exporterType, exporterID)
	}
}
