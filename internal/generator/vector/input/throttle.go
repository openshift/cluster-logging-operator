package input

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/common"
)

const (
	perContainerLimitKeyField = "{{ _internal.file }}"
)

func AddThrottleToInput(input string, maxRecordsPerSec int64) types.Transform {
	return common.NewThrottle(
		[]string{input},
		maxRecordsPerSec,
		perContainerLimitKeyField,
	)
}
