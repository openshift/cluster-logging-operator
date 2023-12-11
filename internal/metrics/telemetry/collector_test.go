package telemetry

import (
	"context"
	"strings"

	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/status"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	testVersion = "test-version"
)

var (
	singletonMeta = metav1.ObjectMeta{
		Name:      constants.SingletonName,
		Namespace: constants.OpenshiftNS,
	}

	testDefaultLokiForwarder = &loggingv1.ClusterLogForwarder{
		ObjectMeta: singletonMeta,
		Spec: loggingv1.ClusterLogForwarderSpec{
			Outputs: []loggingv1.OutputSpec{
				{
					Name: "default-loki-apps",
					Type: loggingv1.OutputTypeLoki,
				},
				{
					Name: "default-loki-infra",
					Type: loggingv1.OutputTypeLoki,
				},
			},
			Pipelines: []loggingv1.PipelineSpec{
				{
					OutputRefs: []string{
						"default-loki-apps",
					},
					InputRefs: []string{
						loggingv1.InputNameApplication,
					},
				},
				{
					OutputRefs: []string{
						"default-loki-infra",
					},
					InputRefs: []string{
						loggingv1.InputNameInfrastructure,
					},
				},
			},
		},
	}
)

func testReadyClusterLogging(logStoreType loggingv1.LogStoreType) *loggingv1.ClusterLogging {
	return &loggingv1.ClusterLogging{
		ObjectMeta: singletonMeta,
		Spec: loggingv1.ClusterLoggingSpec{
			ManagementState: loggingv1.ManagementStateManaged,
			LogStore: &loggingv1.LogStoreSpec{
				Type: logStoreType,
			},
		},
		Status: loggingv1.ClusterLoggingStatus{
			Conditions: []status.Condition{
				loggingv1.CondReady,
			},
		},
	}
}

// Test if ServiceMonitor spec is correct or not, also Prometheus Metrics get Registered, Updated, Retrieved properly or not
var _ = Describe("Telemetry Collector", func() {
	When("Registering it to a Prometheus registry", func() {
		It("should not return an error", func() {
			k8s := fake.NewFakeClient()
			collector := newTelemetryCollector(context.Background(), k8s, testVersion)
			registry := prometheus.NewPedanticRegistry()

			err := registry.Register(collector)
			Expect(err).To(BeNil())
		})
	})

	When("Initializing a new collector", func() {
		Context("with no resources in the Kubernetes cluster", func() {
			It("should provide the collector errors and LFME metric", func() {
				wantMetrics := `# HELP log_collector_error_count_total Counts the number of errors encountered by the operator reconciling the collector configuration
# TYPE log_collector_error_count_total counter
log_collector_error_count_total{version="test-version"} 0
# HELP log_file_metric_exporter_info Info metric containing information about usage the file metric exporter. Value is always 1.
# TYPE log_file_metric_exporter_info gauge
log_file_metric_exporter_info{deployed="0",healthStatus="0"} 1
`

				ctx := context.Background()
				k8s := fake.NewFakeClient()
				collector := newTelemetryCollector(ctx, k8s, testVersion)

				metricsReader := strings.NewReader(wantMetrics)
				err := testutil.CollectAndCompare(collector, metricsReader)
				Expect(err).To(BeNil())
			})
		})

		Context("with ClusterLogging present", func() {
			It("should provide all metrics", func() {
				cl := &loggingv1.ClusterLogging{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "instance",
						Namespace: "openshift-logging",
					},
					Spec: loggingv1.ClusterLoggingSpec{
						ManagementState: loggingv1.ManagementStateUnmanaged,
					},
				}
				wantMetrics := `# HELP log_collector_error_count_total Counts the number of errors encountered by the operator reconciling the collector configuration
# TYPE log_collector_error_count_total counter
log_collector_error_count_total{version="test-version"} 0
# HELP log_file_metric_exporter_info Info metric containing information about usage the file metric exporter. Value is always 1.
# TYPE log_file_metric_exporter_info gauge
log_file_metric_exporter_info{deployed="0",healthStatus="0"} 1
# HELP log_forwarder_input_info Info metric containing information about usage of pre-defined input names. Value is always 1.
# TYPE log_forwarder_input_info gauge
log_forwarder_input_info{application="0",audit="0",http="0",infrastructure="0",resource_name="instance",resource_namespace="openshift-logging",syslog="0"} 1
# HELP log_forwarder_output_info Info metric containing information about usage of available output types. Value is always 1.
# TYPE log_forwarder_output_info gauge
log_forwarder_output_info{azureMonitor="0",cloudwatch="0",default="0",elasticsearch="0",fluentdForward="0",googleCloudLogging="0",http="0",kafka="0",loki="0",resource_name="instance",resource_namespace="openshift-logging",splunk="0",syslog="0"} 1
# HELP log_forwarder_pipeline_info Info metric containing information about deployed forwarders. Value is always 1.
# TYPE log_forwarder_pipeline_info gauge
log_forwarder_pipeline_info{healthStatus="0",pipelineInfo="0",resource_name="instance",resource_namespace="openshift-logging"} 1
# HELP log_logging_info Info metric containing general information about installed operator. Value is always 1.
# TYPE log_logging_info gauge
log_logging_info{healthStatus="0",managedStatus="0",resource_name="instance",resource_namespace="openshift-logging",version="test-version"} 1
`

				ctx := context.Background()
				k8s := fake.NewFakeClient(cl)
				collector := newTelemetryCollector(ctx, k8s, testVersion)

				metricsReader := strings.NewReader(wantMetrics)
				err := testutil.CollectAndCompare(collector, metricsReader)
				Expect(err).To(BeNil())
			})
		})

		Context("with ClusterLogForwarder present", func() {
			It("should provide only CLF and LFME metrics", func() {
				clf := &loggingv1.ClusterLogForwarder{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test-namespace",
						Name:      "test-name",
					},
				}
				wantMetrics := `# HELP log_file_metric_exporter_info Info metric containing information about usage the file metric exporter. Value is always 1.
# TYPE log_file_metric_exporter_info gauge
log_file_metric_exporter_info{deployed="0",healthStatus="0"} 1
# HELP log_collector_error_count_total Counts the number of errors encountered by the operator reconciling the collector configuration
# TYPE log_collector_error_count_total counter
log_collector_error_count_total{version="test-version"} 0
# HELP log_forwarder_input_info Info metric containing information about usage of pre-defined input names. Value is always 1.
# TYPE log_forwarder_input_info gauge
log_forwarder_input_info{application="0",audit="0",http="0",infrastructure="0",resource_name="test-name",resource_namespace="test-namespace",syslog="0"} 1
# HELP log_forwarder_output_info Info metric containing information about usage of available output types. Value is always 1.
# TYPE log_forwarder_output_info gauge
log_forwarder_output_info{azureMonitor="0",cloudwatch="0",default="0",elasticsearch="0",fluentdForward="0",googleCloudLogging="0",http="0",kafka="0",loki="0",resource_name="test-name",resource_namespace="test-namespace",splunk="0",syslog="0"} 1
# HELP log_forwarder_pipeline_info Info metric containing information about deployed forwarders. Value is always 1.
# TYPE log_forwarder_pipeline_info gauge
log_forwarder_pipeline_info{healthStatus="0",pipelineInfo="0",resource_name="test-name",resource_namespace="test-namespace"} 1
`

				ctx := context.Background()
				k8s := fake.NewFakeClient(clf)
				collector := newTelemetryCollector(ctx, k8s, testVersion)

				metricsReader := strings.NewReader(wantMetrics)
				err := testutil.CollectAndCompare(collector, metricsReader)
				Expect(err).To(BeNil())
			})
		})

		Context("with LogFileMetricExporter present", func() {
			It("should show a ready and deployed exporter", func() {
				lfme := &loggingv1alpha1.LogFileMetricExporter{
					ObjectMeta: singletonMeta,
					Status: loggingv1alpha1.LogFileMetricExporterStatus{
						Conditions: []status.Condition{
							loggingv1.CondReady,
						},
					},
				}
				wantMetrics := `# HELP log_collector_error_count_total Counts the number of errors encountered by the operator reconciling the collector configuration
# TYPE log_collector_error_count_total counter
log_collector_error_count_total{version="test-version"} 0
# HELP log_file_metric_exporter_info Info metric containing information about usage the file metric exporter. Value is always 1.
# TYPE log_file_metric_exporter_info gauge
log_file_metric_exporter_info{deployed="1",healthStatus="1"} 1
`

				ctx := context.Background()
				k8s := fake.NewFakeClient(lfme)
				collector := newTelemetryCollector(ctx, k8s, testVersion)

				metricsReader := strings.NewReader(wantMetrics)
				err := testutil.CollectAndCompare(collector, metricsReader)
				Expect(err).To(BeNil())
			})

			It("should show an unready status", func() {
				lfme := &loggingv1alpha1.LogFileMetricExporter{
					ObjectMeta: singletonMeta,
				}
				wantMetrics := `# HELP log_collector_error_count_total Counts the number of errors encountered by the operator reconciling the collector configuration
# TYPE log_collector_error_count_total counter
log_collector_error_count_total{version="test-version"} 0
# HELP log_file_metric_exporter_info Info metric containing information about usage the file metric exporter. Value is always 1.
# TYPE log_file_metric_exporter_info gauge
log_file_metric_exporter_info{deployed="1",healthStatus="0"} 1
`

				ctx := context.Background()
				k8s := fake.NewFakeClient(lfme)
				collector := newTelemetryCollector(ctx, k8s, testVersion)

				metricsReader := strings.NewReader(wantMetrics)
				err := testutil.CollectAndCompare(collector, metricsReader)
				Expect(err).To(BeNil())
			})

			It("should ignore non-singleton instances", func() {
				lfme := &loggingv1alpha1.LogFileMetricExporter{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test-namespace",
						Name:      "test-name",
					},
				}
				wantMetrics := `# HELP log_collector_error_count_total Counts the number of errors encountered by the operator reconciling the collector configuration
# TYPE log_collector_error_count_total counter
log_collector_error_count_total{version="test-version"} 0
# HELP log_file_metric_exporter_info Info metric containing information about usage the file metric exporter. Value is always 1.
# TYPE log_file_metric_exporter_info gauge
log_file_metric_exporter_info{deployed="0",healthStatus="0"} 1
`

				ctx := context.Background()
				k8s := fake.NewFakeClient(lfme)
				collector := newTelemetryCollector(ctx, k8s, testVersion)

				metricsReader := strings.NewReader(wantMetrics)
				err := testutil.CollectAndCompare(collector, metricsReader)
				Expect(err).To(BeNil())
			})
		})

		Context("with all resources present", func() {
			It("should provide all metrics", func() {
				cl := testReadyClusterLogging(loggingv1.LogStoreTypeLokiStack)
				lfme := &loggingv1alpha1.LogFileMetricExporter{
					ObjectMeta: singletonMeta,
					Status: loggingv1alpha1.LogFileMetricExporterStatus{
						Conditions: []status.Condition{
							loggingv1.CondReady,
						},
					},
				}
				wantMetrics := `# HELP log_collector_error_count_total Counts the number of errors encountered by the operator reconciling the collector configuration
# TYPE log_collector_error_count_total counter
log_collector_error_count_total{version="test-version"} 0
# HELP log_file_metric_exporter_info Info metric containing information about usage the file metric exporter. Value is always 1.
# TYPE log_file_metric_exporter_info gauge
log_file_metric_exporter_info{deployed="1",healthStatus="1"} 1
# HELP log_forwarder_input_info Info metric containing information about usage of pre-defined input names. Value is always 1.
# TYPE log_forwarder_input_info gauge
log_forwarder_input_info{application="1",audit="0",http="0",infrastructure="1",resource_name="instance",resource_namespace="openshift-logging",syslog="0"} 1
# HELP log_forwarder_output_info Info metric containing information about usage of available output types. Value is always 1.
# TYPE log_forwarder_output_info gauge
log_forwarder_output_info{azureMonitor="0",cloudwatch="0",default="1",elasticsearch="0",fluentdForward="0",googleCloudLogging="0",http="0",kafka="0",loki="1",resource_name="instance",resource_namespace="openshift-logging",splunk="0",syslog="0"} 1
# HELP log_forwarder_pipeline_info Info metric containing information about deployed forwarders. Value is always 1.
# TYPE log_forwarder_pipeline_info gauge
log_forwarder_pipeline_info{healthStatus="1",pipelineInfo="2",resource_name="instance",resource_namespace="openshift-logging",} 1
# HELP log_logging_info Info metric containing general information about installed operator. Value is always 1.
# TYPE log_logging_info gauge
log_logging_info{healthStatus="1",managedStatus="1",resource_name="instance",resource_namespace="openshift-logging",version="test-version"} 1
`

				ctx := context.Background()
				k8s := fake.NewFakeClient(cl, testDefaultLokiForwarder, lfme)
				collector := newTelemetryCollector(ctx, k8s, testVersion)
				collector.updateDefaultInfo(testDefaultLokiForwarder)

				metricsReader := strings.NewReader(wantMetrics)
				err := testutil.CollectAndCompare(collector, metricsReader)
				Expect(err).To(BeNil())
			})
		})
	})
})
