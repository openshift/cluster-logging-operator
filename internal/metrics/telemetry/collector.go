package telemetry

import (
	"context"
	"strconv"

	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	observabilityv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/status"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	boolYes = "1"
	boolNo  = "0"
)

type telemetryCollector struct {
	ctx     context.Context
	client  client.Client
	version string
}

func newTelemetryCollector(ctx context.Context, k8sClient client.Client, version string) *telemetryCollector {
	return &telemetryCollector{
		ctx:     ctx,
		client:  k8sClient,
		version: version,
	}
}

var _ prometheus.Collector = &telemetryCollector{}

func (t *telemetryCollector) Describe(descs chan<- *prometheus.Desc) {
	descs <- clusterLogForwarderDesc
	descs <- forwarderInputInfoDesc
	descs <- forwarderOutputInfoDesc
	descs <- logFileMetricExporterInfoDesc
}

func (t *telemetryCollector) Collect(m chan<- prometheus.Metric) {
	if err := t.collectForwarder(m); err != nil {
		log.V(1).Error(err, "Error collecting telemetry for cluster log forwarders")
	}

	if err := t.collectLogFileMetricExporter(m); err != nil {
		log.V(1).Error(err, "Error collecting telemetry for LogFileMetricExporter")
	}
}

func (t *telemetryCollector) collectForwarder(m chan<- prometheus.Metric) error {
	clfList := &observabilityv1.ClusterLogForwarderList{}
	if err := t.client.List(t.ctx, clfList); err != nil {
		return err
	}

	for _, clf := range clfList.Items {
		healthy := hasReadyCondition(clf.Status.Conditions)
		pipelines, inputs, outputs := gatherForwarderInfo(&clf)

		t.collectForwarderMetrics(m, clf.Namespace, clf.Name, healthy, pipelines, inputs, outputs)
	}

	return nil
}

func (t *telemetryCollector) collectForwarderMetrics(m chan<- prometheus.Metric, namespace, name string, healthy bool, pipelines uint, inputs, outputs []string) {
	m <- prometheus.MustNewConstMetric(clusterLogForwarderDesc, prometheus.GaugeValue, 1.0,
		namespace, name, boolLabel(healthy), uintLabel(pipelines))

	inputLabels := append([]string{namespace, name}, inputs...)
	m <- prometheus.MustNewConstMetric(forwarderInputInfoDesc, prometheus.GaugeValue, 1.0, inputLabels...)

	outputLabels := append([]string{namespace, name}, outputs...)
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
		healthy = hasLegacyReadyCondition(lfme.Status.Conditions)
	}

	m <- prometheus.MustNewConstMetric(logFileMetricExporterInfoDesc, prometheus.GaugeValue, 1.0, boolLabel(deployed), boolLabel(healthy))
	return nil
}

func boolLabel(value bool) string {
	if value {
		return boolYes
	}

	return boolNo
}

func uintLabel(value uint) string {
	return strconv.FormatUint(uint64(value), 10)
}

func makeZeroStrings(length int) []string {
	result := make([]string, length)
	for i := range result {
		result[i] = boolNo
	}

	return result
}

func hasLegacyReadyCondition(conditions status.Conditions) bool {
	for _, c := range conditions {
		if c.Type == loggingv1.ConditionReady && c.Status == corev1.ConditionTrue {
			return true
		}
	}

	return false
}

func hasReadyCondition(conditions []metav1.Condition) bool {
	for _, c := range conditions {
		if c.Type == observabilityv1.ConditionTypeReady && c.Status == metav1.ConditionTrue {
			return true
		}
	}

	return false
}
