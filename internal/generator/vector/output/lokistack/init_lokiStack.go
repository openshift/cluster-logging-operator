package lokistack

import (
	"fmt"
	"slices"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	lokioutput "github.com/openshift/cluster-logging-operator/internal/generator/vector/output/loki"
)

const (
	lokiOtlpEndpoint  = "/otlp/v1/logs"
	auditVerbLabelKey = "verb"
	unknownAuditVerb  = "unknown"
)

// GenerateOutput returns either a Loki or OTLP output spec when migrating Lokistacks
func GenerateOutput(outSpec obs.OutputSpec, tenant string) obs.OutputSpec {
	obsOut := obs.OutputSpec{
		Name:  fmt.Sprintf("%s-%s", outSpec.Name, tenant),
		TLS:   outSpec.TLS,
		Limit: outSpec.Limit,
	}
	if outSpec.LokiStack.DataModel == obs.LokiStackDataModelOpenTelemetry {
		obsOut.Type = obs.OutputTypeOTLP
		obsOut.OTLP = GenerateOtlpSpec(outSpec.LokiStack, tenant)
	} else {
		obsOut.Type = obs.OutputTypeLoki
		obsOut.Loki = GenerateLokiSpec(outSpec.LokiStack, tenant)
	}

	return obsOut
}

// GenerateLokiSpec generates and returns a Loki spec for the defined lokistack output
func GenerateLokiSpec(ls *obs.LokiStack, tenant string) *obs.Loki {
	url := lokiStackURL(ls, tenant, false)
	if url == "" {
		panic("LokiStack output has no valid URL")
	}
	return &obs.Loki{
		URLSpec: obs.URLSpec{
			URL: url,
		},
		Authentication: &obs.HTTPAuthentication{
			Token: ls.Authentication.Token,
		},
		Tuning:    ls.Tuning,
		LabelKeys: lokiStackLabelKeysForTenant(ls.LabelKeys, tenant, defaultLabelKeysForTenant(tenant)),
	}
}

// defaultLabelKeysForTenant returns the ViaQ Loki stream label keys used when a
// tenant does not customize labelKeys. Audit includes verb so common filters
// such as {verb="create"} do not require parsing every log line as JSON.
func defaultLabelKeysForTenant(tenant string) []string {
	if obs.InputType(tenant) != obs.InputTypeAudit {
		return lokioutput.DefaultLabelKeys
	}
	return append(slices.Clone(lokioutput.DefaultLabelKeys), auditVerbLabelKey)
}

// GenerateOtlpSpec generates and returns an OTLP spec for the defined lokistack output
// Note: OTLP does not support compression type `snappy` where loki does
// This also does not support LabelKeys
func GenerateOtlpSpec(ls *obs.LokiStack, tenant string) *obs.OTLP {
	return &obs.OTLP{
		URL: lokiStackURL(ls, tenant, true),
		Authentication: &obs.HTTPAuthentication{
			Token: ls.Authentication.Token,
		},
		Tuning: (*obs.OTLPTuningSpec)(ls.Tuning),
	}
}

func lokiStackURL(lokiStackSpec *obs.LokiStack, tenant string, otlp bool) string {
	service := lokiStackGatewayService(lokiStackSpec.Target.Name)
	if !internalobs.ReservedInputTypes.Has(tenant) {
		return ""
	}
	baseURL := fmt.Sprintf("https://%s.%s.svc:8080/api/logs/v1/%s", service, lokiStackSpec.Target.Namespace, tenant)

	// return OTLP endpoint appended to the base lokistack URL if output is OTLP
	if otlp {
		return baseURL + lokiOtlpEndpoint
	}
	return baseURL
}

func lokiStackGatewayService(lokiStackServiceName string) string {
	return fmt.Sprintf("%s-gateway-http", lokiStackServiceName)
}

// lokiStackLabelKeysForTenant returns the per-tenant labelKeys for a Loki output based on the LokiStack configuration.
// A return value of "nil" indicates that the defaults of the Loki output should be used.
// Audit is an exception: its defaults include verb in addition to the Loki output
// defaults, so unconfigured audit tenants return those keys explicitly.
func lokiStackLabelKeysForTenant(labelKeys *obs.LokiStackLabelKeys, tenant string, defaultKeys []string) []string {
	if labelKeys == nil {
		return unconfiguredTenantLabelKeys(tenant, defaultKeys)
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

	if len(keys) == 0 {
		if ignoreGlobal {
			return keys
		}
		return unconfiguredTenantLabelKeys(tenant, defaultKeys)
	}

	if len(keys) > 1 {
		slices.Sort(keys)
		keys = slices.Compact(keys)
	}

	return keys
}

func unconfiguredTenantLabelKeys(tenant string, defaultKeys []string) []string {
	if obs.InputType(tenant) != obs.InputTypeAudit {
		return nil
	}
	keys := slices.Clone(defaultKeys)
	slices.Sort(keys)
	return slices.Compact(keys)
}
