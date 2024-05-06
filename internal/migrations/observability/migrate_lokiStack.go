package observability

import (
	"fmt"
	"sort"
	"strings"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MigrateLokiStack migrates a lokistack output into appropriate loki outputs based on defined inputs
func MigrateLokiStack(spec obs.ClusterLogForwarderSpec) (obs.ClusterLogForwarderSpec, []metav1.Condition) {
	var outputs []obs.OutputSpec
	var pipelines []obs.PipelineSpec
	var migrationConditions []metav1.Condition
	outputs, pipelines, migrationConditions = processForwarderPipelines(spec)

	spec.Outputs = outputs
	spec.Pipelines = pipelines

	return spec, migrationConditions
}

// findLokiStackOutputs identifies and returns a set of lokistack output names that need to be migrated
func findLokiStackOutputs(outputMap map[string]*obs.OutputSpec, outputRefs []string) *sets.String {
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

// processForwarderPipelines migrates a lokistack output into separate pipelines based on inputRefs
func processForwarderPipelines(spec obs.ClusterLogForwarderSpec) ([]obs.OutputSpec, []obs.PipelineSpec, []metav1.Condition) {
	inPipelines := spec.Pipelines
	pipelines := []obs.PipelineSpec{}
	outputMap := utils.OutputMap(&spec)
	migratedLokiStackConditions := []metav1.Condition{}

	finalOutputs := []obs.OutputSpec{}

	// Remove lokistacks that will be migrated and add a condition message
	for _, o := range spec.Outputs {
		if o.Type == obs.OutputTypeLokiStack {
			migratedLokiStackConditions = append(migratedLokiStackConditions,
				metav1.Condition{
					Type:    obs.ConditionMigrate,
					Status:  metav1.ConditionTrue,
					Reason:  obs.ReasonMigrateOutput,
					Message: fmt.Sprintf("lokistack: %q migrated to loki output/s", o.Name),
				})
			continue
		}
		finalOutputs = append(finalOutputs, o)
	}

	for _, p := range inPipelines {
		needMigrations := findLokiStackOutputs(outputMap, p.OutputRefs)
		// Ignore migration if lokistack not referenced in outputRef of pipeline
		if needMigrations.Len() == 0 {
			pipelines = append(pipelines, p)
			continue
		}

		needOutput := map[string][]obs.OutputSpec{}

		// Make map of input and list of lokistacks for the input
		for _, i := range p.InputRefs {
			for _, lokistackName := range needMigrations.List() {
				needOutput[i] = append(needOutput[i], *outputMap[lokistackName])
			}
		}

		// Create pipeline for each tenant
		for i, input := range p.InputRefs {
			pOut := p.DeepCopy()
			pOut.InputRefs = []string{input}

			for i, output := range pOut.OutputRefs {
				if !needMigrations.Has(output) {
					// Leave non-lokistack output names as-is
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
				finalOutputs = append(finalOutputs, obs.OutputSpec{
					Name: fmt.Sprintf("%s-%s", outSpec.Name, input),
					Type: obs.OutputTypeLoki,
					Loki: &obs.Loki{
						URLSpec: obs.URLSpec{
							URL: lokiStackURL(outSpec.LokiStack, tenant),
						},
						Authentication: outSpec.LokiStack.Authentication,
						Tuning:         outSpec.LokiStack.Tuning,
					},
					Tuning: outSpec.Tuning,
					TLS:    outSpec.TLS,
					Limit:  outSpec.Limit,
				})
			}
		}
	}

	// Sort outputs, because we have tests depending on the exact generated configuration
	sort.Slice(finalOutputs, func(i, j int) bool {
		return strings.Compare(finalOutputs[i].Name, finalOutputs[j].Name) < 0
	})

	return finalOutputs, pipelines, migratedLokiStackConditions
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
			if input.Audit != nil || (input.Receiver.Type == obs.ReceiverTypeHttp && input.Receiver.HTTP != nil && input.Receiver.HTTP.Format == obs.HTTPReceiverFormatKubeAPIAudit) {
				return string(obs.InputTypeAudit)
			}
		}
	}
	return ""
}

// lokiStackURL returns the URL of the LokiStack API for a specific tenant.
func lokiStackURL(lokiStackSpec *obs.LokiStack, tenant string) string {
	service := LokiStackGatewayService(lokiStackSpec.Target.Name)
	if !obs.ReservedInputTypes.Has(tenant) {
		return ""
	}
	return fmt.Sprintf("https://%s.%s.svc:8080/api/logs/v1/%s", service, lokiStackSpec.Target.Namespace, tenant)
}

// LokiStackGatewayService returns the name of LokiStack gateway service.
func LokiStackGatewayService(lokiStackServiceName string) string {
	return fmt.Sprintf("%s-gateway-http", lokiStackServiceName)
}
