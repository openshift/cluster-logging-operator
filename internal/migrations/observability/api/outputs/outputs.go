package outputs

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/migrations/observability/api/outputs/azuremonitor"
	"github.com/openshift/cluster-logging-operator/internal/migrations/observability/api/outputs/cloudwatch"
	"github.com/openshift/cluster-logging-operator/internal/migrations/observability/api/outputs/common"
	"github.com/openshift/cluster-logging-operator/internal/migrations/observability/api/outputs/elasticsearch"
	"github.com/openshift/cluster-logging-operator/internal/migrations/observability/api/outputs/gcl"
	"github.com/openshift/cluster-logging-operator/internal/migrations/observability/api/outputs/http"
	"github.com/openshift/cluster-logging-operator/internal/migrations/observability/api/outputs/kafka"
	"github.com/openshift/cluster-logging-operator/internal/migrations/observability/api/outputs/loki"
	"github.com/openshift/cluster-logging-operator/internal/migrations/observability/api/outputs/splunk"
	"github.com/openshift/cluster-logging-operator/internal/migrations/observability/api/outputs/syslog"
	corev1 "k8s.io/api/core/v1"
)

func ConvertOutputs(loggingClfSpec *logging.ClusterLogForwarderSpec, secrets map[string]*corev1.Secret) []obs.OutputSpec {
	obsOutputs := []obs.OutputSpec{}
	for _, output := range loggingClfSpec.Outputs {

		obsOut := &obs.OutputSpec{}
		obsOut.Name = output.Name
		obsOut.Type = obs.OutputType(output.Type)

		switch obsOut.Type {
		case obs.OutputTypeAzureMonitor:
			obsOut.AzureMonitor = azuremonitor.MapAzureMonitor(output, secrets[output.Name])
		case obs.OutputTypeCloudwatch:
			obsOut.Cloudwatch = cloudwatch.MapCloudwatch(output, secrets[output.Name])
		case obs.OutputTypeElasticsearch:
			obsOut.Elasticsearch = elasticsearch.MapElasticsearch(output, secrets[output.Name])
		case obs.OutputTypeGoogleCloudLogging:
			obsOut.GoogleCloudLogging = gcl.MapGoogleCloudLogging(output, secrets[output.Name])
		case obs.OutputTypeHTTP:
			obsOut.HTTP = http.MapHTTP(output, secrets[output.Name])
		case obs.OutputTypeKafka:
			obsOut.Kafka = kafka.MapKafka(output, secrets[output.Name])
		case obs.OutputTypeLoki:
			obsOut.Loki = loki.MapLoki(output, secrets[output.Name])
		case obs.OutputTypeSplunk:
			obsOut.Splunk = splunk.MapSplunk(output, secrets[output.Name])
		case obs.OutputTypeSyslog:
			obsOut.Syslog = syslog.MapSyslog(output)
		}
		// Limits
		if output.Limit != nil {
			obsOut.Limit = &obs.LimitSpec{
				MaxRecordsPerSecond: output.Limit.MaxRecordsPerSecond,
			}
		}

		// TLS Settings
		obsOut.TLS = common.MapOutputTls(output.TLS, secrets[output.Name])

		// Add output to obs clf
		obsOutputs = append(obsOutputs, *obsOut)
	}
	// Set observability CLF outputs to converted outputs
	return obsOutputs
}

// ReferencesFluentDForward determines if FluentDForward is a defined output
func ReferencesFluentDForward(loggingClfSpec *logging.ClusterLogForwarderSpec) bool {
	for _, output := range loggingClfSpec.Outputs {
		if output.Type == logging.OutputTypeFluentdForward {
			return true
		}
	}
	return false
}
