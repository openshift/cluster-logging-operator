package common

import (
	"fmt"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
)

const (
	UserDefinedSinkThrottle = `sink_throttle_%s`
)

func AddThrottleForSink(spec *logging.OutputSpec, inputs []string) []framework.Element {
	id := fmt.Sprintf(UserDefinedSinkThrottle, spec.Name)
	return normalize.NewThrottle(id, inputs, spec.Limit.MaxRecordsPerSecond, "")
}
