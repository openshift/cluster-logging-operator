package source

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	helpers2 "github.com/openshift/cluster-logging-operator/internal/generator/utils"
	"strings"

	configv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/tls"
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

func LogSources(spec *logging.ClusterLogForwarderSpec, namespace string, op framework.Options) []framework.Element {
	var el []framework.Element = make([]framework.Element, 0)
	types := helpers2.GatherSources(spec, op)
	if types.Has(logging.InputNameApplication) || types.Has(logging.InputNameInfrastructure) {
		el = append(el,
			KubernetesLogs{
				ComponentID:  "raw_container_logs",
				Desc:         "Logs from containers (including openshift containers)",
				ExcludePaths: ExcludeContainerPaths(namespace),
			})
	}
	if types.Has(logging.InputNameInfrastructure) {
		el = append(el,
			JournalLog{
				ComponentID:  "raw_journal_logs",
				Desc:         "Logs from linux journal",
				TemplateName: "inputSourceJournalTemplate",
				TemplateStr:  JournalLogTemplate,
			})
	}
	if types.Has(logging.InputNameAudit) {
		el = append(el,
			HostAuditLog{
				ComponentID:  RawHostAuditLogs,
				Desc:         "Logs from host audit",
				TemplateName: "inputSourceHostAuditTemplate",
				TemplateStr:  HostAuditLogTemplate,
			},
			K8sAuditLog{
				ComponentID:  RawK8sAuditLogs,
				Desc:         "Logs from kubernetes audit",
				TemplateName: "inputSourceK8sAuditTemplate",
				TemplateStr:  K8sAuditLogTemplate,
			},
			OpenshiftAuditLog{
				ComponentID:  RawOpenshiftAuditLogs,
				Desc:         "Logs from openshift audit",
				TemplateName: "inputSourceOpenShiftAuditTemplate",
				TemplateStr:  OpenshiftAuditLogTemplate,
			},
			OVNAuditLog{
				ComponentID:  RawOvnAuditLogs,
				Desc:         "Logs from ovn audit",
				TemplateName: "inputSourceOVNAuditTemplate",
				TemplateStr:  OVNAuditLogTemplate,
			})
	}
	return el
}

func ContainerLogPaths() string {
	return fmt.Sprintf("%q", "/var/log/pods/*/*/*.log")
}

func LogFilesMetricExporterLogsPath(namespace string) string {
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
			fmt.Sprintf("%q", fmt.Sprintf(LogFilesMetricExporterLogsPath(namespace), constants.LogfilesmetricexporterName)),
			fmt.Sprintf("%q", fmt.Sprintf(ElasticSearchLogStoreLogsPath(namespace), constants.ElasticsearchName)),
			fmt.Sprintf("%q", fmt.Sprintf(LokiLogStoreLogsPath(namespace), constants.LokiName)),
			fmt.Sprintf("%q", fmt.Sprintf(VisualizationLogsPath(namespace), constants.KibanaName)),
			fmt.Sprintf("%q", fmt.Sprintf("/var/log/pods/%s_*/%s/*.log", namespace, "gateway")),
			fmt.Sprintf("%q", fmt.Sprintf("/var/log/pods/%s_*/%s/*.log", namespace, "opa")),
			fmt.Sprintf("%q", "/var/log/pods/*/*/*.gz"),
			fmt.Sprintf("%q", "/var/log/pods/*/*/*.tmp"),
		},
		", ",
	))
}

func HttpSources(spec *logging.ClusterLogForwarderSpec, op framework.Options) []framework.Element {
	var minTlsVersion, cipherSuites string
	if _, ok := op[framework.ClusterTLSProfileSpec]; ok {
		tlsProfileSpec := op[framework.ClusterTLSProfileSpec].(configv1.TLSProfileSpec)
		minTlsVersion = tls.MinTLSVersion(tlsProfileSpec)
		cipherSuites = strings.Join(tls.TLSCiphers(tlsProfileSpec), `,`)
	}

	el := []framework.Element{}
	for _, input := range spec.Inputs {
		if input.Receiver != nil && input.Receiver.HTTP != nil {
			el = append(el, HttpReceiver{
				ID:            input.Name,
				ListenAddress: helpers.ListenOnAllLocalInterfacesAddress(),
				ListenPort:    input.Receiver.HTTP.GetPort(),
				Format:        input.Receiver.HTTP.Format,
				TlsMinVersion: minTlsVersion,
				CipherSuites:  cipherSuites,
			})
		}
	}
	return el
}

type HttpReceiver struct {
	ID            string
	ListenAddress string
	ListenPort    int32
	Format        string
	TlsMinVersion string
	CipherSuites  string
}

func (HttpReceiver) Name() string {
	return "httpReceiver"
}

func (i HttpReceiver) Template() string {
	return `
{{define "` + i.Name() + `" -}}
[sources.{{.ID}}]
type = "http_server"
address = "{{.ListenAddress}}:{{.ListenPort}}"
decoding.codec = "json"

[sources.{{.ID}}.tls]
enabled = true
key_file = "/etc/collector/{{.ID}}/tls.key"
crt_file = "/etc/collector/{{.ID}}/tls.crt"
{{- if ne .TlsMinVersion "" }}
min_tls_version = "{{ .TlsMinVersion }}"
{{- end }}
{{- if ne .CipherSuites "" }}
ciphersuites = "{{ .CipherSuites }}"
{{- end }}

[transforms.{{.ID}}_split]
type = "remap"
inputs = ["{{.ID}}"]
source = '''
  if exists(.items) && is_array(.items) {. = unnest!(.items)} else {.}
'''

[transforms.{{.ID}}_items]
type = "remap"
inputs = ["{{.ID}}_split"]
source = '''
  if exists(.items) {. = .items} else {.}
'''
{{end}}
`
}

func MetricsSources(id string) []framework.Element {
	return []framework.Element{
		InternalMetrics{
			ID:                id,
			ScrapeIntervalSec: 2,
		},
	}
}
