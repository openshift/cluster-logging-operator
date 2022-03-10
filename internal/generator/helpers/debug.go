package helpers

import (
	"github.com/openshift/cluster-logging-operator/internal/generator"
	elements2 "github.com/openshift/cluster-logging-operator/internal/generator/fluentd/elements"
)

const (
	EnableDebugOutput = "debug-output"
)

func IsDebugOutput(op generator.Options) bool {
	_, ok := op[EnableDebugOutput]
	return ok
}

var DebugOutput = generator.ConfLiteral{
	Desc:         "Sending records to stdout for debug purposes",
	TemplateName: "toStdout",
	Pattern:      "**",
	TemplateStr:  elements2.ToStdOut,
}
