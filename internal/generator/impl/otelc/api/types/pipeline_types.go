package types

type PipelineType string

const (
	PipelineTypeLogs = "logs"
)

type Pipeline interface {
	ID() string
	PipelineType() PipelineType
}
