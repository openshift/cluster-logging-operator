package fluentd

import (
	"fmt"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/pkg/generator"
	"github.com/openshift/cluster-logging-operator/pkg/generator/fluentd/source"
)

func Sources(spec *logging.ClusterLogForwarderSpec, o Options) []Element {
	return MergeElements(
		MetricSources(spec, o),
		LogSources(spec, o),
	)
}

func MetricSources(spec *logging.ClusterLogForwarderSpec, o Options) []Element {
	return []Element{
		PrometheusMonitor{
			Desc:         "Prometheus Monitoring",
			TemplateName: "PrometheusMonitor",
			TemplateStr:  PrometheusMonitorTemplate,
		},
	}
}

func LogSources(spec *logging.ClusterLogForwarderSpec, o Options) []Element {
	var el []Element = make([]Element, 0)
	types := GatherSources(spec, o)
	types = AddLegacySources(types, o)
	if types.Has(logging.InputNameInfrastructure) {
		el = append(el,
			source.JournalLog{
				Desc:         "Logs from linux journal",
				OutLabel:     "MEASURE",
				TemplateName: "inputSourceJournalTemplate",
				TemplateStr:  source.JournalLogTemplate,
			})
	}
	if types.Has(logging.InputNameApplication) || types.Has(logging.InputNameInfrastructure) {
		el = append(el,
			source.ContainerLogs{
				Desc:         "Logs from containers (including openshift containers)",
				Paths:        ContainerLogPaths(),
				ExcludePaths: ExcludeContainerPaths(),
				PosFile:      "/var/lib/fluentd/pos/es-containers.log.pos",
				OutLabel:     "MEASURE",
			})
	}
	if types.Has(logging.InputNameAudit) {
		el = append(el,
			source.HostAuditLog{
				Desc:         "linux audit logs",
				OutLabel:     "MEASURE",
				TemplateName: "inputSourceHostAuditTemplate",
				TemplateStr:  source.HostAuditLogTemplate,
			},
			source.K8sAuditLog{
				Desc:         "k8s audit logs",
				OutLabel:     "MEASURE",
				TemplateName: "inputSourceK8sAuditTemplate",
				TemplateStr:  source.K8sAuditLogTemplate,
			},
			source.OpenshiftAuditLog{
				Desc:         "Openshift audit logs",
				OutLabel:     "MEASURE",
				TemplateName: "inputSourceOpenShiftAuditTemplate",
				TemplateStr:  source.OpenshiftAuditLogTemplate,
			},
			source.OVNAuditLogs{
				Desc:         "Openshift Virtual Network (OVN) audit logs",
				OutLabel:     "MEASURE",
				TemplateName: "inputSourceOVNAuditTemplate",
				TemplateStr:  source.OVNAuditLogTemplate,
			},
		)
	}
	return el
}

func ContainerLogPaths() string {
	return fmt.Sprintf("%q", "/var/log/containers/*.log")
}

func ExcludeContainerPaths() string {
	return fmt.Sprintf("[%s]", strings.Join(
		[]string{
			fmt.Sprintf("%q", fmt.Sprintf(CollectorLogsPath(), FluentdCollectorPodNamePrefix)),
			fmt.Sprintf("%q", fmt.Sprintf(LogStoreLogsPath(), ESLogStorePodNamePrefix)),
			fmt.Sprintf("%q", VisualizationLogsPath()),
		},
		", ",
	))
}
