package outputs

import (
	"fmt"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/aws/cloudwatch"
)

func validateCloudwatchMaxWrite(output obs.OutputSpec) (results []string) {
	if output.Type != obs.OutputTypeCloudwatch || output.Cloudwatch == nil {
		return results
	}
	if output.Cloudwatch.Tuning == nil || output.Cloudwatch.Tuning.MaxWrite == nil {
		return results
	}
	maxWrite := output.Cloudwatch.Tuning.MaxWrite.Value()
	if maxWrite > cloudwatch.CloudwatchDefaultMaxBytes {
		results = append(results, fmt.Sprintf("maxWrite must not exceed %d bytes for CloudWatch, got %d bytes", cloudwatch.CloudwatchDefaultMaxBytes, maxWrite))
	}
	return results
}
