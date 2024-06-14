package telemetry

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	metricsPrefix = "log_"

	labelVersion = "version"

	labelResourceNamespace = "resource_namespace"
	labelResourceName      = "resource_name"

	labelHealthStatus = "healthStatus"
	labelDeployed     = "deployed"

	labelInput  = "input"
	labelOutput = "output"
)

var (
	logFileMetricExporterInfoDesc = prometheus.NewDesc(
		metricsPrefix+"file_metric_exporter_info",
		"Info metric containing information about usage the file metric exporter. Value is always 1.",
		[]string{labelVersion, labelDeployed, labelHealthStatus}, nil,
	)

	forwarderPipelinesDesc = prometheus.NewDesc(
		metricsPrefix+"forwarder_pipelines",
		"Metric counting the number of pipelines in a forwarder.",
		[]string{labelVersion, labelResourceNamespace, labelResourceName}, nil,
	)
	forwarderInputTypeDesc = prometheus.NewDesc(
		metricsPrefix+"forwarder_input_type",
		"Shows which input types a forwarder uses.",
		[]string{labelVersion, labelResourceNamespace, labelResourceName, labelInput}, nil,
	)
	forwarderOutputTypeDesc = prometheus.NewDesc(
		metricsPrefix+"forwarder_output_type",
		"Shows which output types a forwarder uses.",
		[]string{labelVersion, labelResourceNamespace, labelResourceName, labelOutput}, nil,
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
