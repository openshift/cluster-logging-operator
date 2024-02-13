package outputs

import (
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
)

const (
	compressionNotSupportedForType = "compression is not supported for the output type"

	maxRetryDurationNotSupportedForType = "maxRetryDuration is not supported for the output type"
	minRetryDurationNotSupportedForType = "minRetryDuration is not supported for the output type"
)

var (
	unsupportedCompression = sets.NewString(loggingv1.OutputTypeSyslog, loggingv1.OutputTypeAzureMonitor)
	unsupportedRequest     = sets.NewString(loggingv1.OutputTypeSyslog, loggingv1.OutputTypeKafka)
)

func VerifyTuning(spec loggingv1.OutputSpec) (valid bool, msg string) {
	if spec.Tuning == nil {
		return true, ""
	}

	//compression
	if unsupportedCompression.Has(spec.Type) && spec.Tuning.Compression != "" && spec.Tuning.Compression != "none" {
		return false, compressionNotSupportedForType
	}

	//MaxRetryDuration
	if unsupportedRequest.Has(spec.Type) && spec.Tuning.MaxRetryDuration != nil && spec.Tuning.MaxRetryDuration.Seconds() > 0 {
		return false, maxRetryDurationNotSupportedForType
	}
	//MaxRetryDuration
	if unsupportedRequest.Has(spec.Type) && spec.Tuning.MinRetryDuration != nil && spec.Tuning.MinRetryDuration.Seconds() > 0 {
		return false, minRetryDurationNotSupportedForType
	}

	return true, ""
}
