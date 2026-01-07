package metrics

import (
	"strings"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/transforms"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

func NewUnmatched(id string, op utils.Options, labels map[string]string) types.Transform {
	logToMetric := transforms.NewLogToMetric(
		"component_event_unmatched_count",
		transforms.MetricsTypeCounter,
		transforms.Tags{
			"log_type":     "{{ log_type }}",
			"log_source":   "{{ log_source }}",
			"output_type":  strings.ToLower(string(obs.OutputTypeLokiStack)),
			"component_id": id,
		},
		transforms.UnmatchedRoute(id),
	)
	logToMetric.Metrics[0].Tags.AddAll(labels)
	unmatchedID := vectorhelpers.MakeID(id, transforms.Unmatched)
	op.AddToStringSet(framework.OptionLogsToMetricInputs, unmatchedID)
	return logToMetric
}
