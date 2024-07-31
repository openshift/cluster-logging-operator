package http

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/migrations/observability/api/outputs/common"
	corev1 "k8s.io/api/core/v1"
)

func MapHTTP(loggingOutSpec logging.OutputSpec, secret *corev1.Secret) *obs.HTTP {
	obsHTTP := &obs.HTTP{
		URLSpec: obs.URLSpec{
			URL: loggingOutSpec.URL,
		},
	}

	if secret != nil {
		obsHTTP.Authentication = common.MapHTTPAuth(secret)
	}

	if loggingOutSpec.Tuning != nil {
		obsHTTP.Tuning = &obs.HTTPTuningSpec{
			BaseOutputTuningSpec: *common.MapBaseOutputTuning(*loggingOutSpec.Tuning),
			Compression:          loggingOutSpec.Tuning.Compression,
		}
	}

	loggingHTTP := loggingOutSpec.Http
	if loggingHTTP == nil {
		return obsHTTP
	}

	obsHTTP.Headers = loggingHTTP.Headers
	obsHTTP.Timeout = loggingHTTP.Timeout
	obsHTTP.Method = loggingHTTP.Method

	return obsHTTP
}
