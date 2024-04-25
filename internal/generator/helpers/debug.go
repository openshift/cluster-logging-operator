package helpers

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
)

const (
	EnableDebugOutput = "debug-output"
)

func IsDebugOutput(op framework.Options) bool {
	_, ok := op[EnableDebugOutput]
	return ok
}
