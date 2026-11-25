package types

type ProcessorType string

const (
	ProcessorTypeK8sAttributes ProcessorType = "k8sattributes"
	ProcessorTypeResource      ProcessorType = "resource"
)

// Processor is an OpenTelemetry Collector processor for operating on signals
// inside the collector
type Processor interface {
	ID() string
	ProcessorType() ProcessorType
}
