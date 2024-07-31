package azuremonitor

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers/security"
	"github.com/openshift/cluster-logging-operator/internal/migrations/observability/api/outputs/common"

	corev1 "k8s.io/api/core/v1"
)

func MapAzureMonitor(loggingOutSpec logging.OutputSpec, secret *corev1.Secret) *obs.AzureMonitor {
	obsAzMon := &obs.AzureMonitor{}

	// Authentication
	if secret != nil {
		obsAzMon.Authentication = &obs.AzureMonitorAuthentication{}
		if security.HasSharedKey(secret) {
			obsAzMon.Authentication.SharedKey = &obs.SecretReference{
				Key:        constants.SharedKey,
				SecretName: secret.Name,
			}
		}
	}

	// Tuning Specs
	if loggingOutSpec.Tuning != nil {
		obsAzMon.Tuning = common.MapBaseOutputTuning(*loggingOutSpec.Tuning)
	}

	loggingAzMon := loggingOutSpec.AzureMonitor
	if loggingAzMon == nil {
		return obsAzMon
	}

	obsAzMon.CustomerId = loggingAzMon.CustomerId
	obsAzMon.LogType = loggingAzMon.LogType
	obsAzMon.AzureResourceId = loggingAzMon.AzureResourceId
	obsAzMon.Host = loggingAzMon.Host

	return obsAzMon
}
