package common

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
)

func NewApiEncoding(codecType api.CodecType) (e *sinks.Encoding) {
	e = &sinks.Encoding{
		Codec:        codecType,
		ExceptFields: []string{"_internal"},
	}
	return e
}
