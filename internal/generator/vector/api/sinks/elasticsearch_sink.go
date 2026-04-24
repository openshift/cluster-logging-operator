package sinks

import (
	"sort"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

type Elasticsearch struct {
	Type       types.SinkType          `json:"type,omitempty" yaml:"type,omitempty" toml:"type,omitempty"`
	Inputs     []string                `json:"inputs,omitempty" yaml:"inputs,omitempty" toml:"inputs,omitempty"`
	Endpoints  []string                `json:"endpoints,omitempty" yaml:"endpoints,omitempty" toml:"endpoints,omitempty"`
	IdKey      string                  `json:"id_key,omitempty" yaml:"id_key,omitempty" toml:"id_key,omitempty"`
	ApiVersion ElasticsearchApiVersion `json:"api_version,omitempty" yaml:"api_version,omitempty" toml:"api_version,omitempty"`
	Bulk       *Bulk                   `json:"bulk,omitempty" yaml:"bulk,omitempty" toml:"bulk,omitempty"`
	Auth       *ElasticsearchAuth      `json:"auth,omitempty" yaml:"auth,omitempty" toml:"auth,omitempty"`
	BaseSink
	Proxy *Proxy `json:"proxy,omitempty" yaml:"proxy,omitempty" toml:"proxy,omitempty"`
}

func NewElasticsearch(url string, init func(s *Elasticsearch), inputs ...string) (s *Elasticsearch) {
	sort.Strings(inputs)
	s = &Elasticsearch{
		Type:      types.SinkTypeElasticsearch,
		Inputs:    inputs,
		Endpoints: []string{url},
	}
	if init != nil {
		init(s)
	}
	return s
}

func (s *Elasticsearch) SinkType() types.SinkType {
	return s.Type
}

type ElasticsearchAuth struct {
	Strategy HttpAuthStrategy `json:"strategy,omitempty" yaml:"strategy,omitempty" toml:"strategy,omitempty"`
	HttpAuthBasic
}

type ElasticsearchApiVersion string

const (
	ElasticsearchApiVersion6 ElasticsearchApiVersion = "v6"
	ElasticsearchApiVersion7 ElasticsearchApiVersion = "v7"
	ElasticsearchApiVersion8 ElasticsearchApiVersion = "v8"
)

func (v ElasticsearchApiVersion) Int() int {
	switch v {
	case ElasticsearchApiVersion6:
		return 6
	case ElasticsearchApiVersion7:
		return 7
	case ElasticsearchApiVersion8:
		return 8
	default:
		return 8
	}
}

type BulkActionType string

const (
	BulkActionCreate BulkActionType = "create"
)

type Bulk struct {
	Index  string         `json:"index,omitempty" yaml:"index,omitempty" toml:"index,omitempty"`
	Action BulkActionType `json:"action,omitempty" yaml:"action,omitempty" toml:"action,omitempty"`
}
