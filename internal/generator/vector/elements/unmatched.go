package elements

import (
	"fmt"
	"strings"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/transforms/logtometric"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

type Unmatched struct {
	api.Config
}

func NewUnmatched(id string, op utils.Options, labels map[string]string) Unmatched {
	logToMetric := logtometric.New(
		"component_event_unmatched_count",
		logtometric.MetricsTypeCounter,
		logtometric.Tags{
			"log_type":     "{{ log_type }}",
			"log_source":   "{{ log_source }}",
			"output_type":  strings.ToLower(string(obs.OutputTypeLokiStack)),
			"component_id": id,
		},
		fmt.Sprintf("%s._unmatched", id),
	)
	logToMetric.Metrics[0].Tags.AddAll(labels)
	unmatchedID := vectorhelpers.MakeID(id, "unmatched")
	op.AddToStringSet(framework.OptionLogsToMetricInputs, unmatchedID)
	u := Unmatched{
		Config: api.Config{
			Transforms: map[string]interface{}{
				unmatchedID: logToMetric,
			},
		},
	}
	return u
}

func (r Unmatched) Name() string {
	return r.Config.Name()
}

func (r Unmatched) Template() string {
	return r.Config.Template()
}
