package common

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types/codec"
)

func NewApiEncoding(codecType codec.CodecType) (e *sinks.Encoding) {
	e = &sinks.Encoding{
		Codec:        codecType,
		ExceptFields: []string{"_internal"},
	}
	return e
}
