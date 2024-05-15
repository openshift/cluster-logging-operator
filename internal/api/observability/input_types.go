package observability

import obs "github.com/openshift/cluster-logging-operator/api/observability/v1"

func MaxRecordsPerSecond(input obs.InputSpec) (int64, bool) {
	if input.Application != nil &&
		input.Application.Tuning != nil &&
		input.Application.Tuning.RateLimitPerContainer != nil {
		return Threshold(input.Application.Tuning.RateLimitPerContainer)
	}
	return 0, false
}

func Threshold(ls *obs.LimitSpec) (int64, bool) {
	if ls == nil {
		return 0, false
	}
	return ls.MaxRecordsPerSecond, true
}
