package input

import (
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
)

const (
	perContainerLimitKeyField = `"{{ _internal.file }}"`
)

func AddThrottleToInput(id, input string, maxRecordsPerSec int64) []Element {
	throttleKey := perContainerLimitKeyField
	return normalize.NewThrottle(
		id,
		[]string{input},
		maxRecordsPerSec,
		throttleKey,
	)
}
