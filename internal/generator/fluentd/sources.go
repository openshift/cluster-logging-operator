package fluentd

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/utils"
	"strings"

	configv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/source"
	"github.com/openshift/cluster-logging-operator/internal/tls"
)

func Sources(clspec *logging.CollectionSpec, spec *logging.ClusterLogForwarderSpec, namespace string, o framework.Options) []framework.Element {
	var tunings *logging.FluentdInFileSpec
	if clspec != nil && clspec.Fluentd != nil && clspec.Fluentd.InFile != nil {
		tunings = clspec.Fluentd.InFile
	}
	return framework.MergeElements(
		MetricSources(spec, o),
		LogSources(spec, tunings, namespace, o),
	)
}

func MetricSources(spec *logging.ClusterLogForwarderSpec, o framework.Options) []framework.Element {
	tlsProfileSpec := o[framework.ClusterTLSProfileSpec].(configv1.TLSProfileSpec)
	return []framework.Element{
		PrometheusMonitor{
			TlsMinVersion: helpers.TLSMinVersion(tls.MinTLSVersion(tlsProfileSpec)),
			CipherSuites:  helpers.TLSCiphers(tls.TLSCiphers(tlsProfileSpec)),
		},
	}
}

func LogSources(spec *logging.ClusterLogForwarderSpec, tunings *logging.FluentdInFileSpec, namespace string, o framework.Options) []framework.Element {
	var el []framework.Element = make([]framework.Element, 0)
	types := utils.GatherSources(spec, o)
	if types.Has(logging.InputNameInfrastructure) {
		el = append(el,
			source.JournalLog{
				Desc:         "Logs from linux journal",
				OutLabel:     "INGRESS",
				TemplateName: "inputSourceJournalTemplate",
				TemplateStr:  source.JournalLogTemplate,
			})
	}
	if types.Has(logging.InputNameApplication) || types.Has(logging.InputNameInfrastructure) {
		el = append(el,
			source.ContainerLogs{
				Desc:         "Logs from containers (including openshift containers)",
				Paths:        ContainerLogPaths(),
				ExcludePaths: ExcludeContainerPaths(namespace),
				PosFile:      "/var/lib/fluentd/pos/es-containers.log.pos",
				OutLabel:     "CONCAT",
				Tunings:      tunings,
			})
	}
	if types.Has(logging.InputNameAudit) {
		el = append(el,
			source.AuditLog{
				AuditLogLiteral: source.AuditLogLiteral{
					Desc:         "linux audit logs",
					OutLabel:     "INGRESS",
					TemplateName: "inputSourceHostAuditTemplate",
					TemplateStr:  source.HostAuditLogTemplate,
				},
				Tunings: tunings,
			},
			source.AuditLog{
				AuditLogLiteral: source.AuditLogLiteral{
					Desc:         "k8s audit logs",
					OutLabel:     "INGRESS",
					TemplateName: "inputSourceK8sAuditTemplate",
					TemplateStr:  source.K8sAuditLogTemplate,
				},
				Tunings: tunings,
			},
			source.AuditLog{
				AuditLogLiteral: source.AuditLogLiteral{
					Desc:         "Openshift audit logs",
					OutLabel:     "INGRESS",
					TemplateName: "inputSourceOpenShiftAuditTemplate",
					TemplateStr:  source.OpenshiftAuditLogTemplate,
				},
				Tunings: tunings,
			},
			source.AuditLog{
				AuditLogLiteral: source.AuditLogLiteral{
					Desc:         "Openshift Virtual Network (OVN) audit logs",
					OutLabel:     "INGRESS",
					TemplateName: "inputSourceOVNAuditTemplate",
					TemplateStr:  source.OVNAuditLogTemplate,
				},
				Tunings: tunings,
			},
		)
	}
	return el
}

func ContainerLogPaths() string {
	return fmt.Sprintf("%q", "/var/log/pods/*/*/*.log")
}

func ExcludeContainerPaths(namespace string) string {
	paths := []string{}
	for _, comp := range []string{constants.CollectorName, constants.LogfilesmetricexporterName, constants.ElasticsearchName, constants.KibanaName} {
		paths = append(paths, fmt.Sprintf("\"/var/log/pods/%s_%s-*/*/*.log\"", namespace, comp))
	}
	paths = append(paths, fmt.Sprintf("\"/var/log/pods/%s_*/%s*/*.log\"", namespace, constants.LokiName))
	// in loki stack there 2 container without 'loki' as prefix in name: "gateway" and "opa"
	paths = append(paths, fmt.Sprintf("\"/var/log/pods/%s_*/%s/*.log\"", namespace, "gateway"))
	paths = append(paths, fmt.Sprintf("\"/var/log/pods/%s_*/%s/*.log\"", namespace, "opa"))
	paths = append(paths, "\"/var/log/pods/*/*/*.gz\"", "\"/var/log/pods/*/*/*.tmp\"")

	return fmt.Sprintf("[%s]", strings.Join(paths, ", "))
}
