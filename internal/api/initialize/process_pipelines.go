package initialize

import (
	"fmt"
	"sort"
	"strings"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
)

// ProcessForwarderPipelines migrates the specified output type and name to appropriate outputs and pipelines
func ProcessForwarderPipelines(spec obs.ClusterLogForwarderSpec, wantedOutputType obs.OutputType, wantedOutputName string, findByTypeAndName bool) ([]obs.OutputSpec, []obs.PipelineSpec) {
	inPipelines := spec.Pipelines
	pipelines := []obs.PipelineSpec{}
	outputMap := utils.OutputMap(&spec)
	finalOutputs := []obs.OutputSpec{}

	// Remove migrated outputs
	for _, o := range spec.Outputs {
		if o.Type == wantedOutputType && (!findByTypeAndName || o.Name == wantedOutputName) {
			continue
		}
		finalOutputs = append(finalOutputs, o)
	}

	for _, p := range inPipelines {
		needMigrations := findSpecifiedOutputs(outputMap, p.OutputRefs, wantedOutputType, wantedOutputName, findByTypeAndName)
		if needMigrations.Len() == 0 {
			pipelines = append(pipelines, p)
			continue
		}

		needOutput := map[string][]obs.OutputSpec{}

		// Make map of input and list of specified outputs for the input
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
				var obsSpec obs.OutputSpec
				switch wantedOutputType {
				case obs.OutputTypeElasticsearch:
					obsSpec = GenerateESOutput(outSpec, input, tenant)
				case obs.OutputTypeLokiStack:
					obsSpec = GenerateLokiOutput(outSpec, input, tenant)
				}
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
	if obs.ReservedInputTypes.Has(inputName) {
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

// findSpecifiedOutputs determines if the wanted output type and/or name is referenced and needs to be migrated
func findSpecifiedOutputs(outputMap map[string]*obs.OutputSpec, outputRefs []string, wantedOutputType obs.OutputType, wantedOutputName string, findByNameAndType bool) *sets.String {
	needMigrations := sets.NewString()
	for _, outName := range outputRefs {
		if outSpec, ok := outputMap[outName]; ok && outSpec.Type == wantedOutputType {
			if !findByNameAndType || outSpec.Name == wantedOutputName {
				needMigrations.Insert(outSpec.Name)
			}
		}
	}
	return needMigrations
}
