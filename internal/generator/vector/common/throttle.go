package common

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/transforms"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
)

func NewThrottle(inputs []string, threshHold int64, throttleKey string) types.Transform {
	return transforms.NewThrottle(func(t *transforms.Throttle) {
		t.WindowSecs = 1
		t.Threshold = uint64(threshHold)
		t.KeyField = throttleKey
	}, inputs...)
}
