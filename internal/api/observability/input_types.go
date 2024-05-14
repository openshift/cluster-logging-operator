package observability

import obs "github.com/openshift/cluster-logging-operator/api/observability/v1"

func MaxRecordsPerSecond(input obs.InputSpec) (int64, bool) {
	if input.Application != nil &&
		input.Application.Tuning != nil &&
		input.Application.Tuning.RateLimitPerContainer != nil {
		return input.Application.Tuning.RateLimitPerContainer.MaxRecordsPerSecond, true
	}
	return 0, false
}
