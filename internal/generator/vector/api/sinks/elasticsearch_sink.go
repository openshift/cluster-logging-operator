package sinks

type Elasticsearch struct {
	Type       SinkType           `json:"type,omitempty" yaml:"type,omitempty" toml:"type,omitempty"`
	Inputs     []string           `json:"inputs,omitempty" yaml:"inputs,omitempty" toml:"inputs,omitempty"`
	Endpoints  []string           `json:"endpoints,omitempty" yaml:"endpoints,omitempty" toml:"endpoints,omitempty"`
	IdKey      string             `json:"id_key,omitempty" yaml:"id_key,omitempty" toml:"id_key,omitempty"`
	ApiVersion string             `json:"api_version,omitempty" yaml:"api_version,omitempty" toml:"api_version,omitempty"`
	Bulk       *Bulk              `json:"bulk,omitempty" yaml:"bulk,omitempty" toml:"bulk,omitempty"`
	Auth       *ElasticsearchAuth `json:"auth,omitempty" yaml:"auth,omitempty" toml:"auth,omitempty"`
	BaseSink
	Proxy *Proxy `json:"proxy,omitempty" yaml:"proxy,omitempty" toml:"proxy,omitempty"`
}

func NewElasticsearch(url string, init func(s *Elasticsearch), inputs ...string) (s *Elasticsearch) {
	s = &Elasticsearch{
		Type:      SinkTypeElasticsearch,
		Inputs:    inputs,
		Endpoints: []string{url},
	}
	if init != nil {
		init(s)
	}
	return s
}

type ElasticsearchAuth struct {
	Strategy HttpAuthStrategy `json:"strategy,omitempty" yaml:"strategy,omitempty" toml:"strategy,omitempty"`
	HttpAuthBasic
}

type BulkActionType string

const (
	BulkActionCreate BulkActionType = "create"
)

type Bulk struct {
	Index  string         `json:"index,omitempty" yaml:"index,omitempty" toml:"index,omitempty"`
	Action BulkActionType `json:"action,omitempty" yaml:"action,omitempty" toml:"action,omitempty"`
}
