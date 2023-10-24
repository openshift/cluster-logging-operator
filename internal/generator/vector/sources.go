package vector

import (
	"fmt"
	"strings"

	configv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
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

func Sources(spec *logging.ClusterLogForwarderSpec, namespace string, op generator.Options) []generator.Element {
	return generator.MergeElements(
		LogSources(spec, namespace, op),
		HttpSources(spec, op),
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

func HttpSources(spec *logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
	var minTlsVersion, cipherSuites string
	if _, ok := op[generator.ClusterTLSProfileSpec]; ok {
		tlsProfileSpec := op[generator.ClusterTLSProfileSpec].(configv1.TLSProfileSpec)
		minTlsVersion = tls.MinTLSVersion(tlsProfileSpec)
		cipherSuites = strings.Join(tls.TLSCiphers(tlsProfileSpec), `,`)
	}

	el := []generator.Element{}
	for _, input := range spec.Inputs {
		if input.Receiver != nil {
			if input.Receiver.HTTP != nil {
				el = append(el, HttpReceiver{
					ID:            input.Name,
					ListenAddress: helpers.ListenOnAllLocalInterfacesAddress(),
					ListenPort:    input.Receiver.HTTP.GetPort(),
					Format:        input.Receiver.HTTP.Format,
					TlsMinVersion: minTlsVersion,
					CipherSuites:  cipherSuites,
				})
			}
			if input.Receiver.Syslog != nil {
				el = append(el, SyslogReceiver{
					ID:            input.Name,
					ListenAddress: helpers.ListenOnAllLocalInterfacesAddress(),
					ListenPort:    input.Receiver.HTTP.GetPort(),
					Format:        input.Receiver.HTTP.Format,
				})
			}
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

type SyslogReceiver struct {
	ID            string
	ListenAddress string
	ListenPort    int32
	Format        string
}

func (SyslogReceiver) Name() string {
	return "syslogReceiver"
}

func (i SyslogReceiver) Template() string {
	return `
{{define "` + i.Name() + `" -}}
[sources.{{.ID}}]
type = "syslog"
address = "{{.ListenAddress}}:{{.ListenPort}}"

[sources.{{.ID}}.tls]
enabled = false
key_file = "/etc/collector/{{.ID}}/tls.key"
crt_file = "/etc/collector/{{.ID}}/tls.crt"

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

func MetricsSources(id string) []generator.Element {
	return []generator.Element{
		InternalMetrics{
			ID:                id,
			ScrapeIntervalSec: 2,
		},
	}
}
