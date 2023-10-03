package source

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	common "github.com/openshift/cluster-logging-operator/internal/generator/vector"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
)

func Sources(spec logging.ClusterLogForwarderSpec, namespace string, op generator.Options) []generator.Element {
	return generator.MergeElements(
		LogSources(spec, namespace, op),
		common.HttpSources(&spec, op),
		common.MetricsSources(common.InternalMetricsSourceName),
	)
}

func LogSources(spec logging.ClusterLogForwarderSpec, namespace string, op generator.Options) []generator.Element {
	el := []generator.Element{}

	types := generator.GatherSources(&spec, op)
	if types.Has(logging.InputNameContainer) || types.Has(logging.InputNameApplication) || types.Has(logging.InputNameInfrastructure) {
		el = append(el,
			KubernetesLogs{
				ComponentID:  "raw_container_logs",
				Desc:         "Logs from containers (including openshift containers)",
				ExcludePaths: common.ExcludeContainerPaths(namespace),
			})
	}
	if types.Has(logging.InputNameNode) || types.Has(logging.InputNameInfrastructure) {
		el = append(el,
			source.JournalLog{
				ComponentID:  "raw_node_logs",
				Desc:         "Logs from linux journal",
				TemplateName: "inputSourceJournalTemplate",
				TemplateStr:  source.JournalLogTemplate,
			})
	}
	if types.Has(logging.InputNameAudit) {
		el = append(el,
			source.HostAuditLog{
				ComponentID:  common.RawHostAuditLogs,
				Desc:         "Logs from host audit",
				TemplateName: "inputSourceHostAuditTemplate",
				TemplateStr:  source.HostAuditLogTemplate,
			},
			source.K8sAuditLog{
				ComponentID:  common.RawK8sAuditLogs,
				Desc:         "Logs from kubernetes audit",
				TemplateName: "inputSourceK8sAuditTemplate",
				TemplateStr:  source.K8sAuditLogTemplate,
			},
			source.OpenshiftAuditLog{
				ComponentID:  common.RawOpenshiftAuditLogs,
				Desc:         "Logs from openshift audit",
				TemplateName: "inputSourceOpenShiftAuditTemplate",
				TemplateStr:  source.OpenshiftAuditLogTemplate,
			},
			source.OVNAuditLog{
				ComponentID:  common.RawOvnAuditLogs,
				Desc:         "Logs from ovn audit",
				TemplateName: "inputSourceOVNAuditTemplate",
				TemplateStr:  source.OVNAuditLogTemplate,
			})
	}
	return el
}
