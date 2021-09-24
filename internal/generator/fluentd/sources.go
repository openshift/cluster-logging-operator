package fluentd

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	source2 "github.com/openshift/cluster-logging-operator/internal/generator/fluentd/source"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

func Sources(spec *logging.ClusterLogForwarderSpec, o generator.Options) []generator.Element {
	return generator.MergeElements(
		MetricSources(spec, o),
		LogSources(spec, o),
	)
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

func LogSources(spec *logging.ClusterLogForwarderSpec, o generator.Options) []generator.Element {
	var el []generator.Element = make([]generator.Element, 0)
	types := generator.GatherSources(spec, o)
	if types.Has(logging.InputNameInfrastructure) {
		el = append(el,
			source2.JournalLog{
				Desc:         "Logs from linux journal",
				OutLabel:     "MEASURE",
				TemplateName: "inputSourceJournalTemplate",
				TemplateStr:  source2.JournalLogTemplate,
			})
	}
	if types.Has(logging.InputNameApplication) || types.Has(logging.InputNameInfrastructure) {
		el = append(el,
			source2.ContainerLogs{
				Desc:         "Logs from containers (including openshift containers)",
				Paths:        ContainerLogPaths(),
				ExcludePaths: ExcludeContainerPaths(),
				PosFile:      "/var/lib/fluentd/pos/es-containers.log.pos",
				OutLabel:     "MEASURE",
			})
	}
	if types.Has(logging.InputNameAudit) {
		el = append(el,
			source2.HostAuditLog{
				Desc:         "linux audit logs",
				OutLabel:     "MEASURE",
				TemplateName: "inputSourceHostAuditTemplate",
				TemplateStr:  source2.HostAuditLogTemplate,
			},
			source2.K8sAuditLog{
				Desc:         "k8s audit logs",
				OutLabel:     "MEASURE",
				TemplateName: "inputSourceK8sAuditTemplate",
				TemplateStr:  source2.K8sAuditLogTemplate,
			},
			source2.OpenshiftAuditLog{
				Desc:         "Openshift audit logs",
				OutLabel:     "MEASURE",
				TemplateName: "inputSourceOpenShiftAuditTemplate",
				TemplateStr:  source2.OpenshiftAuditLogTemplate,
			},
			source2.OVNAuditLogs{
				Desc:         "Openshift Virtual Network (OVN) audit logs",
				OutLabel:     "MEASURE",
				TemplateName: "inputSourceOVNAuditTemplate",
				TemplateStr:  source2.OVNAuditLogTemplate,
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
			fmt.Sprintf("%q", fmt.Sprintf(generator.CollectorLogsPath(), generator.FluentdCollectorPodNamePrefix)),
			fmt.Sprintf("%q", fmt.Sprintf(generator.LogStoreLogsPath(), generator.ESLogStorePodNamePrefix)),
			fmt.Sprintf("%q", generator.VisualizationLogsPath()),
		},
		", ",
	))
}
