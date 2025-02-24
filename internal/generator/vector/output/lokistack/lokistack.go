package lokistack

import (
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/loki"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/otlp"
	"strings"
)

// New creates generate elements that represent configuration to forward logs to Loki using OpenShift Logging tenancy model
func New(id string, o obs.OutputSpec, inputs []string, secrets observability.Secrets, strategy common.ConfigStrategy, op framework.Options) []framework.Element {
	routeID := vectorhelpers.MakeID(id, "route")
	routes := map[string]string{}
	for _, inputType := range observability.ReservedInputTypes.List() {
		routes[strings.ToLower(string(inputType))] = fmt.Sprintf("'.log_type == \"%s\"'", inputType)
	}
	confs := []framework.Element{
		elements.Route{
			ComponentID: routeID,
			Inputs:      vectorhelpers.MakeInputs(inputs...),
			Routes:      routes,
		},
	}
	for _, inputType := range observability.ReservedInputTypes.List() {
		outputID := vectorhelpers.MakeID(id, string(inputType))
		migratedOutput := GenerateOutput(o, string(inputType))
		log.V(4).Info("migrated lokistack output", "spec", migratedOutput)
		factory := loki.New
		if migratedOutput.Type == obs.OutputTypeOTLP {
			factory = otlp.New
			switch obs.InputType(inputType) {
			case obs.InputTypeApplication:
				op[otlp.OtlpLogSourcesOption] = observability.ReservedApplicationSources.List()
			case obs.InputTypeInfrastructure:
				op[otlp.OtlpLogSourcesOption] = observability.ReservedInfrastructureSources.List()
			case obs.InputTypeAudit:
				op[otlp.OtlpLogSourcesOption] = observability.ReservedAuditSources.List()
			}
		}
		confs = append(confs, factory(outputID, migratedOutput, []string{vectorhelpers.MakeRouteInputID(routeID, string(inputType))}, secrets, strategy, op)...)
	}
	return confs
}
