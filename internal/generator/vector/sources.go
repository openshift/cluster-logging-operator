package vector

import (
	"fmt"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

const (
	RawHostAuditLogs = "raw_host_audit_logs"
	HostAuditLogs    = "host_audit_logs"

	RawK8sAuditLogs = "raw_k8s_audit_logs"
	K8sAuditLogs    = "k8s_audit_logs"

	RawOpenshiftAuditLogs = "raw_openshift_audit_logs"
	OpenshiftAuditLogs    = "openshift_audit_logs"

	RawOvnAuditLogs = "raw_ovn_audit_logs"
	OvnAuditLogs    = "ovn_audit_logs"
)

func Sources(spec *logging.ClusterLogForwarderSpec, namespace string, op generator.Options) []generator.Element {
	return generator.MergeElements(
		LogSources(spec, namespace, op),
		MetricsSources(InternalMetricsSourceName),
	)
}

func LogSources(spec *logging.ClusterLogForwarderSpec, namespace string, op generator.Options) []generator.Element {
	var el []generator.Element = make([]generator.Element, 0)
	types := generator.GatherSources(spec, op)
	if types.Has(logging.InputNameApplication) || types.Has(logging.InputNameInfrastructure) {
		el = append(el,
			source.KubernetesLogs{
				ComponentID:  "raw_container_logs",
				Desc:         "Logs from containers (including openshift containers)",
				ExcludePaths: ExcludeContainerPaths(namespace),
			})
	}
	if types.Has(logging.InputNameInfrastructure) {
		el = append(el,
			source.JournalLog{
				ComponentID:  "raw_journal_logs",
				Desc:         "Logs from linux journal",
				TemplateName: "inputSourceJournalTemplate",
				TemplateStr:  source.JournalLogTemplate,
			})
	}
	if types.Has(logging.InputNameAudit) {
		el = append(el,
			source.HostAuditLog{
				ComponentID:  RawHostAuditLogs,
				Desc:         "Logs from host audit",
				TemplateName: "inputSourceHostAuditTemplate",
				TemplateStr:  source.HostAuditLogTemplate,
			},
			source.K8sAuditLog{
				ComponentID:  RawK8sAuditLogs,
				Desc:         "Logs from kubernetes audit",
				TemplateName: "inputSourceK8sAuditTemplate",
				TemplateStr:  source.K8sAuditLogTemplate,
			},
			source.OpenshiftAuditLog{
				ComponentID:  RawOpenshiftAuditLogs,
				Desc:         "Logs from openshift audit",
				TemplateName: "inputSourceOpenShiftAuditTemplate",
				TemplateStr:  source.OpenshiftAuditLogTemplate,
			},
			source.OVNAuditLog{
				ComponentID:  RawOvnAuditLogs,
				Desc:         "Logs from ovn audit",
				TemplateName: "inputSourceOVNAuditTemplate",
				TemplateStr:  source.OVNAuditLogTemplate,
			})
	}
	return el
}

func ContainerLogPaths() string {
	return fmt.Sprintf("%q", "/var/log/pods/*/*/*.log")
}

func CollectorLogsPath(namespace string) string {
	return fmt.Sprintf("/var/log/pods/%s_%%s-*/*/*.log", namespace)
}

func ElasticSearchLogStoreLogsPath(namespace string) string {
	return fmt.Sprintf("/var/log/pods/%s_%%s-*/*/*.log", namespace)
}

func LokiLogStoreLogsPath(namespace string) string {
	return fmt.Sprintf("/var/log/pods/%s_*/%%s*/*.log", namespace)
}

func VisualizationLogsPath(namespace string) string {
	return fmt.Sprintf("/var/log/pods/%s_%%s-*/*/*.log", namespace)
}

func ExcludeContainerPaths(namespace string) string {
	return fmt.Sprintf("[%s]", strings.Join(
		[]string{
			fmt.Sprintf("%q", fmt.Sprintf(CollectorLogsPath(namespace), constants.CollectorName)),
			fmt.Sprintf("%q", fmt.Sprintf(ElasticSearchLogStoreLogsPath(namespace), constants.ElasticsearchName)),
			fmt.Sprintf("%q", fmt.Sprintf(LokiLogStoreLogsPath(namespace), constants.LokiName)),
			fmt.Sprintf("%q", fmt.Sprintf(VisualizationLogsPath(namespace), constants.KibanaName)),
			fmt.Sprintf("%q", "/var/log/pods/*/*/*.gz"),
			fmt.Sprintf("%q", "/var/log/pods/*/*/*.tmp"),
		},
		", ",
	))
}

func MetricsSources(id string) []generator.Element {
	return []generator.Element{
		InternalMetrics{
			ID:                id,
			ScrapeIntervalSec: 2,
		},
	}
}
