package telemetry

import (
	"context"

	observabilityv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	labelResourceNamespace = "resource_namespace"
	labelResourceName      = "resource_name"

	labelHealthStatus = "healthStatus"
	labelPipelineInfo = "pipelineInfo"
	labelDeployed     = "deployed"

	metricsPrefix = "log_"
)

var (
	forwarderInputTypes  = convertInputs(observabilityv1.InputTypes)
	forwarderOutputTypes = convertOutputs(observabilityv1.OutputTypes)

	logFileMetricExporterInfoDesc = prometheus.NewDesc(
		metricsPrefix+"file_metric_exporter_info",
		"Info metric containing information about usage the file metric exporter. Value is always 1.",
		[]string{labelDeployed, labelHealthStatus}, nil,
	)

	clusterLogForwarderDesc = prometheus.NewDesc(
		metricsPrefix+"forwarder_pipeline_info",
		"Info metric containing information about deployed forwarders. Value is always 1.",
		[]string{labelResourceNamespace, labelResourceName, labelHealthStatus, labelPipelineInfo}, nil,
	)
	forwarderInputInfoDesc = prometheus.NewDesc(
		metricsPrefix+"forwarder_input_info",
		"Info metric containing information about usage of pre-defined input names. Value is always 1.",
		append([]string{labelResourceNamespace, labelResourceName}, forwarderInputTypes...), nil,
	)
	forwarderOutputInfoDesc = prometheus.NewDesc(
		metricsPrefix+"forwarder_output_info",
		"Info metric containing information about usage of available output types. Value is always 1.",
		append([]string{labelResourceNamespace, labelResourceName}, forwarderOutputTypes...), nil,
	)
)

// Setup initializes the telemetry collector and registers it with the given Prometheus registry.
func Setup(ctx context.Context, k8sClient client.Client, registry prometheus.Registerer, version string) error {
	collector := newTelemetryCollector(ctx, k8sClient, version)

	if err := registry.Register(collector); err != nil {
		return err
	}

	return nil
}

func convertInputs(types []observabilityv1.InputType) []string {
	result := make([]string, len(types))
	for i, t := range types {
		result[i] = string(t)
	}
	return result
}

func convertOutputs(types []observabilityv1.OutputType) []string {
	result := make([]string, len(types))
	for i, t := range types {
		result[i] = string(t)
	}
	return result
}
