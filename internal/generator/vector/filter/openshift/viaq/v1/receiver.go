package v1

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"strings"
)

func receiverLogs() string {
	return fmt.Sprintf(`
if ._internal.log_type == "%s" {
  %s
}
`, obs.InputTypeReceiver, receiverLogsVRL())
}

func receiverLogsVRL() string {
	return strings.Join(helpers.TrimSpaces([]string{
		`.message = ._internal.message`,
		fmt.Sprintf(`.log_type = "%s"`, obs.InputTypeInfrastructure),
		fmt.Sprintf(`.log_source = "%s"`, obs.InfrastructureSourceNode),
	}), "\n\n")
}
