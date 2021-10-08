package vector

import (
	"fmt"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/constants"

	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

func Sources(spec *logging.ClusterLogForwarderSpec, op generator.Options) []source.LogSource {
	return LogSources(spec, op)
}

func MetricSources(spec *logging.ClusterLogForwarderSpec, o generator.Options) []generator.Element {
	return []generator.Element{
		PrometheusMonitor{
			Desc:         "Prometheus Monitoring",
			TemplateName: "PrometheusMonitor",
			TemplateStr:  PrometheusMonitorTemplate,
		},
	}
}

func LogSources(spec *logging.ClusterLogForwarderSpec, op generator.Options) []source.LogSource {
	var el []source.LogSource = make([]source.LogSource, 0)
	types := generator.GatherSources(spec, op)
	if types.Has(logging.InputNameApplication) {
		el = append(el,
			source.KubernetesLogs{
				SourceID:     logging.InputNameApplication,
				SourceType:   "kubernetes_logs",
				Desc:         "Logs from containers",
				ExcludePaths: ExcludeContainerPaths(),
			})
	}
	if !types.Has(logging.InputNameApplication) && types.Has(logging.InputNameInfrastructure) {
		el = append(el,
			source.KubernetesLogs{
				SourceID:     logging.InputNameInfrastructure,
				SourceType:   "kubernetes_logs",
				Desc:         "Logs from containers",
				ExcludePaths: ExcludeContainerPaths(),
			})
	}
	if types.Has(logging.InputNameInfrastructure) {
		el = append(el,
			source.JournalLogs{
				SourceID:   logging.InputNameInfrastructure,
				SourceType: "journald",
				Desc:       "Logs from journald",
			})
	}
	return el
}

func VectorCollectorLogsPath() string {
	return fmt.Sprintf("/var/log/pods/%%s-*_%s_*.log", constants.OpenshiftNS)
}

func VectorLogStoreLogsPath() string {
	return fmt.Sprintf("/var/log/pods/%%s-*_%s_*.log", constants.OpenshiftNS)
}

func VectorVisualizationLogsPath() string {
	return fmt.Sprintf("/var/log/pods/%s-*_%s_*.log", constants.KibanaName, constants.OpenshiftNS)
}

func ExcludeContainerPaths() string {
	return fmt.Sprintf("[%s]", strings.Join(
		[]string{
			fmt.Sprintf("%q", fmt.Sprintf(VectorCollectorLogsPath(), constants.CollectorName)),
			fmt.Sprintf("%q", fmt.Sprintf(VectorLogStoreLogsPath(), constants.ElasticsearchName)),
			fmt.Sprintf("%q", VectorVisualizationLogsPath()),
		},
		", ",
	))
}
