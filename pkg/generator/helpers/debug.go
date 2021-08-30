package helpers

import (
	. "github.com/openshift/cluster-logging-operator/pkg/generator"
	"github.com/openshift/cluster-logging-operator/pkg/generator/fluentd/elements"
)

func IsDebugOutput(op Options) bool {
	_, ok := op["debug_output"]
	return ok
}

var DebugOutput = ConfLiteral{
	Desc:         "Sending records to stdout for debug purposes",
	TemplateName: "toStdout",
	Pattern:      "**",
	TemplateStr:  elements.ToStdOut,
}
