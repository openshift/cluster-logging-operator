package initialize

import (
	"fmt"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/migrations/observability/api"
)

const (
	AppIndex   = "app-write"
	InfraIndex = "infra-write"
	AuditIndex = "audit-write"
)

// DefaultElasticsearch migrates a default ES output into appropriate ES outputs based on defined inputs
func DefaultElasticsearch(spec obs.ClusterLogForwarderSpec) obs.ClusterLogForwarderSpec {
	var outputs []obs.OutputSpec
	var pipelines []obs.PipelineSpec
	outputs, pipelines = ProcessForwarderPipelines(spec, obs.OutputTypeElasticsearch, api.DefaultEsName, true)

	spec.Outputs = outputs
	spec.Pipelines = pipelines

	return spec
}

func GenerateESOutput(outSpec obs.OutputSpec, input, tenant string) obs.OutputSpec {
	return obs.OutputSpec{
		Name: fmt.Sprintf("%s-%s", outSpec.Name, input),
		Type: obs.OutputTypeElasticsearch,
		Elasticsearch: &obs.Elasticsearch{
			URLSpec: outSpec.Elasticsearch.URLSpec,
			Version: outSpec.Elasticsearch.Version,
			IndexSpec: obs.IndexSpec{
				Index: elasticsearchIndex(tenant),
			},
			Authentication: outSpec.Elasticsearch.Authentication,
			Tuning:         outSpec.Elasticsearch.Tuning,
		},
		TLS:   outSpec.TLS,
		Limit: outSpec.Limit,
	}
}

// elasticsearchIndex returns the index for the specified input type
func elasticsearchIndex(tenant string) string {
	switch tenant {
	case string(obs.InputTypeApplication):
		return AppIndex
	case string(obs.InputTypeAudit):
		return AuditIndex
	case string(obs.InputTypeInfrastructure):
		return InfraIndex
	}
	return "{{.log_type}}"
}
