package input

import (
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

const (
	perContainerLimitKeyField = `"{{ file }}"`
)

func AddThrottleToInput(id, input string, spec logging.InputSpec) []Element {
	var (
		threshold    int64
		throttle_key string
	)

	if spec.Application.ContainerLimit != nil {
		threshold = spec.Application.ContainerLimit.MaxRecordsPerSecond
		throttle_key = perContainerLimitKeyField
	} else {
		threshold = spec.Application.GroupLimit.MaxRecordsPerSecond
	}

	return normalize.NewThrottle(
		id,
		[]string{input},
		threshold,
		throttle_key,
	)
}
