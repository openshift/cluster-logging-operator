package sinks

import (
	"sort"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

type Elasticsearch struct {
	Type       types.SinkType          `json:"type,omitempty"`
	Inputs     []string                `json:"inputs,omitempty"`
	Endpoints  []string                `json:"endpoints,omitempty"`
	IdKey      string                  `json:"id_key,omitempty"`
	ApiVersion ElasticsearchApiVersion `json:"api_version,omitempty"`
	BaseSink
	Bulk  *Bulk              `json:"bulk,omitempty"`
	Auth  *ElasticsearchAuth `json:"auth,omitempty"`
	Proxy *Proxy             `json:"proxy,omitempty"`
}

func NewElasticsearch(endpoints []string, init func(s *Elasticsearch), inputs ...string) (s *Elasticsearch) {
	sort.Strings(inputs)
	s = &Elasticsearch{
		Type:      types.SinkTypeElasticsearch,
		Inputs:    inputs,
		Endpoints: endpoints,
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
	Strategy HttpAuthStrategy `json:"strategy,omitempty"`
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
	Index  string         `json:"index,omitempty"`
	Action BulkActionType `json:"action,omitempty"`
}
