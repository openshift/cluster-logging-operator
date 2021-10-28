package vector

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

func Sources(spec *logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
	return generator.MergeElements(
		LogSources(spec, op),
	)
}

func LogSources(spec *logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
	var el []generator.Element = make([]generator.Element, 0)
	types := generator.GatherSources(spec, op)
	if types.Has(logging.InputNameApplication) || types.Has(logging.InputNameInfrastructure) {
		el = append(el,
			source.KubernetesLogs{
				ComponentID:  "container_logs",
				Desc:         "Logs from containers (including openshift containers)",
				ExcludePaths: ExcludeContainerPaths(),
			})
	}
	if types.Has(logging.InputNameInfrastructure) {
		el = append(el,
			source.JournalLog{
				ComponentID:  "journal_logs",
				Desc:         "Logs from linux journal",
				TemplateName: "inputSourceJournalTemplate",
				TemplateStr:  source.JournalLogTemplate,
			})
	}
	if types.Has(logging.InputNameAudit) {
		el = append(el,
			source.HostAuditLog{
				ComponentID:  "host_audit_logs",
				Desc:         "Logs from host audit",
				TemplateName: "inputSourceHostAuditTemplate",
				TemplateStr:  source.HostAuditLogTemplate,
			},
			source.K8sAuditLog{
				ComponentID:  "k8s_audit_logs",
				Desc:         "Logs from kubernetes audit",
				TemplateName: "inputSourceK8sAuditTemplate",
				TemplateStr:  source.K8sAuditLogTemplate,
			},
			source.OpenshiftAuditLog{
				ComponentID:  "openshift_audit_logs",
				Desc:         "Logs from openshift audit",
				TemplateName: "inputSourceOpenShiftAuditTemplate",
				TemplateStr:  source.OpenshiftAuditLogTemplate,
			})
	}
	return el
}

func ContainerLogPaths() string {
	return fmt.Sprintf("%q", "/var/log/containers/*.log")
}

func CollectorLogsPath() string {
	return fmt.Sprintf("/var/log/pods/%s_%%s-*/*/*.log", constants.OpenshiftNS)
}

func LogStoreLogsPath() string {
	return fmt.Sprintf("/var/log/pods/%s_%%s-*/*/*.log", constants.OpenshiftNS)
}

func VisualizationLogsPath() string {
	return fmt.Sprintf("/var/log/pods/%s_%%s-*/*/*.log", constants.OpenshiftNS)
}

//var/log/pods/openshift-logging_collector-jhfhl_f95933e2-863f-421a-a399-a87ff1849eaf/collector/0.log
func ExcludeContainerPaths() string {
	return fmt.Sprintf("[%s]", strings.Join(
		[]string{
			fmt.Sprintf("%q", fmt.Sprintf(CollectorLogsPath(), constants.CollectorName)),
			fmt.Sprintf("%q", fmt.Sprintf(LogStoreLogsPath(), constants.ElasticsearchName)),
			fmt.Sprintf("%q", fmt.Sprintf(VisualizationLogsPath(), constants.KibanaName)),
		},
		", ",
	))
}
