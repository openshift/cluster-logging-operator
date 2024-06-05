package observability

import (
	"fmt"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MigrateLokiStack migrates a lokistack output into appropriate loki outputs based on defined inputs
func MigrateLokiStack(spec obs.ClusterLogForwarder, options utils.Options) (obs.ClusterLogForwarder, []metav1.Condition) {
	var outputs []obs.OutputSpec
	var pipelines []obs.PipelineSpec
	var migrationConditions []metav1.Condition
	outputs, pipelines, migrationConditions = ProcessForwarderPipelines(spec.Spec, obs.OutputTypeLokiStack, "", false)

	spec.Spec.Outputs = outputs
	spec.Spec.Pipelines = pipelines

	return spec, migrationConditions
}

func GenerateLokiOutput(outSpec obs.OutputSpec, input, tenant string) obs.OutputSpec {
	return obs.OutputSpec{
		Name: fmt.Sprintf("%s-%s", outSpec.Name, input),
		Type: obs.OutputTypeLoki,
		Loki: &obs.Loki{
			URLSpec: obs.URLSpec{
				URL: lokiStackURL(outSpec.LokiStack, tenant),
			},
			Authentication: outSpec.LokiStack.Authentication,
			Tuning:         outSpec.LokiStack.Tuning,
		},
		TLS:   outSpec.TLS,
		Limit: outSpec.Limit,
	}
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
