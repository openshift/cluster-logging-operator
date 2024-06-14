package telemetry

import (
	"context"
	"strings"

	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	observabilityv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/status"
	"github.com/openshift/cluster-logging-operator/internal/validations/clusterlogforwarder/conditions"

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
)

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
			It("should only provide LFME metric", func() {
				wantMetrics := `# HELP log_file_metric_exporter_info Info metric containing information about usage the file metric exporter. Value is always 1.
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

		Context("with observabilityv1 ClusterLogForwarder present", func() {
			It("should provide CLF and LFME metrics", func() {
				clf := &observabilityv1.ClusterLogForwarder{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test-namespace",
						Name:      "test-name",
					},
					Spec: observabilityv1.ClusterLogForwarderSpec{
						Outputs: []observabilityv1.OutputSpec{
							{
								Name: "output",
								Type: observabilityv1.OutputTypeLokiStack,
							},
						},
						Pipelines: []observabilityv1.PipelineSpec{
							{
								Name: "pipeline",
								InputRefs: []string{
									"application",
									"infrastructure",
								},
								OutputRefs: []string{
									"output",
								},
							},
						},
					},
					Status: observabilityv1.ClusterLogForwarderStatus{
						Conditions: []metav1.Condition{
							{
								Type:   observabilityv1.ConditionTypeReady,
								Status: metav1.ConditionTrue,
							},
						},
					},
				}
				wantMetrics := `# HELP log_file_metric_exporter_info Info metric containing information about usage the file metric exporter. Value is always 1.
# TYPE log_file_metric_exporter_info gauge
log_file_metric_exporter_info{deployed="0",healthStatus="0"} 1
# HELP log_forwarder_input_info Info metric containing information about usage of pre-defined input names. Value is always 1.
# TYPE log_forwarder_input_info gauge
log_forwarder_input_info{application="1",audit="0",infrastructure="1",receiver="0",resource_name="test-name",resource_namespace="test-namespace"} 1
# HELP log_forwarder_output_info Info metric containing information about usage of available output types. Value is always 1.
# TYPE log_forwarder_output_info gauge
log_forwarder_output_info{azureMonitor="0",cloudwatch="0",elasticsearch="0",googleCloudLogging="0",http="0",kafka="0",loki="0",lokiStack="1",otlp="0",resource_name="test-name",resource_namespace="test-namespace",splunk="0",syslog="0"} 1
# HELP log_forwarder_pipeline_info Info metric containing information about deployed forwarders. Value is always 1.
# TYPE log_forwarder_pipeline_info gauge
log_forwarder_pipeline_info{healthStatus="1",pipelineInfo="1",resource_name="test-name",resource_namespace="test-namespace"} 1
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
							conditions.CondReady,
						},
					},
				}
				wantMetrics := `# HELP log_file_metric_exporter_info Info metric containing information about usage the file metric exporter. Value is always 1.
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
				wantMetrics := `# HELP log_file_metric_exporter_info Info metric containing information about usage the file metric exporter. Value is always 1.
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
				wantMetrics := `# HELP log_file_metric_exporter_info Info metric containing information about usage the file metric exporter. Value is always 1.
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
	})
})
