package loki

import (
	"fmt"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/migrations/observability/api/outputs/common"
	corev1 "k8s.io/api/core/v1"
)

func MapLoki(loggingOutSpec logging.OutputSpec, secret *corev1.Secret) *obs.Loki {
	obsLoki := &obs.Loki{
		URLSpec: obs.URLSpec{
			URL: loggingOutSpec.URL,
		},
	}

	if secret != nil {
		obsLoki.Authentication = common.MapHTTPAuth(secret)
	}

	if loggingOutSpec.Tuning != nil {
		obsLoki.Tuning = &obs.LokiTuningSpec{
			BaseOutputTuningSpec: *common.MapBaseOutputTuning(*loggingOutSpec.Tuning),
			Compression:          loggingOutSpec.Tuning.Compression,
		}
	}

	loggingLoki := loggingOutSpec.Loki
	if loggingLoki == nil {
		return obsLoki
	}

	if loggingLoki.TenantKey != "" {
		tenantKey := loggingLoki.TenantKey
		if !strings.HasPrefix(tenantKey, ".") {
			tenantKey = "." + tenantKey
		}
		obsLoki.TenantKey = fmt.Sprintf(`{%s||"none"}`, tenantKey)
	}
	obsLoki.LabelKeys = loggingLoki.LabelKeys

	return obsLoki
}
