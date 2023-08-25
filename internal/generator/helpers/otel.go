package helpers

import (
	"github.com/openshift/cluster-logging-operator/internal/generator"
	elements2 "github.com/openshift/cluster-logging-operator/internal/generator/fluentd/elements"
)

const (
	EnableOtel = "otel-output"
)

func IsOtelOutput(op generator.Options) bool {
	_, ok := op[EnableOtel]
	return ok
}

var OtelOutput = generator.ConfLiteral{
	Desc:         "Sending records to stdout for debug purposes",
	TemplateName: "toStdout",
	Pattern:      "**",
	TemplateStr:  elements2.ToStdOut,
}