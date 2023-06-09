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
	CLInfo              *utils.StringMap
	CLLogStoreType      *utils.StringMap
	CollectorErrorCount *utils.Float64Map
	CLFInfo             *utils.StringMap
	CLFInputType        *utils.StringMap
	CLFOutputType       *utils.StringMap
	LFMEInfo            *utils.StringMap
}

// IsNotPresent stands for managedStatus and healthStatus true and healthy
func NewTD() *TData {
	return &TData{
		CLInfo:              utils.InitStringMap(map[string]string{VersionNo: version.Version, ManagedStatus: IsNotPresent, HealthStatus: IsNotPresent}),
		CLLogStoreType:      utils.InitStringMap(map[string]string{OutputTypeElasticsearch: IsNotPresent, OutputTypeLoki: IsNotPresent}),
		CollectorErrorCount: utils.InitFloat64Map(map[string]float64{"CollectorErrorCount": 0}),
		CLFInfo:             utils.InitStringMap(map[string]string{HealthStatus: IsNotPresent, PipelineNo: IsNotPresent}),
		CLFInputType:        utils.InitStringMap(map[string]string{InputNameApplication: IsNotPresent, InputNameAudit: IsNotPresent, InputNameInfrastructure: IsNotPresent}),
		CLFOutputType: utils.InitStringMap(map[string]string{
			OutputTypeDefault:            IsNotPresent,
			OutputTypeElasticsearch:      IsNotPresent,
			OutputTypeFluentdForward:     IsNotPresent,
			OutputTypeSyslog:             IsNotPresent,
			OutputTypeKafka:              IsNotPresent,
			OutputTypeLoki:               IsNotPresent,
			OutputTypeCloudwatch:         IsNotPresent,
			OutputTypeHttp:               IsNotPresent,
			OutputTypeSplunk:             IsNotPresent,
			OutputTypeGoogleCloudLogging: IsNotPresent}),
		LFMEInfo: utils.InitStringMap(map[string]string{Deployed: IsNotPresent, HealthStatus: IsNotPresent}),
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
	CLInfo := Data.CLInfo
	mCLInfo.With(prometheus.Labels{
		VersionNo:     CLInfo.Get(VersionNo),
		ManagedStatus: CLInfo.Get(ManagedStatus),
		HealthStatus:  CLInfo.Get(HealthStatus)}).Set(value)
}

func SetCLFMetrics(value float64) {
	CLInfo := Data.CLInfo
	CErrorCount := Data.CollectorErrorCount
	CLFInfo := Data.CLFInfo
	CLFInputType := Data.CLFInputType
	CLFOutputType := Data.CLFOutputType

	mCollectorErrorCount.With(prometheus.Labels{
		VersionNo: CLInfo.Get(VersionNo)}).Set(CErrorCount.Get("CollectorErrorCount"))

	mCLFInfo.With(prometheus.Labels{
		HealthStatus: CLFInfo.Get(HealthStatus),
		PipelineNo:   CLFInfo.Get(PipelineNo)}).Set(value)

	mCLFInputType.With(prometheus.Labels{
		InputNameApplication:    CLFInputType.Get(InputNameApplication),
		InputNameAudit:          CLFInputType.Get(InputNameAudit),
		InputNameInfrastructure: CLFInputType.Get(InputNameInfrastructure)}).Set(value)

	mCLFOutputType.With(prometheus.Labels{
		OutputTypeDefault:            CLFOutputType.Get(OutputTypeDefault),
		OutputTypeElasticsearch:      CLFOutputType.Get(OutputTypeElasticsearch),
		OutputTypeFluentdForward:     CLFOutputType.Get(OutputTypeFluentdForward),
		OutputTypeSyslog:             CLFOutputType.Get(OutputTypeSyslog),
		OutputTypeKafka:              CLFOutputType.Get(OutputTypeKafka),
		OutputTypeLoki:               CLFOutputType.Get(OutputTypeLoki),
		OutputTypeCloudwatch:         CLFOutputType.Get(OutputTypeCloudwatch),
		OutputTypeHttp:               CLFOutputType.Get(OutputTypeHttp),
		OutputTypeSplunk:             CLFOutputType.Get(OutputTypeSplunk),
		OutputTypeGoogleCloudLogging: CLFOutputType.Get(OutputTypeGoogleCloudLogging)}).Set(value)
}

func SetLFMEMetrics(value float64) {
	LFMEInfo := Data.LFMEInfo

	mLFMEInfo.With(prometheus.Labels{
		Deployed:     LFMEInfo.Get(Deployed),
		HealthStatus: LFMEInfo.Get(HealthStatus),
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
