package source

import (
	"fmt"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	common "github.com/openshift/cluster-logging-operator/internal/generator/vector"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
	"github.com/openshift/cluster-logging-operator/internal/utils/sets"
)

const (
	AuditHost       = "audit_host"
	AuditKubernetes = "audit_k8s"
	AuditOpenShift  = "audit_openshift"
	AuditOVN        = "audit_ovn"
)

var (
	AuditLogTypes     = []string{AuditHost, AuditKubernetes, AuditOpenShift, AuditOVN}
	ContainerLogTypes = []string{logging.InputNameContainer, logging.InputNameApplication, logging.InputNameInfrastructure}
	JournalLogTypes   = []string{logging.InputNameNode, logging.InputNameInfrastructure}
)

func MakeID(name string) string {
	return fmt.Sprintf("input_%s", name)
}

func MakeIDsFor(source string, types sets.String) []string {
	switch {
	case types.Has(logging.InputNameContainer) || types.Has(logging.InputNameApplication) || types.Has(logging.InputNameInfrastructure):
		return []string{MakeID(logging.InputNameContainer)}
	case types.Has(logging.InputNameNode) || types.Has(logging.InputNameInfrastructure):
		return []string{MakeID(logging.InputNameNode)}
	case types.Has(logging.InputNameAudit):
		return []string{MakeID(AuditHost), MakeID(AuditKubernetes), MakeID(AuditOpenShift), MakeID(AuditOVN)}
	}
	return []string{}
}

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
				ComponentID:  MakeID(logging.InputNameContainer),
				Desc:         "Logs from containers (including openshift containers)",
				ExcludePaths: common.ExcludeContainerPaths(namespace),
			})
	}
	if types.Has(logging.InputNameNode) || types.Has(logging.InputNameInfrastructure) {
		el = append(el,
			source.JournalLog{
				ComponentID:  MakeID(logging.InputNameNode),
				Desc:         "Logs from linux journal",
				TemplateName: "inputSourceJournalTemplate",
				TemplateStr:  source.JournalLogTemplate,
			})
	}
	if types.Has(logging.InputNameAudit) {
		el = append(el,
			source.HostAuditLog{
				ComponentID:  MakeID(AuditHost),
				Desc:         "Logs from host audit",
				TemplateName: "inputSourceHostAuditTemplate",
				TemplateStr:  source.HostAuditLogTemplate,
			},
			source.K8sAuditLog{
				ComponentID:  MakeID(AuditKubernetes),
				Desc:         "Logs from kubernetes audit",
				TemplateName: "inputSourceK8sAuditTemplate",
				TemplateStr:  source.K8sAuditLogTemplate,
			},
			source.OpenshiftAuditLog{
				ComponentID:  MakeID(AuditOpenShift),
				Desc:         "Logs from openshift audit",
				TemplateName: "inputSourceOpenShiftAuditTemplate",
				TemplateStr:  source.OpenshiftAuditLogTemplate,
			},
			source.OVNAuditLog{
				ComponentID:  MakeID(AuditOVN),
				Desc:         "Logs from ovn audit",
				TemplateName: "inputSourceOVNAuditTemplate",
				TemplateStr:  source.OVNAuditLogTemplate,
			})
	}
	return el
}
