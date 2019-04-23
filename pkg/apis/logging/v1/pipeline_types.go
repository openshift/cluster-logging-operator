package v1

//PipelinesSpec specifies log pipelines from a defined sourct to dest endpoings
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

//PipelineTargetCertificatesSpec specifies secretes for pipelines
type PipelineTargetCertificatesSpec struct {
	SecreteName string `json:"secretName"`
}

//PipelineTargetType defines the type of endpoint that will receive messages
type PipelineTargetType string

const (
	//PipelineTargetTypeElasticsearch configures pipeline to send messages to elasticsearch
	PipelineTargetTypeElasticsearch PipelineTargetType = "elasticsearch"
)
