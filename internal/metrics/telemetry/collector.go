package telemetry

import (
	"context"
	"strconv"

	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/status"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type clusterLogForwarderData struct {
	NumPipelines uint
	// Inputs contains the label values for the different inputs listed in "forwarderInputTypes".
	// The value can either be "0" or "1" depending on whether the listed input is present in the configuration.
	// The order of the values needs to match the keys present in "forwarderInputTypes".
	Inputs []string
	// Outputs contains the label values for the different inputs listed in "forwarderOutputTypes".
	// The value can either be "0" or "1" depending on whether the listed input is present in the configuration.
	// The order of the values needs to match the keys present in "forwarderOutputTypes".
	Outputs []string
}

type telemetryCollector struct {
	ctx     context.Context
	client  client.Client
	version string

	collectorErrors prometheus.Counter
	defaultCLFInfo  clusterLogForwarderData
}

func newTelemetryCollector(ctx context.Context, k8sClient client.Client, version string) *telemetryCollector {
	return &telemetryCollector{
		ctx:     ctx,
		client:  k8sClient,
		version: version,

		collectorErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: metricsPrefix + "collector_error_count_total",
			Help: "Counts the number of errors encountered by the operator reconciling the collector configuration",
			ConstLabels: map[string]string{
				labelVersion: version,
			},
		}),
		defaultCLFInfo: clusterLogForwarderData{
			NumPipelines: 0,
			Inputs:       makeZeroStrings(len(forwarderInputTypes)),
			Outputs:      makeZeroStrings(len(forwarderOutputTypes)),
		},
	}
}

var _ prometheus.Collector = &telemetryCollector{}

func (t *telemetryCollector) Describe(descs chan<- *prometheus.Desc) {
	descs <- t.collectorErrors.Desc()
	descs <- clusterLoggingInfoDesc
	descs <- clusterLogForwarderDesc
	descs <- forwarderInputInfoDesc
	descs <- forwarderOutputInfoDesc
	descs <- logFileMetricExporterInfoDesc
}

func (t *telemetryCollector) Collect(m chan<- prometheus.Metric) {
	m <- t.collectorErrors

	if err := t.collectForwarder(m); err != nil {
		log.V(1).Error(err, "Error collecting telemetry for cluster logging and forwarders")
	}

	if err := t.collectLogFileMetricExporter(m); err != nil {
		log.V(1).Error(err, "Error collecting telemetry for LogFileMetricExporter")
	}
}

func (t *telemetryCollector) collectForwarder(m chan<- prometheus.Metric) error {
	cloList := &loggingv1.ClusterLoggingList{}
	if err := t.client.List(t.ctx, cloList); err != nil {
		return err
	}

	clfList := &loggingv1.ClusterLogForwarderList{}
	if err := t.client.List(t.ctx, clfList); err != nil {
		return err
	}

	if len(cloList.Items) == 0 && len(clfList.Items) == 0 {
		// No resources present, no telemetry needed
		return nil
	}

	for _, cl := range cloList.Items {
		if cl.Namespace == constants.OpenshiftNS && cl.Name == constants.SingletonName {
			if err := t.collectDefaultInstance(m, &cl); err != nil {
				return err
			}
		}

		managedLabel := boolLabel(cl.Spec.ManagementState == loggingv1.ManagementStateManaged)
		healthyLabel := boolLabel(hasReadyCondition(cl.Status.Conditions))
		m <- prometheus.MustNewConstMetric(clusterLoggingInfoDesc, prometheus.GaugeValue, 1.0,
			cl.Namespace, cl.Name, t.version, managedLabel, healthyLabel)
	}

	for _, clf := range clfList.Items {
		if clf.Namespace == constants.OpenshiftNS && clf.Name == constants.SingletonName {
			// Default instance has already been collected above
			continue
		}

		healthy := hasReadyCondition(clf.Status.Conditions)
		pipelines, inputs, outputs := gatherForwarderInfo(&clf)

		t.collectForwarderMetrics(m, clf.Namespace, clf.Name, healthy, false, pipelines, inputs, outputs)
	}

	return nil
}

func (t *telemetryCollector) collectDefaultInstance(m chan<- prometheus.Metric, cl *loggingv1.ClusterLogging) error {
	healthy := hasReadyCondition(cl.Status.Conditions)
	defaultStorage := cl.Spec.LogStore != nil &&
		(cl.Spec.LogStore.Type == loggingv1.LogStoreTypeElasticsearch ||
			cl.Spec.LogStore.Type == loggingv1.LogStoreTypeLokiStack)

	t.collectForwarderMetrics(m, constants.OpenshiftNS, constants.SingletonName,
		healthy, defaultStorage,
		t.defaultCLFInfo.NumPipelines, t.defaultCLFInfo.Inputs, t.defaultCLFInfo.Outputs,
	)
	return nil
}

func (t *telemetryCollector) collectForwarderMetrics(m chan<- prometheus.Metric, namespace, name string, healthy, defaultStorage bool, pipelines uint, inputs, outputs []string) {
	m <- prometheus.MustNewConstMetric(clusterLogForwarderDesc, prometheus.GaugeValue, 1.0,
		namespace, name, boolLabel(healthy), uintLabel(pipelines))

	inputLabels := append([]string{namespace, name}, inputs...)
	m <- prometheus.MustNewConstMetric(forwarderInputInfoDesc, prometheus.GaugeValue, 1.0, inputLabels...)

	outputLabels := append([]string{namespace, name, boolLabel(defaultStorage)}, outputs...)
	m <- prometheus.MustNewConstMetric(forwarderOutputInfoDesc, prometheus.GaugeValue, 1.0, outputLabels...)
}

func (t *telemetryCollector) collectLogFileMetricExporter(m chan<- prometheus.Metric) error {
	lfmeList := &loggingv1alpha1.LogFileMetricExporterList{}
	if err := t.client.List(t.ctx, lfmeList); err != nil {
		return err
	}

	deployed := false
	healthy := false

	for _, lfme := range lfmeList.Items {
		if lfme.Namespace != constants.OpenshiftNS || lfme.Name != constants.SingletonName {
			// Only singleton instance is valid
			continue
		}

		deployed = true
		healthy = hasReadyCondition(lfme.Status.Conditions)
	}

	m <- prometheus.MustNewConstMetric(logFileMetricExporterInfoDesc, prometheus.GaugeValue, 1.0, boolLabel(deployed), boolLabel(healthy))
	return nil
}

func boolLabel(value bool) string {
	if value {
		return "1"
	}

	return "0"
}

func uintLabel(value uint) string {
	return strconv.FormatUint(uint64(value), 10)
}

func makeZeroStrings(length int) []string {
	result := make([]string, length)
	for i := range result {
		result[i] = "0"
	}

	return result
}

func hasReadyCondition(conditions status.Conditions) bool {
	for _, c := range conditions {
		if c.Type == loggingv1.ConditionReady && c.Status == corev1.ConditionTrue {
			return true
		}
	}

	return false
}
