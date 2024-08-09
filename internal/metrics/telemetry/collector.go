package telemetry

import (
	"context"

	log "github.com/ViaQ/logerr/v2/log/static"
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
	observabilityv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/prometheus/client_golang/prometheus"
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
	descs <- logFileMetricExporterInfoDesc

	descs <- forwarderPipelinesDesc
	descs <- forwarderInputTypeDesc
	descs <- forwarderOutputTypeDesc
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
		pipelines := float64(len(clf.Spec.Pipelines))
		m <- prometheus.MustNewConstMetric(forwarderPipelinesDesc, prometheus.GaugeValue, pipelines,
			t.version, clf.Namespace, clf.Name)

		inputTypes := map[observabilityv1.InputType]int{}
		for _, i := range clf.Spec.Inputs {
			count := inputTypes[i.Type]
			inputTypes[i.Type] = count + 1
		}

		for _, p := range clf.Spec.Pipelines {
			for _, ir := range p.InputRefs {
				input := observabilityv1.InputType(ir)
				switch input {
				case observabilityv1.InputTypeApplication, observabilityv1.InputTypeInfrastructure, observabilityv1.InputTypeAudit:
					// Treat predefined input references as "input types"
					count := inputTypes[input]
					inputTypes[input] = count + 1
				default:
				}
			}
		}

		for input, c := range inputTypes {
			m <- prometheus.MustNewConstMetric(forwarderInputTypeDesc, prometheus.GaugeValue, float64(c),
				t.version, clf.Namespace, clf.Name, string(input))
		}

		outputTypes := map[observabilityv1.OutputType]int{}
		for _, o := range clf.Spec.Outputs {
			count := outputTypes[o.Type]
			outputTypes[o.Type] = count + 1
		}

		for output, c := range outputTypes {
			m <- prometheus.MustNewConstMetric(forwarderOutputTypeDesc, prometheus.GaugeValue, float64(c),
				t.version, clf.Namespace, clf.Name, string(output))
		}
	}

	return nil
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

	m <- prometheus.MustNewConstMetric(logFileMetricExporterInfoDesc, prometheus.GaugeValue, 1.0,
		t.version, boolLabel(deployed), boolLabel(healthy))
	return nil
}

func boolLabel(value bool) string {
	if value {
		return boolYes
	}

	return boolNo
}

func hasReadyCondition(conditions []metav1.Condition) bool {
	for _, c := range conditions {
		if c.Type == observabilityv1.ConditionTypeReady && c.Status == metav1.ConditionTrue {
			return true
		}
	}

	return false
}
