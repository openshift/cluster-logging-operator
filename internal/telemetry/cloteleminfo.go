package telemetry

import (
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/version"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// placeholder for keeping clo info which will be used for clo metrics update
type TData struct {
	CLInfo              utils.StringMap
	CLLogStoreType      utils.StringMap
	CollectorErrorCount utils.Float64Map
	CLFInfo             utils.StringMap
	CLFInputType        utils.StringMap
	CLFOutputType       utils.StringMap
}

// "0" stands for managedStatus and healthStatus true and healthy
func NewTD() *TData {
	return &TData{
		CLInfo:              utils.StringMap{M: map[string]string{"version": version.Version, "managedStatus": "0", "healthStatus": "0"}},
		CLLogStoreType:      utils.StringMap{M: map[string]string{"elasticsearch": "0", "loki": "0"}},
		CollectorErrorCount: utils.Float64Map{M: map[string]float64{"CollectorErrorCount": 0}},
		CLFInfo:             utils.StringMap{M: map[string]string{"healthStatus": "0", "pipelineInfo": "0"}},
		CLFInputType:        utils.StringMap{M: map[string]string{"application": "0", "audit": "0", "infrastructure": "0"}},
		CLFOutputType:       utils.StringMap{M: map[string]string{"elasticsearch": "0", "fluentdForward": "0", "syslog": "0", "kafka": "0", "loki": "0", "cloudwatch": "0"}},
	}
}

var (
	Data = NewTD()

	mCLInfo = NewInfoVec(
		"log_logging_info",
		"Clo version managementState healthState specific metric",
		[]string{"version", "managedStatus", "healthStatus"},
	)
	mCollectorErrorCount = NewInfoVec(
		"log_collector_error_count_total",
		"log collector total number of error counts ",
		[]string{"version"},
	)
	mCLFInfo = NewInfoVec(
		"log_forwarder_pipeline_info",
		"Clf healthState and pipelineInfo specific metric",
		[]string{"healthStatus", "pipelineInfo"},
	)

	mCLFInputType = NewInfoVec(
		"log_forwarder_input_info",
		"Clf input type specific metric",
		[]string{"application", "audit", "infrastructure"},
	)

	mCLFOutputType = NewInfoVec(
		"log_forwarder_output_info",
		"Clf output type specific metric",
		[]string{"default", "elasticsearch", "fluentdForward", "syslog", "kafka", "loki", "cloudwatch"},
	)

	MetricCLList = []prometheus.Collector{
		mCLInfo,
	}

	MetricCLFList = []prometheus.Collector{
		mCollectorErrorCount,
		mCLFInfo,
		mCLFInputType,
		mCLFOutputType,
	}
)

func RegisterMetrics() error {

	for _, metric := range MetricCLList {
		metrics.Registry.MustRegister(metric)
	}

	for _, metric := range MetricCLFList {
		metrics.Registry.MustRegister(metric)
	}

	return nil
}

func UpdateCLMetrics() error {

	CLInfo := Data.CLInfo.M

	mCLInfo.With(prometheus.Labels{
		"version":       CLInfo["version"],
		"managedStatus": CLInfo["managedStatus"],
		"healthStatus":  CLInfo["healthStatus"]}).Set(1)

	return nil
}

func UpdateCLFMetrics() error {

	CLInfo := Data.CLInfo.M
	CErrorCount := Data.CollectorErrorCount.M
	CLFInfo := Data.CLFInfo.M
	CLFInputType := Data.CLFInputType.M
	CLFOutputType := Data.CLFOutputType.M

	mCollectorErrorCount.With(prometheus.Labels{
		"version": CLInfo["version"]}).Set(CErrorCount["CollectorErrorCount"])

	mCLFInfo.With(prometheus.Labels{
		"healthStatus": CLFInfo["healthStatus"],
		"pipelineInfo": CLFInfo["pipelineInfo"]}).Set(1)

	mCLFInputType.With(prometheus.Labels{
		"application":    CLFInputType["application"],
		"audit":          CLFInputType["audit"],
		"infrastructure": CLFInputType["infrastructure"]}).Set(1)

	mCLFOutputType.With(prometheus.Labels{
		"default":        CLFOutputType["default"],
		"elasticsearch":  CLFOutputType["elasticsearch"],
		"fluentdForward": CLFOutputType["fluentdForward"],
		"syslog":         CLFOutputType["syslog"],
		"kafka":          CLFOutputType["kafka"],
		"loki":           CLFOutputType["loki"],
		"cloudwatch":     CLFOutputType["cloudwatch"]}).Set(1)

	return nil
}
func NewInfoVec(metricname string, metrichelp string, labelNames []string) *prometheus.GaugeVec {

	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: metricname,
			Help: metrichelp,
		},
		labelNames,
	)
}
