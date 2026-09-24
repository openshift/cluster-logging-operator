package lokistack

import (
	"fmt"
	"slices"
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/adapters"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/transforms"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/metrics"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/loki"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/otlp"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
)

// New creates generate elements that represent configuration to forward logs to Loki using OpenShift Logging tenancy model
func New(id string, o *adapters.Output, inputs []string, secrets observability.Secrets, op utils.Options) (sinks api.Sinks, tfs api.Transforms) {
	tfs = api.Transforms{}
	sinks = api.Sinks{}
	clfSpec, _ := utils.GetOption(op, vectorhelpers.CLFSpec, observability.ClusterLogForwarderSpec{})
	if len(clfSpec.Inputs) == 0 || len(clfSpec.Pipelines) == 0 || len(clfSpec.Outputs) == 0 {
		panic("ClusterLogForwarderSpec not found while generating LokiStack config")
	}

	// Add trace context extraction remap if data model is OpenTelemetry as top level remap
	if o.LokiStack != nil && o.LokiStack.DataModel == obs.LokiStackDataModelOpenTelemetry {
		transformTraceContextID := vectorhelpers.MakeID(id, "trace", "context")
		tfs[transformTraceContextID] = otlp.TransformTraceContext(inputs)
		inputs = []string{transformTraceContextID}
	}

	inputSpecs := clfSpec.InputSpecsTo(o.OutputSpec)
	tenants := determineTenants(inputSpecs)
	routeID := vectorhelpers.MakeID(id, "route")
	tfs[routeID] = transforms.NewRoute(func(r *transforms.Route) {
		r.Routes = buildRoutes(tenants)
	}, inputs...)

	tfs[vectorhelpers.MakeID(routeID, transforms.Unmatched)] = metrics.NewUnmatched(routeID, op, map[string]string{
		"output_type": strings.ToLower(obs.OutputTypeLokiStack.String()),
	})

	for _, inputType := range tenants.List() {
		tenantId, tenantSink, tenantTfs := generateSinkForTenant(id, routeID, inputType, o.OutputSpec, inputSpecs, secrets, op)
		sinks[tenantId] = tenantSink
		for transformId, t := range tenantTfs {
			tfs[transformId] = t
		}
	}

	return sinks, tfs
}

func determineTenants(inputSpecs []obs.InputSpec) *sets.String {
	tenants := sets.NewString()

	for _, inputSpec := range inputSpecs {
		switch inputSpec.Type {
		case obs.InputTypeApplication:
			tenants.Insert(string(obs.InputTypeApplication))
			if observability.IncludesInfraNamespace(inputSpec.Application) {
				tenants.Insert(string(obs.InputTypeInfrastructure))
			}
		case obs.InputTypeAudit:
			tenants.Insert(string(obs.InputTypeAudit))
		case obs.InputTypeInfrastructure:
			tenants.Insert(string(obs.InputTypeInfrastructure))
		case obs.InputTypeReceiver:
			tenants.Insert(getTenantForReceiver(inputSpec.Receiver.Type))
		}
	}

	return tenants
}

func getTenantForReceiver(receiverType obs.ReceiverType) string {
	if receiverType == obs.ReceiverTypeHTTP {
		return string(obs.InputTypeAudit)
	}
	return string(obs.InputTypeInfrastructure)
}

func buildRoutes(tenants *sets.String) map[string]string {
	routes := make(map[string]string, tenants.Len())
	for _, tenant := range tenants.List() {
		routes[tenant] = fmt.Sprintf(".log_type == %q", tenant)
	}
	return routes
}

func generateSinkForTenant(id, routeID, inputType string, o obs.OutputSpec, inputSpecs []obs.InputSpec,
	secrets observability.Secrets, op utils.Options) (string, types.Sink, api.Transforms) {

	outputID := vectorhelpers.MakeID(id, inputType)
	migratedOutput := GenerateOutput(o, inputType)
	log.V(4).Info("migrated lokistack output", "spec", migratedOutput)

	factoryInput := vectorhelpers.MakeRouteInputID(routeID, inputType)

	if migratedOutput.Type == obs.OutputTypeOTLP {
		op[otlp.OtlpLogSourcesOption] = getInputSources(inputSpecs, obs.InputType(inputType))
		op[otlp.MigratedFromLokistackOption] = true
		adapter := adapters.NewOutput(migratedOutput)
		adapter.InputIDs = append(adapter.InputIDs, factoryInput)
		return otlp.New(outputID, adapter, []string{factoryInput}, secrets, op)
	}

	tenantID, sink, tfs := loki.New(outputID, adapters.NewOutput(migratedOutput), []string{factoryInput}, secrets, op)
	applyMissingAuditVerbFallback(outputID, inputType, migratedOutput.Loki, tfs)
	return tenantID, sink, tfs
}

// applyMissingAuditVerbFallback sets .verb to "unknown" when the field is
// missing or null so ViaQ audit streams stay labeled without changing the
// shared Loki remap used by application and infrastructure sinks.
func applyMissingAuditVerbFallback(outputID, inputType string, lokiSpec *obs.Loki, tfs api.Transforms) {
	if obs.InputType(inputType) != obs.InputTypeAudit {
		return
	}
	if lokiSpec == nil || !slices.Contains(lokiSpec.LabelKeys, auditVerbLabelKey) {
		return
	}
	remapLabelID := vectorhelpers.MakeID(outputID, "remap_label")
	remap, ok := tfs[remapLabelID].(*transforms.Remap)
	if !ok {
		panic(fmt.Sprintf("expected remap transform %q for audit LokiStack sink", remapLabelID))
	}
	vrl := fmt.Sprintf(
		"\nif !(exists(.%s) && !is_null(.%s)) {\n  .%s = %q\n}",
		auditVerbLabelKey, auditVerbLabelKey, auditVerbLabelKey, unknownAuditVerb,
	)
	tfs[remapLabelID] = transforms.NewRemap(string(remap.Source)+vrl, remap.Inputs...)
}

func getInputSources(inputSpecs []obs.InputSpec, inputType obs.InputType) []string {
	inputSources := observability.Inputs(inputSpecs).InputSources(inputType)
	addReceiverSources(&inputSources, inputSpecs, inputType)
	return inputSources
}

func addReceiverSources(inputSources *[]string, inputSpecs []obs.InputSpec, inputType obs.InputType) {
	for _, is := range inputSpecs {
		if is.Type != obs.InputTypeReceiver {
			continue
		}

		if inputType == obs.InputTypeAudit && is.Receiver.Type == obs.ReceiverTypeHTTP {
			*inputSources = append(*inputSources, "receiver.http")
		}

		if inputType == obs.InputTypeInfrastructure && is.Receiver.Type == obs.ReceiverTypeSyslog {
			*inputSources = append(*inputSources, "receiver.syslog")
		}
	}
}
