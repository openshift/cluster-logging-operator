package fluentd

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/helpers"
	"strings"

	configv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	source2 "github.com/openshift/cluster-logging-operator/internal/generator/fluentd/source"
	"github.com/openshift/cluster-logging-operator/internal/tls"
)

func Sources(clspec *logging.CollectionSpec, spec *logging.ClusterLogForwarderSpec, namespace string, o generator.Options) []generator.Element {
	var tunings *logging.FluentdInFileSpec
	if clspec != nil && clspec.Fluentd != nil && clspec.Fluentd.InFile != nil {
		tunings = clspec.Fluentd.InFile
	}
	return generator.MergeElements(
		MetricSources(spec, o),
		LogSources(spec, tunings, namespace, o),
	)
}

func MetricSources(spec *logging.ClusterLogForwarderSpec, o generator.Options) []generator.Element {
	tlsProfileSpec := o[generator.ClusterTLSProfileSpec].(configv1.TLSProfileSpec)
	return []generator.Element{
		PrometheusMonitor{
			TlsMinVersion: helpers.TLSMinVersion(tls.MinTLSVersion(tlsProfileSpec)),
			CipherSuites:  helpers.TLSCiphers(tls.TLSCiphers(tlsProfileSpec)),
		},
	}
}

func LogSources(spec *logging.ClusterLogForwarderSpec, tunings *logging.FluentdInFileSpec, namespace string, o generator.Options) []generator.Element {
	var el []generator.Element = make([]generator.Element, 0)
	types := generator.GatherSources(spec, o)
	if types.Has(logging.InputNameInfrastructure) {
		el = append(el,
			source2.JournalLog{
				Desc:         "Logs from linux journal",
				OutLabel:     "INGRESS",
				TemplateName: "inputSourceJournalTemplate",
				TemplateStr:  source2.JournalLogTemplate,
			})
	}
	if types.Has(logging.InputNameApplication) || types.Has(logging.InputNameInfrastructure) {
		el = append(el,
			source2.ContainerLogs{
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
			source2.AuditLog{
				AuditLogLiteral: source2.AuditLogLiteral{
					Desc:         "linux audit logs",
					OutLabel:     "INGRESS",
					TemplateName: "inputSourceHostAuditTemplate",
					TemplateStr:  source2.HostAuditLogTemplate,
				},
				Tunings: tunings,
			},
			source2.AuditLog{
				AuditLogLiteral: source2.AuditLogLiteral{
					Desc:         "k8s audit logs",
					OutLabel:     "INGRESS",
					TemplateName: "inputSourceK8sAuditTemplate",
					TemplateStr:  source2.K8sAuditLogTemplate,
				},
				Tunings: tunings,
			},
			source2.AuditLog{
				AuditLogLiteral: source2.AuditLogLiteral{
					Desc:         "Openshift audit logs",
					OutLabel:     "INGRESS",
					TemplateName: "inputSourceOpenShiftAuditTemplate",
					TemplateStr:  source2.OpenshiftAuditLogTemplate,
				},
				Tunings: tunings,
			},
			source2.AuditLog{
				AuditLogLiteral: source2.AuditLogLiteral{
					Desc:         "Openshift Virtual Network (OVN) audit logs",
					OutLabel:     "INGRESS",
					TemplateName: "inputSourceOVNAuditTemplate",
					TemplateStr:  source2.OVNAuditLogTemplate,
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
