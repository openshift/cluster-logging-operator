package lokistack

import (
	"fmt"
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/loki"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/otlp"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
)

// New creates generate elements that represent configuration to forward logs to Loki using OpenShift Logging tenancy model
func New(id string, o obs.OutputSpec, inputs []string, secrets observability.Secrets, strategy common.ConfigStrategy, op utils.Options) []framework.Element {
	clfSpec, _ := utils.GetOption(op, vectorhelpers.CLFSpec, observability.ClusterLogForwarderSpec{})
	if len(clfSpec.Inputs) == 0 || len(clfSpec.Pipelines) == 0 || len(clfSpec.Outputs) == 0 {
		panic("ClusterLogForwarderSpec not found while generating LokiStack config")
	}

	inputSpecs := clfSpec.InputSpecsTo(o)
	tenants := determineTenants(inputSpecs)

	routeID := vectorhelpers.MakeID(id, "route")
	confs := []framework.Element{
		elements.Route{
			ComponentID: routeID,
			Inputs:      vectorhelpers.MakeInputs(inputs...),
			Routes:      buildRoutes(tenants),
		},
	}

	confs = append(confs, elements.NewUnmatched(routeID, op, map[string]string{
		"output_type": strings.ToLower(obs.OutputTypeLokiStack.String()),
	}))

	for _, inputType := range tenants.List() {
		confs = append(confs, generateSinkForTenant(id, routeID, inputType, o, inputSpecs, secrets, strategy, op)...)
	}

	return confs
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
		routes[tenant] = fmt.Sprintf("'.log_type == \"%s\"'", tenant)
	}
	return routes
}

func generateSinkForTenant(id, routeID, inputType string, o obs.OutputSpec, inputSpecs []obs.InputSpec,
	secrets observability.Secrets, strategy common.ConfigStrategy, op utils.Options) []framework.Element {

	outputID := vectorhelpers.MakeID(id, inputType)
	migratedOutput := GenerateOutput(o, inputType)
	log.V(4).Info("migrated lokistack output", "spec", migratedOutput)

	factoryInput := vectorhelpers.MakeRouteInputID(routeID, inputType)

	if migratedOutput.Type == obs.OutputTypeOTLP {
		op[otlp.OtlpLogSourcesOption] = getInputSources(inputSpecs, obs.InputType(inputType))
		return otlp.New(outputID, migratedOutput, []string{factoryInput}, secrets, strategy, op)
	}

	return loki.New(outputID, migratedOutput, []string{factoryInput}, secrets, strategy, op)
}

func getInputSources(inputSpecs []obs.InputSpec, inputType obs.InputType) []string {
	inputSources := observability.Inputs(inputSpecs).InputSources(inputType)

	if len(inputSources) == 0 && inputType == obs.InputTypeInfrastructure {
		inputSources = append(inputSources, observability.ReservedInfrastructureSources.List()...)
	}

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