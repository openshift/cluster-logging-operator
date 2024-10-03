package initialize

import (
	"fmt"
	"sort"
	"strings"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
)

// ProcessForwarderPipelines migrates the specified output type and name to appropriate outputs and pipelines
func ProcessForwarderPipelines(spec obs.ClusterLogForwarderSpec) ([]obs.OutputSpec, []obs.PipelineSpec) {
	inPipelines := spec.Pipelines
	pipelines := []obs.PipelineSpec{}
	outputMap := utils.OutputMap(&spec)
	finalOutputs := []obs.OutputSpec{}

	// Remove migrated outputs
	for _, o := range spec.Outputs {
		if o.Type == obs.OutputTypeLokiStack {
			continue
		}
		finalOutputs = append(finalOutputs, o)
	}

	for _, p := range inPipelines {
		needMigrations := findLokistackOutputs(outputMap, p.OutputRefs)
		if needMigrations.Len() == 0 {
			pipelines = append(pipelines, p)
			continue
		}

		needOutput := map[string][]obs.OutputSpec{}

		// Make map of input and list of lokistacks for the input
		for _, i := range p.InputRefs {
			for _, outputName := range needMigrations.List() {
				needOutput[i] = append(needOutput[i], *outputMap[outputName])
			}
		}

		// Create pipeline for each tenant
		for i, input := range p.InputRefs {
			pOut := p.DeepCopy()
			pOut.InputRefs = []string{input}

			for i, output := range pOut.OutputRefs {
				if !needMigrations.Has(output) {
					// Leave output names as-is if not needing to migrate
					continue
				}
				// Format output name with input
				pOut.OutputRefs[i] = fmt.Sprintf("%s-%s", output, input)
			}

			// Generate pipeline name
			if pOut.Name != "" && i > 0 {
				pOut.Name = fmt.Sprintf("%s-%d", pOut.Name, i)
			}

			pipelines = append(pipelines, *pOut)
		}

		// Create output/s from each input
		for input, outputSpecList := range needOutput {
			tenant := getInputTypeFromName(spec, input)
			for _, outSpec := range outputSpecList {
				// Generate appropriate OTLP out or loki out
				obsSpec := GenerateOutput(outSpec, input, tenant)
				finalOutputs = append(finalOutputs, obsSpec)
			}
		}
	}

	// Sort outputs, because we have tests depending on the exact generated configuration
	sort.Slice(finalOutputs, func(i, j int) bool {
		return strings.Compare(finalOutputs[i].Name, finalOutputs[j].Name) < 0
	})

	return finalOutputs, pipelines
}

func getInputTypeFromName(spec obs.ClusterLogForwarderSpec, inputName string) string {
	if internalobs.ReservedInputTypes.Has(inputName) {
		// use name as type
		return inputName
	}

	for _, input := range spec.Inputs {
		if input.Name == inputName {
			if input.Application != nil {
				return string(obs.InputTypeApplication)
			}
			if input.Infrastructure != nil || input.Receiver.Type == obs.ReceiverTypeSyslog {
				return string(obs.InputTypeInfrastructure)
			}
			if input.Audit != nil || (input.Receiver.Type == obs.ReceiverTypeHTTP && input.Receiver.HTTP != nil && input.Receiver.HTTP.Format == obs.HTTPReceiverFormatKubeAPIAudit) {
				return string(obs.InputTypeAudit)
			}
		}
	}
	return ""
}

// findLokistackOutputs identifies and returns a set of lokistack output names that need to be migrated
func findLokistackOutputs(outputMap map[string]*obs.OutputSpec, outputRefs []string) *sets.String {
	needMigrations := sets.NewString()
	for _, outName := range outputRefs {
		if outSpec, ok := outputMap[outName]; ok {
			if outSpec.Type == obs.OutputTypeLokiStack {
				needMigrations.Insert(outSpec.Name)
			}
		}
	}
	return needMigrations
}
