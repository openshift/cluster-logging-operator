package vector

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	source2 "github.com/openshift/cluster-logging-operator/internal/generator/fluentd/source"
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
			source2.JournalLog{
				ComponentID:  "journal_logs",
				Desc:         "Logs from linux journal",
				TemplateName: "inputSourceJournalTemplate",
				TemplateStr:  source2.JournalLogTemplate,
			})
	}
	if types.Has(logging.InputNameAudit) {
		el = append(el,
			source2.HostAuditLog{
				ComponentID:  "host_audit_logs",
				Desc:         "Logs from host audit",
				TemplateName: "inputSourceHostAuditTemplate",
				TemplateStr:  source2.HostAuditLogTemplate,
			},
			source2.K8sAuditLog{
				ComponentID:  "k8s_audit_logs",
				Desc:         "Logs from kubernetes audit",
				TemplateName: "inputSourceK8sAuditTemplate",
				TemplateStr:  source2.K8sAuditLogTemplate,
			},
			source2.OpenshiftAuditLog{
				ComponentID:  "openshift_audit_logs",
				Desc:         "Logs from openshift audit",
				TemplateName: "inputSourceOpenShiftAuditTemplate",
				TemplateStr:  source2.OpenshiftAuditLogTemplate,
			})
	}
	return el
}

func ContainerLogPaths() string {
	return fmt.Sprintf("%q", "/var/log/containers/*.log")
}

func ExcludeContainerPaths() string {
	return fmt.Sprintf("[%s]", strings.Join(
		[]string{
			fmt.Sprintf("%q", fmt.Sprintf(generator.CollectorLogsPath(), constants.CollectorName)),
			fmt.Sprintf("%q", fmt.Sprintf(generator.LogStoreLogsPath(), constants.ElasticsearchName)),
			fmt.Sprintf("%q", generator.VisualizationLogsPath()),
		},
		", ",
	))
}
