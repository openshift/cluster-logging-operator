package elasticsearch

import (
	"fmt"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/migrations/observability/api/outputs/common"
	corev1 "k8s.io/api/core/v1"
)

func MapElasticsearch(loggingOutSpec logging.OutputSpec, secret *corev1.Secret) *obs.Elasticsearch {
	obsEs := &obs.Elasticsearch{
		URLSpec: obs.URLSpec{
			URL: loggingOutSpec.URL,
		},
		Version: 8,
		Index:   `{.log_type||"none"}`,
	}

	if secret != nil {
		obsEs.Authentication = common.MapHTTPAuth(secret)
	}

	if loggingOutSpec.Tuning != nil {
		obsEs.Tuning = &obs.ElasticsearchTuningSpec{
			BaseOutputTuningSpec: *common.MapBaseOutputTuning(*loggingOutSpec.Tuning),
			Compression:          loggingOutSpec.Tuning.Compression,
		}
	}

	loggingES := loggingOutSpec.Elasticsearch
	if loggingES == nil {
		return obsEs
	}

	obsEs.Version = loggingES.Version

	if loggingES.StructuredTypeKey != "" && loggingES.StructuredTypeName != "" {
		// Ensure StructuredTypeKey begins with `.`
		structuredTypeKey := loggingES.StructuredTypeKey
		if !strings.HasPrefix(structuredTypeKey, ".") {
			structuredTypeKey = "." + structuredTypeKey
		}
		obsEs.Index = fmt.Sprintf("{%s||%q}", structuredTypeKey, loggingES.StructuredTypeName)
	}

	return obsEs
}
