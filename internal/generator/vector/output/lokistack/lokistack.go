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
	routeID := vectorhelpers.MakeID(id, "route")
	routes := map[string]string{}
	var clfSpec, _ = utils.GetOption(op, vectorhelpers.CLFSpec, observability.ClusterLogForwarderSpec{})
	if len(clfSpec.Inputs) == 0 || len(clfSpec.Pipelines) == 0 || len(clfSpec.Outputs) == 0 {
		panic("ClusterLogForwarderSpec not found while generating LokiStack config")
	}

	inputSpecs := clfSpec.InputSpecsTo(o)
	inputTypes := sets.NewString()
	for _, inputSpec := range inputSpecs {
		inputType := strings.ToLower(inputSpec.Type.String())
		inputTypes.Insert(inputType)
		routes[inputType] = fmt.Sprintf("'.log_type == \"%s\"'", inputType)
		if inputSpec.Type == obs.InputTypeApplication && observability.IncludesInfraNamespace(inputSpec.Application) {
			inputType = strings.ToLower(obs.InputTypeInfrastructure.String())
			routes[inputType] = fmt.Sprintf("'.log_type == \"%s\"'", inputType)
			inputTypes.Insert(inputType)
		}
	}
	confs := []framework.Element{
		elements.Route{
			ComponentID: routeID,
			Inputs:      vectorhelpers.MakeInputs(inputs...),
			Routes:      routes,
		},
	}
	confs = append(confs, elements.NewUnmatched(routeID, op, map[string]string{"output_type": strings.ToLower(obs.OutputTypeLokiStack.String())}))
	for _, inputType := range inputTypes.List() {
		outputID := vectorhelpers.MakeID(id, inputType)
		migratedOutput := GenerateOutput(o, inputType)
		log.V(4).Info("migrated lokistack output", "spec", migratedOutput)
		factory := loki.New
		if migratedOutput.Type == obs.OutputTypeOTLP {
			factory = otlp.New
		}
		inputSources := observability.Inputs(inputSpecs).InputSources(obs.InputType(inputType))
		if len(inputSources) == 0 && obs.InputType(inputType) == obs.InputTypeInfrastructure {
			inputSources = append(inputSources, observability.ReservedInfrastructureSources.List()...)
		}
		op[otlp.OtlpLogSourcesOption] = inputSources
		confs = append(confs, factory(outputID, migratedOutput, []string{vectorhelpers.MakeRouteInputID(routeID, inputType)}, secrets, strategy, op)...)
	}
	return confs
}
