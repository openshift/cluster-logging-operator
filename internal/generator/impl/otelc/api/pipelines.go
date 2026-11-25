package api

import "github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/types"

type Pipelines map[string]types.Pipeline

func (p *Pipelines) Add(pipeline types.Pipeline) {
	(*p)[pipeline.ID()] = pipeline
}

type Pipeline struct {
	id         string
	Receivers  []string `yaml:"receivers,omitempty"`
	Processors []string `yaml:"processors,omitempty"`
	Exporters  []string `yaml:"exporters,omitempty"`
}

func NewLogsPipeline(name string) *Pipeline {
	return &Pipeline{
		id:        types.MakeComponentID(string(types.PipelineTypeLogs), name),
		Receivers: []string{},
		Exporters: []string{},
	}
}

func (lp *Pipeline) PipelineType() types.PipelineType {
	componentType, _ := types.ParseComponentID(lp.id)
	return types.PipelineType(componentType)
}

func (lp *Pipeline) ID() string {
	return lp.id
}

func (lp *Pipeline) AddReceiver(receiver types.Receiver) {
	lp.Receivers = append(lp.Receivers, receiver.ID())
}
func (lp *Pipeline) AddProcessor(processor types.Processor) {
	lp.Processors = append(lp.Processors, processor.ID())
}

func (lp *Pipeline) AddExporter(exporter types.Exporter) {
	lp.Exporters = append(lp.Exporters, exporter.ID())
}
