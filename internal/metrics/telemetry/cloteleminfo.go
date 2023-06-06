package telemetry

import (
	v1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/version"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	IsPresent    = "1"
	IsNotPresent = "0"

	InputNameApplication    = v1.InputNameApplication
	InputNameAudit          = v1.InputNameAudit
	InputNameInfrastructure = v1.InputNameInfrastructure

	OutputTypeDefault            = "default"
	OutputTypeElasticsearch      = v1.OutputTypeElasticsearch
	OutputTypeFluentdForward     = v1.OutputTypeFluentdForward
	OutputTypeSyslog             = v1.OutputTypeSyslog
	OutputTypeKafka              = v1.OutputTypeKafka
	OutputTypeLoki               = v1.OutputTypeLoki
	OutputTypeCloudwatch         = v1.OutputTypeCloudwatch
	OutputTypeHttp               = v1.OutputTypeHttp
	OutputTypeGoogleCloudLogging = v1.OutputTypeGoogleCloudLogging
	OutputTypeSplunk             = v1.OutputTypeSplunk

	ManagedStatus = "managedStatus"
	HealthStatus  = "healthStatus"
	VersionNo     = "version"
	PipelineNo    = "pipelineInfo"
	Deployed      = "deployed"
)

// placeholder for keeping clo info which will be used for clo metrics update
type TData struct {
	CLInfo              utils.StringMap
	CLLogStoreType      utils.StringMap
	CollectorErrorCount utils.Float64Map
	CLFInfo             utils.StringMap
	CLFInputType        utils.StringMap
	CLFOutputType       utils.StringMap
	LFMEInfo            utils.StringMap
}

// IsNotPresent stands for managedStatus and healthStatus true and healthy
func NewTD() *TData {
	return &TData{
		CLInfo:              utils.StringMap{M: map[string]string{VersionNo: version.Version, ManagedStatus: IsNotPresent, HealthStatus: IsNotPresent}},
		CLLogStoreType:      utils.StringMap{M: map[string]string{OutputTypeElasticsearch: IsNotPresent, OutputTypeLoki: IsNotPresent}},
		CollectorErrorCount: utils.Float64Map{M: map[string]float64{"CollectorErrorCount": 0}},
		CLFInfo:             utils.StringMap{M: map[string]string{HealthStatus: IsNotPresent, PipelineNo: IsNotPresent}},
		CLFInputType:        utils.StringMap{M: map[string]string{InputNameApplication: IsNotPresent, InputNameAudit: IsNotPresent, InputNameInfrastructure: IsNotPresent}},
		CLFOutputType: utils.StringMap{M: map[string]string{
			OutputTypeDefault:            IsNotPresent,
			OutputTypeElasticsearch:      IsNotPresent,
			OutputTypeFluentdForward:     IsNotPresent,
			OutputTypeSyslog:             IsNotPresent,
			OutputTypeKafka:              IsNotPresent,
			OutputTypeLoki:               IsNotPresent,
			OutputTypeCloudwatch:         IsNotPresent,
			OutputTypeHttp:               IsNotPresent,
			OutputTypeSplunk:             IsNotPresent,
			OutputTypeGoogleCloudLogging: IsNotPresent}},
		LFMEInfo: utils.StringMap{M: map[string]string{Deployed: IsNotPresent, HealthStatus: IsNotPresent}},
	}
}

var (
	Data = NewTD()

	mCLInfo = NewInfoVec(
		"log_logging_info",
		"Clo version managementState healthState specific metric",
		[]string{VersionNo, ManagedStatus, HealthStatus},
	)
	mCollectorErrorCount = NewInfoVec(
		"log_collector_error_count_total",
		"log collector total number of error counts ",
		[]string{VersionNo},
	)
	mCLFInfo = NewInfoVec(
		"log_forwarder_pipeline_info",
		"Clf healthState and pipelineInfo specific metric",
		[]string{HealthStatus, PipelineNo},
	)

	mCLFInputType = NewInfoVec(
		"log_forwarder_input_info",
		"Clf input type specific metric",
		[]string{InputNameApplication, InputNameAudit, InputNameInfrastructure},
	)

	mCLFOutputType = NewInfoVec(
		"log_forwarder_output_info",
		"Clf output type specific metric",
		[]string{
			OutputTypeDefault,
			OutputTypeElasticsearch,
			OutputTypeFluentdForward,
			OutputTypeSyslog,
			OutputTypeKafka,
			OutputTypeLoki,
			OutputTypeCloudwatch,
			OutputTypeHttp,
			OutputTypeSplunk,
			OutputTypeGoogleCloudLogging},
	)

	mLFMEInfo = NewInfoVec(
		"log_file_metric_exporter_info",
		"LFME health and deployed status specific metric",
		[]string{Deployed, HealthStatus},
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

	MetricLFMEList = []prometheus.Collector{
		mLFMEInfo,
	}
)

func RegisterMetrics() error {

	for _, metric := range MetricCLList {
		metrics.Registry.MustRegister(metric)
	}

	for _, metric := range MetricCLFList {
		metrics.Registry.MustRegister(metric)
	}

	for _, metric := range MetricLFMEList {
		metrics.Registry.MustRegister(metric)
	}

	return nil
}

func SetCLMetrics(value float64) {
	CLInfo := Data.CLInfo.M
	mCLInfo.With(prometheus.Labels{
		VersionNo:     CLInfo[VersionNo],
		ManagedStatus: CLInfo[ManagedStatus],
		HealthStatus:  CLInfo[HealthStatus]}).Set(value)
}

func SetCLFMetrics(value float64) {
	CLInfo := Data.CLInfo.M
	CErrorCount := Data.CollectorErrorCount.M
	CLFInfo := Data.CLFInfo.M
	CLFInputType := Data.CLFInputType.M
	CLFOutputType := Data.CLFOutputType.M

	mCollectorErrorCount.With(prometheus.Labels{
		VersionNo: CLInfo[VersionNo]}).Set(CErrorCount["CollectorErrorCount"])

	mCLFInfo.With(prometheus.Labels{
		HealthStatus: CLFInfo[HealthStatus],
		PipelineNo:   CLFInfo[PipelineNo]}).Set(value)

	mCLFInputType.With(prometheus.Labels{
		InputNameApplication:    CLFInputType[InputNameApplication],
		InputNameAudit:          CLFInputType[InputNameAudit],
		InputNameInfrastructure: CLFInputType[InputNameInfrastructure]}).Set(value)

	mCLFOutputType.With(prometheus.Labels{
		OutputTypeDefault:            CLFOutputType[OutputTypeDefault],
		OutputTypeElasticsearch:      CLFOutputType[OutputTypeElasticsearch],
		OutputTypeFluentdForward:     CLFOutputType[OutputTypeFluentdForward],
		OutputTypeSyslog:             CLFOutputType[OutputTypeSyslog],
		OutputTypeKafka:              CLFOutputType[OutputTypeKafka],
		OutputTypeLoki:               CLFOutputType[OutputTypeLoki],
		OutputTypeCloudwatch:         CLFOutputType[OutputTypeCloudwatch],
		OutputTypeHttp:               CLFOutputType[OutputTypeHttp],
		OutputTypeSplunk:             CLFOutputType[OutputTypeSplunk],
		OutputTypeGoogleCloudLogging: CLFOutputType[OutputTypeGoogleCloudLogging]}).Set(value)
}

func SetLFMEMetrics(value float64) {
	LFMEInfo := Data.LFMEInfo.M

	mLFMEInfo.With(prometheus.Labels{
		Deployed:     LFMEInfo[Deployed],
		HealthStatus: LFMEInfo[HealthStatus],
	}).Set(value)
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
