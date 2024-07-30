package initialize

import (
	"fmt"
	"slices"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	lokioutput "github.com/openshift/cluster-logging-operator/internal/generator/vector/output/loki"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

// MigrateLokiStack migrates a lokistack output into appropriate loki outputs based on defined inputs
func MigrateLokiStack(spec obs.ClusterLogForwarder, options utils.Options) obs.ClusterLogForwarder {
	var outputs []obs.OutputSpec
	var pipelines []obs.PipelineSpec
	outputs, pipelines = ProcessForwarderPipelines(spec.Spec, obs.OutputTypeLokiStack, "", false)

	spec.Spec.Outputs = outputs
	spec.Spec.Pipelines = pipelines

	return spec
}

func GenerateLokiOutput(outSpec obs.OutputSpec, input, tenant string) obs.OutputSpec {
	return obs.OutputSpec{
		Name: fmt.Sprintf("%s-%s", outSpec.Name, input),
		Type: obs.OutputTypeLoki,
		Loki: &obs.Loki{
			URLSpec: obs.URLSpec{
				URL: lokiStackURL(outSpec.LokiStack, tenant),
			},
			Authentication: &obs.HTTPAuthentication{
				Token: outSpec.LokiStack.Authentication.Token,
			},
			Tuning:    outSpec.LokiStack.Tuning,
			LabelKeys: lokiStackLabelKeysForTenant(outSpec.LokiStack.LabelKeys, tenant, lokioutput.DefaultLabelKeys),
		},
		TLS:   outSpec.TLS,
		Limit: outSpec.Limit,
	}
}

func lokiStackURL(lokiStackSpec *obs.LokiStack, tenant string) string {
	service := lokiStackGatewayService(lokiStackSpec.Target.Name)
	if !obs.ReservedInputTypes.Has(tenant) {
		return ""
	}
	return fmt.Sprintf("https://%s.%s.svc:8080/api/logs/v1/%s", service, lokiStackSpec.Target.Namespace, tenant)
}

func lokiStackGatewayService(lokiStackServiceName string) string {
	return fmt.Sprintf("%s-gateway-http", lokiStackServiceName)
}

// lokiStackLabelKeysForTenant returns the per-tenant labelKeys for a Loki output based on the LokiStack configuration.
// A return value of "nil" indicates that the defaults of the Loki output should be used.
func lokiStackLabelKeysForTenant(labelKeys *obs.LokiStackLabelKeys, tenant string, defaultKeys []string) []string {
	if labelKeys == nil {
		return nil
	}

	var tenantConfig *obs.LokiStackTenantLabelKeys
	switch obs.InputType(tenant) {
	case obs.InputTypeApplication:
		tenantConfig = labelKeys.Application
	case obs.InputTypeInfrastructure:
		tenantConfig = labelKeys.Infrastructure
	case obs.InputTypeAudit:
		tenantConfig = labelKeys.Audit
	}

	var keys []string
	ignoreGlobal := false

	if tenantConfig != nil {
		ignoreGlobal = tenantConfig.IgnoreGlobal
		if len(tenantConfig.LabelKeys) > 0 {
			keys = append(keys, tenantConfig.LabelKeys...)
		}
	}

	if !ignoreGlobal {
		if len(labelKeys.Global) > 0 {
			keys = append(keys, labelKeys.Global...)
		} else if len(keys) > 0 {
			// If the per-tenant configuration is custom, but there is no customization of the global keys
			// then we need to explicitly add the default keys.
			keys = append(keys, defaultKeys...)
		}
	}

	if len(keys) > 1 {
		slices.Sort(keys)
		keys = slices.Compact(keys)
	}

	return keys
}
