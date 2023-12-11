package telemetry

import (
	"context"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"strings"

	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ClusterLogForwarder metrics", func() {
	When("Updating the data from a ClusterLogForwarder", func() {
		Context("with a default ElasticSearch output", func() {
			var (
				cl        = testReadyClusterLogging(loggingv1.LogStoreTypeElasticsearch)
				forwarder = &loggingv1.ClusterLogForwarder{
					ObjectMeta: singletonMeta,
					Spec: loggingv1.ClusterLogForwarderSpec{
						Outputs: []loggingv1.OutputSpec{
							{
								Name: defaultOutputName,
								Type: loggingv1.OutputTypeElasticsearch,
							},
						},
						Pipelines: []loggingv1.PipelineSpec{
							{
								OutputRefs: []string{
									defaultOutputName,
								},
								InputRefs: []string{
									loggingv1.InputNameApplication,
									loggingv1.InputNameInfrastructure,
								},
							},
						},
					},
				}

				wantMetrics = `# HELP log_collector_error_count_total Counts the number of errors encountered by the operator reconciling the collector configuration
# TYPE log_collector_error_count_total counter
log_collector_error_count_total{version="test-version"} 0
# HELP log_file_metric_exporter_info Info metric containing information about usage the file metric exporter. Value is always 1.
# TYPE log_file_metric_exporter_info gauge
log_file_metric_exporter_info{deployed="0",healthStatus="0"} 1
# HELP log_forwarder_input_info Info metric containing information about usage of pre-defined input names. Value is always 1.
# TYPE log_forwarder_input_info gauge
log_forwarder_input_info{application="1",audit="0",http="0",infrastructure="1",resource_name="instance",resource_namespace="openshift-logging",syslog="0"} 1
# HELP log_forwarder_output_info Info metric containing information about usage of available output types. Value is always 1.
# TYPE log_forwarder_output_info gauge
log_forwarder_output_info{azureMonitor="0",cloudwatch="0",default="1",elasticsearch="1",fluentdForward="0",googleCloudLogging="0",http="0",kafka="0",loki="0",resource_name="instance",resource_namespace="openshift-logging",splunk="0",syslog="0"} 1
# HELP log_forwarder_pipeline_info Info metric containing information about deployed forwarders. Value is always 1.
# TYPE log_forwarder_pipeline_info gauge
log_forwarder_pipeline_info{healthStatus="1",pipelineInfo="1",resource_name="instance",resource_namespace="openshift-logging"} 1
# HELP log_logging_info Info metric containing general information about installed operator. Value is always 1.
# TYPE log_logging_info gauge
log_logging_info{healthStatus="1",managedStatus="1",resource_name="instance",resource_namespace="openshift-logging",version="test-version"} 1
`
			)

			It("should update the metrics", func() {
				ctx := context.Background()
				k8s := fake.NewFakeClient(cl, forwarder)

				collector := newTelemetryCollector(ctx, k8s, testVersion)
				collector.updateDefaultInfo(forwarder)

				metricsReader := strings.NewReader(wantMetrics)
				err := testutil.CollectAndCompare(collector, metricsReader)
				Expect(err).To(BeNil())
			})
		})

		Context("with a default LokiStack output", func() {
			var (
				cl = testReadyClusterLogging(loggingv1.LogStoreTypeLokiStack)

				wantMetrics = `# HELP log_collector_error_count_total Counts the number of errors encountered by the operator reconciling the collector configuration
# TYPE log_collector_error_count_total counter
log_collector_error_count_total{version="test-version"} 0
# HELP log_file_metric_exporter_info Info metric containing information about usage the file metric exporter. Value is always 1.
# TYPE log_file_metric_exporter_info gauge
log_file_metric_exporter_info{deployed="0",healthStatus="0"} 1
# HELP log_forwarder_input_info Info metric containing information about usage of pre-defined input names. Value is always 1.
# TYPE log_forwarder_input_info gauge
log_forwarder_input_info{application="1",audit="0",http="0",infrastructure="1",resource_name="instance",resource_namespace="openshift-logging",syslog="0"} 1
# HELP log_forwarder_output_info Info metric containing information about usage of available output types. Value is always 1.
# TYPE log_forwarder_output_info gauge
log_forwarder_output_info{azureMonitor="0",cloudwatch="0",default="1",elasticsearch="0",fluentdForward="0",googleCloudLogging="0",http="0",kafka="0",loki="1",resource_name="instance",resource_namespace="openshift-logging",splunk="0",syslog="0"} 1
# HELP log_forwarder_pipeline_info Info metric containing information about deployed forwarders. Value is always 1.
# TYPE log_forwarder_pipeline_info gauge
log_forwarder_pipeline_info{healthStatus="1",pipelineInfo="2",resource_name="instance",resource_namespace="openshift-logging"} 1
# HELP log_logging_info Info metric containing general information about installed operator. Value is always 1.
# TYPE log_logging_info gauge
log_logging_info{healthStatus="1",managedStatus="1",resource_name="instance",resource_namespace="openshift-logging",version="test-version"} 1
`
			)

			It("should update the metrics", func() {
				ctx := context.Background()
				k8s := fake.NewFakeClient(cl, testDefaultLokiForwarder)

				collector := newTelemetryCollector(ctx, k8s, testVersion)
				collector.updateDefaultInfo(testDefaultLokiForwarder)

				metricsReader := strings.NewReader(wantMetrics)
				err := testutil.CollectAndCompare(collector, metricsReader)
				Expect(err).To(BeNil())
			})
		})
	})
})
