package v1

//LogSourceType is an explicitly defined log source
type LogSourceType string

const (

	//LogSourceTypeApp are container logs from non-infra structure containers
	LogSourceTypeApp LogSourceType = "logs.app"

	//LogSourceTypeInfra are logs from infra structure containers or node logs
	LogSourceTypeInfra LogSourceType = "logs.infra"
)

//ForwardingSpec specifies log forwarding pipelines from a defined sources to dest outputs
type ForwardingSpec struct {
	DisableDefaultForwarding bool           `json:"disableDefaultForwarding,omitempty"`
	Outputs                  []OutputSpec   `json:"outputs,omitempty"`
	Pipelines                []PipelineSpec `json:"pipelines,omitempty"`
}

//PipelineSpec is the sources spec to named targets
type PipelineSpec struct {
	Name       string        `json:"name,omitempty"`
	SourceType LogSourceType `json:"type,omitempty"`

	//OutputRefs is a list of  the names of outputs defined by forwarding.outputs
	OutputRefs []string `json:"outputRefs,omitempty"`
}

//OutputSpec specifies destination config for log message endpoints
type OutputSpec struct {
	Type     OutputType        `json:"type,omitempty"`
	Name     string            `json:"name,omitempty"`
	Endpoint string            `json:"endpoint,omitempty"`
	Secret   *OutputSecretSpec `json:"secret,omitempty"`
}

//OutputSecretSpec specifies secrets for pipelines
type OutputSecretSpec struct {

	//Name is the name of the secret to use with this output
	Name string `json:"name"`
}

//OutputType defines the type of endpoint that will receive messages
type OutputType string

const (
	//OutputTypeElasticsearch configures pipeline to send messages to elasticsearch
	OutputTypeElasticsearch OutputType = "elasticsearch"

	//OutputTypeForward configures the pipeline to send messages via Fluent's secure forward
	OutputTypeForward OutputType = "forward"
)
