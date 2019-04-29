package v1

type PipelineSourceType string

const (
	PipelineSourceTypeLogsApp   PipelineSourceType = "logs.app"
	PipelineSourceTypeLogsInfra PipelineSourceType = "logs.infra"
)

func (p *PipelinesSpec) Map() map[string]*PipelineTargetsSpec {
	m := make(map[string]*PipelineTargetsSpec)
	m[string(PipelineSourceTypeLogsApp)] = p.LogsApp
	m[string(PipelineSourceTypeLogsInfra)] = p.LogsInfra
	return m
}

//PipelinesSpec specifies log pipelines from a defined sourct to dest endpoints
type PipelinesSpec struct {
	LogsApp   *PipelineTargetsSpec `json:"logs.app,omitempty"`
	LogsInfra *PipelineTargetsSpec `json:"logs.infra,omitempty"`
}

//PipelineTargetsSpec is the list of targets for a given source
type PipelineTargetsSpec struct {
	Targets []PipelineTargetSpec `json:"targets,omitempty"`
}

//PipelineTargetSpec specifies destination config for log message endpoints
type PipelineTargetSpec struct {
	Type         PipelineTargetType              `json:"type"`
	Endpoint     string                          `json:"endpoint,omitempty"`
	Certificates *PipelineTargetCertificatesSpec `json:"certificates,omitempty"`
}

//PipelineTargetCertificatesSpec specifies secrets for pipelines
type PipelineTargetCertificatesSpec struct {
	SecretName string `json:"secretName"`
}

//PipelineTargetType defines the type of endpoint that will receive messages
type PipelineTargetType string

const (
	//PipelineTargetTypeElasticsearch configures pipeline to send messages to elasticsearch
	PipelineTargetTypeElasticsearch PipelineTargetType = "elasticsearch"
)
