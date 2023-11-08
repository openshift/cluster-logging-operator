package helpers

import (
	elements2 "github.com/openshift/cluster-logging-operator/internal/generator/fluentd/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
)

const (
	EnableDebugOutput = "debug-output"
)

func IsDebugOutput(op framework.Options) bool {
	_, ok := op[EnableDebugOutput]
	return ok
}

var DebugOutput = framework.ConfLiteral{
	Desc:         "Sending records to stdout for debug purposes",
	TemplateName: "toStdout",
	Pattern:      "**",
	TemplateStr:  elements2.ToStdOut,
}
