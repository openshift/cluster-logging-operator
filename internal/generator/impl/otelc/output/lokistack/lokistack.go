package lokistack

import (
	"fmt"
	"strings"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/forwarder/output/lokistack"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/exporters"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/extensions"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/helpers/tls"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

// New creates OTLPHTTP exporters for LokiStack using OpenShift Logging tenancy model.
// It generates one exporter per tenant (application, infrastructure, audit) based on the input types.
//
// Returns:
//   - exporters: map of exporter ID -> configured OTLPHTTP exporter
//   - exts: map of extension ID -> configured extension (e.g. bearertokenauth)
func New(id string, o obs.OutputSpec, inputSpecs []obs.InputSpec, secrets internalobs.Secrets, op utils.Options) (exportersMap api.Exporters, exts api.Extensions) {
	if o.LokiStack == nil {
		panic("LokiStack output spec is nil")
	}

	exportersMap = make(api.Exporters)
	exts = make(api.Extensions)

	// Create bearertokenauth extension if needed
	var authExt *extensions.BearerTokenAuth
	if o.LokiStack.Authentication != nil && o.LokiStack.Authentication.Token != nil {
		authExt = newBearerTokenAuthExtension(id, o.LokiStack.Authentication.Token, op)
		if authExt != nil {
			exts.Add(authExt)
		}
	}

	// Determine tenants based on input types
	tenants := internalobs.Inputs(inputSpecs).DetermineTenants()

	// Create an exporter for each tenant
	for _, tenant := range tenants {
		exporterID := helpers.MakeOutputID(id, tenant)
		exporter := generateExporterForTenant(exporterID, o, tenant, authExt, op)
		exportersMap[exporter.ID()] = exporter
	}

	return exportersMap, exts
}

func newBearerTokenAuthExtension(outputID string, token *obs.BearerToken, op utils.Options) *extensions.BearerTokenAuth {
	switch token.From {
	case obs.BearerTokenFromServiceAccount:
		if name, found := utils.GetOption[string](op, framework.OptionServiceAccountTokenSecretName, ""); found {
			tokenPath := internalobs.SecretPath(name, constants.TokenKey, "%s")
			return extensions.NewBearerTokenAuth(outputID, tokenPath)
		}
	case obs.BearerTokenFromSecret:
		if token.Secret != nil {
			tokenPath := internalobs.SecretPath(token.Secret.Name, token.Secret.Key, "%s")
			return extensions.NewBearerTokenAuth(outputID, tokenPath)
		}
	}
	return nil
}

// generateExporterForTenant creates an OTLPHTTP exporter configured for a specific tenant.
func generateExporterForTenant(exporterID string, o obs.OutputSpec, tenant string, authExt *extensions.BearerTokenAuth, op utils.Options) *exporters.OtlpHttp {
	// Build the LokiStack OTLP endpoint URL for this tenant
	url := lokistack.OtlpURL(o.LokiStack, tenant)
	url, _ = strings.CutSuffix(url, "/v1/logs")

	if url == "" {
		panic(fmt.Sprintf("LokiStack output has no valid URL for tenant %s", tenant))
	}

	// Create OTLPHTTP exporter
	exporter := exporters.NewOtlpHttp(exporterID, url)

	// Reference the authenticator extension
	if authExt != nil {
		exporter.Auth = &exporters.AuthConfig{
			Authenticator: authExt.ID(),
		}
	}

	// Configure TLS if specified
	if o.TLS != nil {
		exporter.TLS = tls.NewTlsClientConfig(o.TLS, op)
	}

	// Apply tuning configuration
	if o.LokiStack.Tuning != nil {
		applyLokiTuning(exporter, o.LokiStack.Tuning)
	}

	return exporter
}

// applyLokiTuning applies LokiStack tuning configuration to the OTLPHTTP exporter.
func applyLokiTuning(exporter *exporters.OtlpHttp, tuning *obs.LokiTuningSpec) {
	if tuning == nil {
		return
	}

	// Set compression
	if tuning.Compression != "" {
		switch tuning.Compression {
		case "gzip":
			exporter.Compression = "gzip"
		case "snappy":
			exporter.Compression = "gzip"
		case "none":
			exporter.Compression = "none"
		default:
			exporter.Compression = "gzip"
		}
	}

	// Apply delivery mode (batch vs simple)
	if tuning.DeliveryMode == obs.DeliveryModeAtLeastOnce {
		exporter.RetryOnFailure = &types.RetrySettings{
			Enabled:         true,
			InitialInterval: "5s",
			MaxInterval:     "30s",
			MaxElapsedTime:  "5m",
		}

		exporter.SendingQueue = &types.QueueSettings{
			Enabled:      true,
			NumConsumers: 10,
			QueueSize:    1000,
		}
	}
}
