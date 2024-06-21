package otlp

import (
	"fmt"
	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"strings"
)

type Reduce struct {
	ComponentID string
	Desc        string
	Inputs      string
	GroupBy     string
	MaxEvents   string
}

func (r Reduce) Name() string {
	return "reduceTemplate"
}

func (r Reduce) Template() string {
	return `{{define "reduceTemplate" -}}
{{if .Desc -}}
# {{.Desc}}
{{end -}}
[transforms.{{.ComponentID}}]
type = "reduce"
inputs = {{.Inputs}}
expire_after_ms = 15000
max_events = {{.MaxEvents}}
group_by = {{.GroupBy}}
merge_strategies.resource = "retain"
merge_strategies.logRecords = "array"
{{end}}
`
}

func GroupByContainer(id string, inputs []string) Element {
	return Reduce{
		Desc:        "Merge container logs and group by namespace, pod and container",
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
		MaxEvents:   "250",
		GroupBy: MakeGroupBys(".openshift.cluster_id",
			".kubernetes.namespace_name", ".kubernetes.pod_name", ".kubernetes.container_name"),
	}
}

func GroupBySource(id string, inputs []string) Element {
	return Reduce{
		Desc:        "Merge audit api and node logs and group by log_source",
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
		MaxEvents:   "250",
		GroupBy:     MakeGroupBys(".openshift.cluster_id", ".openshift.log_source"),
	}
}

func GroupByHost(id string, inputs []string) Element {
	return Reduce{
		Desc:        "Merge auditd host logs and group by hostname",
		ComponentID: id,
		Inputs:      helpers.MakeInputs(inputs...),
		MaxEvents:   "50",
		GroupBy:     MakeGroupBys(".openshift.cluster_id", ".openshift.hostname"),
	}
}

func MakeGroupBys(fields ...string) string {
	out := make([]string, len(fields))
	for i, o := range fields {
		out[i] = fmt.Sprintf("%q", o)
	}
	return fmt.Sprintf("[%s]", strings.Join(out, ","))
}
