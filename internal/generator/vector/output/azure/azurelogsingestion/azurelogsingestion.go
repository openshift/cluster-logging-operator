package azurelogsingestion

import (
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	collectorcommon "github.com/openshift/cluster-logging-operator/internal/collector/common"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/adapters"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/common/tls"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/utils"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

const (
	azureCredentialKindWorkloadIdentity = "workload_identity"
)

func auth(s *sinks.AzureLogsIngestion, azli *obs.AzureLogsIngestion) {
	if azli == nil || azli.Authentication == nil {
		return
	}

	auth := &sinks.AzureLogsIngestionAuth{}
	azliAuth := azli.Authentication

	// Vector uses different field names for workload identity vs client secret auth
	switch azliAuth.Type {
	case obs.AzureLogsIngestionAuthTypeWorkloadIdentity:
		auth.AzureCredentialKind = azureCredentialKindWorkloadIdentity
		if azliAuth.WorkloadIdentity != nil {
			auth.TenantId = azliAuth.WorkloadIdentity.TenantId
			auth.ClientId = azliAuth.WorkloadIdentity.ClientId
			if azliAuth.WorkloadIdentity.Token != nil {
				switch azliAuth.WorkloadIdentity.Token.From {
				// Return path to the token file in both cases NOT the token itself
				case obs.BearerTokenFromSecret:
					if azliAuth.WorkloadIdentity.Token.Secret != nil {
						auth.TokenFilePath = collectorcommon.SecretPath(azliAuth.WorkloadIdentity.Token.Secret.Name, azliAuth.WorkloadIdentity.Token.Secret.Key)
					}
				default:
					auth.TokenFilePath = collectorcommon.ServiceAccountBasePath(constants.TokenKey)
				}
			}
		}
	default:
		if azliAuth.ClientSecret != nil {
			auth.AzureTenantId = azliAuth.ClientSecret.TenantId
			auth.AzureClientId = azliAuth.ClientSecret.ClientId
			if azliAuth.ClientSecret.Secret != nil {
				auth.AzureClientSecret = vectorhelpers.SecretFrom(azliAuth.ClientSecret.Secret)
			}
		}
	}

	s.Auth = auth
}

func New(id string, o *adapters.Output, inputs []string, secrets observability.Secrets, op utils.Options) (_ string, sink types.Sink, tfs api.Transforms) {
	azli := o.AzureLogsIngestion
	sink = sinks.NewAzureLogsIngestion(func(s *sinks.AzureLogsIngestion) {
		s.Endpoint = azli.URL
		s.DcrImmutableId = azli.DcrImmutableId
		s.StreamName = azli.StreamName
		s.TokenScope = azli.TokenScope
		s.TimestampField = azli.TimestampField
		auth(s, azli)
		s.Encoding = common.NewApiEncoding("")
		s.Batch = common.NewApiBatch(o)
		s.Buffer = common.NewApiBuffer(o)
		s.Request = common.NewApiRequest(o)
		s.TLS = tls.NewTls(o, secrets, op)
	}, inputs...)

	return id, sink, tfs
}
