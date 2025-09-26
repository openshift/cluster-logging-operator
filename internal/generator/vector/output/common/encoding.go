package common

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

const (
	CodecJSON              = sinks.CodecJSON
	TimeStampFormatRFC3339 = sinks.TimeStampFormatRFC3339
)

type Encoding struct {
	ID    string
	Codec helpers.OptionalPair
	//ExceptFields is a VRL acceptable List
	ExceptFields    helpers.OptionalPair
	TimeStampFormat helpers.OptionalPair
}

func NewEncoding(id, codec string, inits ...func(*Encoding)) Encoding {
	e := &Encoding{
		ID:    id,
		Codec: helpers.NewOptionalPair("codec", codec),
		ExceptFields: helpers.NewOptionalPair("except_fields",
			vectorhelpers.MakeInputs("_internal"),
			framework.Option{Name: helpers.OptionFormatter, Value: "%s = %v"},
		),
		TimeStampFormat: helpers.NewOptionalPair("timestamp_format", nil),
	}
	if codec == "" {
		e.Codec.Value = nil
	}
	for _, init := range inits {
		init(e)
	}
	return *e
}

func (e Encoding) Name() string {
	return "encoding"
}

func (e Encoding) Template() string {
	return `{{define "` + e.Name() + `" -}}
[sinks.{{.id}}.encoding]
{{.Codec }}
{{.TimeStampFormat }}
{{.ExceptFields }}
{{end}}`
}
