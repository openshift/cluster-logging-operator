package telemetry

import (
	"context"

	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	labelResourceNamespace = "resource_namespace"
	labelResourceName      = "resource_name"

	labelManagedStatus = "managedStatus"
	labelHealthStatus  = "healthStatus"
	labelVersion       = "version"
	labelPipelineInfo  = "pipelineInfo"

	metricsPrefix     = "log_"
	defaultOutputName = "default"
)

var (
	forwarderInputTypes = []string{
		loggingv1.InputNameAudit,
		loggingv1.InputNameApplication,
		loggingv1.InputNameInfrastructure,
	}
	forwarderOutputTypes = []string{
		loggingv1.OutputTypeElasticsearch,
		loggingv1.OutputTypeFluentdForward,
		loggingv1.OutputTypeSyslog,
		loggingv1.OutputTypeKafka,
		loggingv1.OutputTypeLoki,
		loggingv1.OutputTypeCloudwatch,
		loggingv1.OutputTypeHttp,
		loggingv1.OutputTypeGoogleCloudLogging,
		loggingv1.OutputTypeSplunk,
	}

	clusterLoggingInfoDesc = prometheus.NewDesc(
		metricsPrefix+"logging_info",
		"Info metric containing general information about installed operator. Value is always 1.",
		[]string{labelResourceNamespace, labelResourceName, labelVersion, labelManagedStatus, labelHealthStatus}, nil,
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
		append([]string{labelResourceNamespace, labelResourceName, defaultOutputName}, forwarderOutputTypes...), nil,
	)

	// data will contain the singleton collector instance used by the public helper functions.
	// It is initialized by Setup. It's safe to use the helper functions before calling Setup.
	data *telemetryCollector
)

// Setup initializes the telemetry collector and registers it with the given Prometheus registry.
func Setup(ctx context.Context, k8sClient client.Client, registry prometheus.Registerer, version string) error {
	collector := newTelemetryCollector(ctx, k8sClient, version)

	if err := registry.Register(collector); err != nil {
		return err
	}

	data = collector
	return nil
}

// IncreaseCollectorErrors increases the collector error count metric by one.
func IncreaseCollectorErrors() {
	if data == nil {
		return
	}

	data.collectorErrors.Inc()
}

// UpdateDefaultForwarderInfo updates the data used for the telemetry metrics about the "openshift-logging/instance"
// ClusterLogForwarder. This currently needs a separate code path, because this resource might not really exist
// in the Kubernetes cluster when it is generated from data in ClusterLogging.
//
// It automatically discards forwarders which are not the "singleton instance", so this function can be called with
// any ClusterLogForwarder instance.
func UpdateDefaultForwarderInfo(forwarder *loggingv1.ClusterLogForwarder) {
	if data == nil {
		return
	}

	data.updateDefaultInfo(forwarder)
}
