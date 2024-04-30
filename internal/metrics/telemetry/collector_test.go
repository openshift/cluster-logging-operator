package telemetry

import (
	"context"
	"strings"

	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/status"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

	testCondReady = status.Condition{
		Type:   loggingv1.ConditionReady,
		Status: corev1.ConditionTrue,
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
		Status: loggingv1.ClusterLogForwarderStatus{
			Conditions: []status.Condition{
				testCondReady,
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
				testCondReady,
			},
		},
	}
}

// Test if ServiceMonitor spec is correct or not, also Prometheus Metrics get Registered, Updated, Retrieved properly or not
var _ = Describe("Telemetry Collector", func() {
	When("Registering it to a Prometheus registry", func() {
		It("should not return an error", func() {
			k8s := newFakeClient()
			collector := newTelemetryCollector(context.Background(), k8s, testVersion)
			registry := prometheus.NewPedanticRegistry()

			err := registry.Register(collector)
			Expect(err).To(BeNil())
		})
	})

	When("Initializing a new collector", func() {
		Context("with no resources in the Kubernetes cluster", func() {
			It("should provide the collector errors metric", func() {
				wantMetrics := `# HELP log_collector_error_count_total Counts the number of errors encountered by the operator reconciling the collector configuration
# TYPE log_collector_error_count_total counter
log_collector_error_count_total{version="test-version"} 0
`

				ctx := context.Background()
				k8s := newFakeClient()
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
# HELP log_forwarder_input_info Info metric containing information about usage of pre-defined input names. Value is always 1.
# TYPE log_forwarder_input_info gauge
log_forwarder_input_info{application="0",audit="0",infrastructure="0",resource_name="instance",resource_namespace="openshift-logging"} 1
# HELP log_forwarder_output_info Info metric containing information about usage of available output types. Value is always 1.
# TYPE log_forwarder_output_info gauge
log_forwarder_output_info{cloudwatch="0",default="0",elasticsearch="0",fluentdForward="0",googleCloudLogging="0",kafka="0",loki="0",resource_name="instance",resource_namespace="openshift-logging",splunk="0",syslog="0"} 1
# HELP log_forwarder_pipeline_info Info metric containing information about deployed forwarders. Value is always 1.
# TYPE log_forwarder_pipeline_info gauge
log_forwarder_pipeline_info{healthStatus="0",pipelineInfo="0",resource_name="instance",resource_namespace="openshift-logging"} 1
# HELP log_logging_info Info metric containing general information about installed operator. Value is always 1.
# TYPE log_logging_info gauge
log_logging_info{healthStatus="0",managedStatus="0",resource_name="instance",resource_namespace="openshift-logging",version="test-version"} 1
`

				ctx := context.Background()
				k8s := newFakeClient(cl)
				collector := newTelemetryCollector(ctx, k8s, testVersion)

				metricsReader := strings.NewReader(wantMetrics)
				err := testutil.CollectAndCompare(collector, metricsReader)
				Expect(err).To(BeNil())
			})
		})

		Context("with all resources present", func() {
			It("should provide all metrics", func() {
				cl := testReadyClusterLogging(loggingv1.LogStoreTypeLokiStack)
				wantMetrics := `# HELP log_collector_error_count_total Counts the number of errors encountered by the operator reconciling the collector configuration
# TYPE log_collector_error_count_total counter
log_collector_error_count_total{version="test-version"} 0
# HELP log_forwarder_input_info Info metric containing information about usage of pre-defined input names. Value is always 1.
# TYPE log_forwarder_input_info gauge
log_forwarder_input_info{application="1",audit="0",infrastructure="1",resource_name="instance",resource_namespace="openshift-logging"} 1
# HELP log_forwarder_output_info Info metric containing information about usage of available output types. Value is always 1.
# TYPE log_forwarder_output_info gauge
log_forwarder_output_info{cloudwatch="0",default="1",elasticsearch="0",fluentdForward="0",googleCloudLogging="0",kafka="0",loki="1",resource_name="instance",resource_namespace="openshift-logging",splunk="0",syslog="0"} 1
# HELP log_forwarder_pipeline_info Info metric containing information about deployed forwarders. Value is always 1.
# TYPE log_forwarder_pipeline_info gauge
log_forwarder_pipeline_info{healthStatus="1",pipelineInfo="2",resource_name="instance",resource_namespace="openshift-logging",} 1
# HELP log_logging_info Info metric containing general information about installed operator. Value is always 1.
# TYPE log_logging_info gauge
log_logging_info{healthStatus="1",managedStatus="1",resource_name="instance",resource_namespace="openshift-logging",version="test-version"} 1
`

				ctx := context.Background()
				k8s := newFakeClient(cl, testDefaultLokiForwarder)
				collector := newTelemetryCollector(ctx, k8s, testVersion)
				collector.updateDefaultInfo(testDefaultLokiForwarder)

				metricsReader := strings.NewReader(wantMetrics)
				err := testutil.CollectAndCompare(collector, metricsReader)
				Expect(err).To(BeNil())
			})
		})
	})
})
