package source

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	helpers2 "github.com/openshift/cluster-logging-operator/internal/generator/utils"
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

func Sources(spec *logging.ClusterLogForwarderSpec, namespace string, op framework.Options) []framework.Element {
	return framework.MergeElements(
		LogSources(spec, namespace, op),
		HttpSources(spec, op),
		MetricsSources(InternalMetricsSourceName),
	)
}

// LogSources generates configuration of the sources to collect excluding the logs where the collector is deployed
func LogSources(spec *logging.ClusterLogForwarderSpec, namespace string, op framework.Options) []framework.Element {
	var el []framework.Element = make([]framework.Element, 0)
	types := helpers2.GatherSources(spec, op)
	if types.Has(logging.InputNameApplication) || types.Has(logging.InputNameInfrastructure) {
		el = append(el,
			NewKubernetesLogsForOpenShiftLogging("raw_container_logs"),
		)
	}
	if types.Has(logging.InputNameInfrastructure) {
		el = append(el,
			NewJournalLog("raw_journal_logs"),
		)
	}
	if types.Has(logging.InputNameAudit) {
		el = append(el,
			NewHostAuditLog(RawHostAuditLogs),
			NewK8sAuditLog(RawK8sAuditLogs),
			NewOpenshiftAuditLog(RawOpenshiftAuditLogs),
			NewOVNAuditLog(RawOvnAuditLogs))
	}
	return el
}

func MetricsSources(id string) []framework.Element {
	return []framework.Element{
		InternalMetrics{
			ID:                id,
			ScrapeIntervalSec: 2,
		},
	}
}
