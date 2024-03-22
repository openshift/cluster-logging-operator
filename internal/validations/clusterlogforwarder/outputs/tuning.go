package outputs

import (
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
)

const (
	compressionNotSupportedForType = "compression is not supported for the output type"

	maxRetryDurationNotSupportedForType = "maxRetryDuration is not supported for the output type"
	minRetryDurationNotSupportedForType = "minRetryDuration is not supported for the output type"
	maxWriteNotSupportedForType         = "maxWrite is not supported for the output type"
)

var (
	compressionOutputMap = map[string]sets.String{
		"gzip": *sets.NewString(
			loggingv1.OutputTypeCloudwatch,
			loggingv1.OutputTypeElasticsearch,
			loggingv1.OutputTypeHttp,
			loggingv1.OutputTypeLoki,
			loggingv1.OutputTypeSplunk),
		"snappy": *sets.NewString(
			loggingv1.OutputTypeCloudwatch,
			loggingv1.OutputTypeHttp,
			loggingv1.OutputTypeKafka,
			loggingv1.OutputTypeLoki),
		"zlib": *sets.NewString(
			loggingv1.OutputTypeCloudwatch,
			loggingv1.OutputTypeElasticsearch,
			loggingv1.OutputTypeHttp),
		"zstd": *sets.NewString(
			loggingv1.OutputTypeCloudwatch,
			loggingv1.OutputTypeHttp,
			loggingv1.OutputTypeKafka,
		),
		"lz4": *sets.NewString(loggingv1.OutputTypeKafka),
	}
	unsupportedRequest = sets.NewString(loggingv1.OutputTypeSyslog, loggingv1.OutputTypeKafka)
	unsupportedBatch   = sets.NewString(loggingv1.OutputTypeSyslog)
)

func VerifyTuning(spec loggingv1.OutputSpec) (bool, string) {
	if spec.Tuning == nil {
		return true, ""
	}

	// compression
	if msg := verifyCompression(spec.Type, spec.Tuning.Compression); msg != "" {
		return false, msg
	}

	// batch
	if unsupportedBatch.Has(spec.Type) && spec.Tuning.MaxWrite != nil && !spec.Tuning.MaxWrite.IsZero() {
		return false, maxWriteNotSupportedForType
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

func verifyCompression(outputType string, compression string) string {
	if compression == "" || compression == "none" {
		return ""
	}

	if outputs, ok := compressionOutputMap[compression]; ok {
		if !outputs.Has(outputType) {
			return compressionNotSupportedForType
		}
		// Compression not in map
	} else {
		return compressionNotSupportedForType
	}

	return ""
}
