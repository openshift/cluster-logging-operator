package gcl

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers/security"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/gcl"
	"github.com/openshift/cluster-logging-operator/internal/migrations/observability/api/outputs/common"
	corev1 "k8s.io/api/core/v1"
)

func MapGoogleCloudLogging(loggingOutSpec logging.OutputSpec, secret *corev1.Secret) *obs.GoogleCloudLogging {
	obsGcp := &obs.GoogleCloudLogging{}

	if secret != nil {
		obsGcp.Authentication = &obs.GoogleCloudLoggingAuthentication{}
		if security.HasGoogleApplicationCredentialsKey(secret) {
			obsGcp.Authentication.Credentials = &obs.SecretReference{
				Key:        gcl.GoogleApplicationCredentialsKey,
				SecretName: secret.Name,
			}
		}
	}
	if loggingOutSpec.Tuning != nil {
		obsGcp.Tuning = &obs.GoogleCloudLoggingTuningSpec{
			BaseOutputTuningSpec: *common.MapBaseOutputTuning(*loggingOutSpec.Tuning),
		}
	}

	loggingGcp := loggingOutSpec.GoogleCloudLogging
	if loggingGcp == nil {
		return obsGcp
	}
	if loggingGcp.BillingAccountID != "" {
		obsGcp.ID = obs.GoogleCloudLoggingID{
			Type:  obs.GoogleCloudLoggingIDTypeBillingAccount,
			Value: loggingGcp.BillingAccountID,
		}
	} else if loggingGcp.OrganizationID != "" {
		obsGcp.ID = obs.GoogleCloudLoggingID{
			Type:  obs.GoogleCloudLoggingIDTypeOrganization,
			Value: loggingGcp.OrganizationID,
		}
	} else if loggingGcp.FolderID != "" {
		obsGcp.ID = obs.GoogleCloudLoggingID{
			Type:  obs.GoogleCloudLoggingIDTypeFolder,
			Value: loggingGcp.FolderID,
		}
	} else if loggingGcp.ProjectID != "" {
		obsGcp.ID = obs.GoogleCloudLoggingID{
			Type:  obs.GoogleCloudLoggingIDTypeProject,
			Value: loggingGcp.ProjectID,
		}
	}

	obsGcp.LogID = loggingGcp.LogID

	return obsGcp
}
