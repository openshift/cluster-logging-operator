package outputs

import (
	"fmt"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/azure/azurelogsingestion"
)

func validateAzureLogsIngestionMaxWrite(output obs.OutputSpec) (results []string) {
	if output.Type != obs.OutputTypeAzureLogsIngestion || output.AzureLogsIngestion == nil {
		return results
	}
	if output.AzureLogsIngestion.Tuning == nil || output.AzureLogsIngestion.Tuning.MaxWrite == nil {
		return results
	}
	maxWrite := output.AzureLogsIngestion.Tuning.MaxWrite.Value()
	if maxWrite > azurelogsingestion.AzureDefaultMaxBytes {
		results = append(results, fmt.Sprintf("maxWrite must not exceed %d bytes for AzureLogsIngestion, got %d bytes", azurelogsingestion.AzureDefaultMaxBytes, maxWrite))
	}
	return results
}
