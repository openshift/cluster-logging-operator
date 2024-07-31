package splunk

import (
	"fmt"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers/security"
	"github.com/openshift/cluster-logging-operator/internal/migrations/observability/api/outputs/common"
	corev1 "k8s.io/api/core/v1"
)

func MapSplunk(loggingOutSpec logging.OutputSpec, secret *corev1.Secret) *obs.Splunk {
	obsSplunk := &obs.Splunk{
		URLSpec: obs.URLSpec{
			URL: loggingOutSpec.URL,
		},
	}

	// Set auth
	if secret != nil {
		obsSplunk.Authentication = &obs.SplunkAuthentication{}
		if security.HasSplunkHecToken(secret) {
			obsSplunk.Authentication.Token = &obs.SecretReference{
				Key:        constants.SplunkHECTokenKey,
				SecretName: secret.Name,
			}
		}
	}

	// Set tuning
	if loggingOutSpec.Tuning != nil {
		obsSplunk.Tuning = &obs.SplunkTuningSpec{
			BaseOutputTuningSpec: *common.MapBaseOutputTuning(*loggingOutSpec.Tuning),
		}
	}

	loggingSplunk := loggingOutSpec.Splunk

	if loggingSplunk == nil {
		return obsSplunk
	}

	// Set index if specified
	var splunkIndex string
	if loggingSplunk.IndexKey != "" {
		indexKey := loggingSplunk.IndexKey
		if !strings.HasPrefix(indexKey, ".") {
			indexKey = "." + indexKey
		}
		splunkIndex = fmt.Sprintf(`{%s||""}`, indexKey)
	} else if loggingSplunk.IndexName != "" {
		splunkIndex = loggingSplunk.IndexName
	}
	obsSplunk.Index = splunkIndex

	return obsSplunk
}
