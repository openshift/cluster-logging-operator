package fluentd

import (
	"fmt"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	source2 "github.com/openshift/cluster-logging-operator/internal/generator/fluentd/source"
	"strings"
)

func Sources(clspec *logging.CollectionSpec, spec *logging.ClusterLogForwarderSpec, o generator.Options) []generator.Element {
	var tunings *logging.FluentdInFileSpec
	if clspec != nil && clspec.Fluentd != nil && clspec.Fluentd.InFile != nil {
		tunings = clspec.Fluentd.InFile
	}
	return generator.MergeElements(
		MetricSources(spec, o),
		LogSources(spec, tunings, o),
	)
}

func MetricSources(spec *logging.ClusterLogForwarderSpec, o generator.Options) []generator.Element {
	return []generator.Element{
		PrometheusMonitor{
			Desc:         "Prometheus Monitoring",
			TemplateName: "PrometheusMonitor",
			TemplateStr:  PrometheusMonitorTemplate,
		},
	}
}

func LogSources(spec *logging.ClusterLogForwarderSpec, tunings *logging.FluentdInFileSpec, o generator.Options) []generator.Element {
	var el []generator.Element = make([]generator.Element, 0)
	types := generator.GatherSources(spec, o)
	if types.Has(logging.InputNameInfrastructure) {
		el = append(el,
			source2.JournalLog{
				Desc:         "Logs from linux journal",
				OutLabel:     "MEASURE",
				TemplateName: "inputSourceJournalTemplate",
				TemplateStr:  source2.JournalLogTemplate,
			})
	}
	if types.Has(logging.InputNameApplication) || types.Has(logging.InputNameInfrastructure) {
		el = append(el,
			source2.ContainerLogs{
				Desc:         "Logs from containers (including openshift containers)",
				Paths:        ContainerLogPaths(),
				ExcludePaths: ExcludeContainerPaths(),
				PosFile:      "/var/lib/fluentd/pos/es-containers.log.pos",
				OutLabel:     "MEASURE",
				Tunings:      tunings,
			})
	}
	if types.Has(logging.InputNameAudit) {
		el = append(el,
			source2.AuditLog{
				AuditLogLiteral: source2.AuditLogLiteral{
					Desc:         "linux audit logs",
					OutLabel:     "MEASURE",
					TemplateName: "inputSourceHostAuditTemplate",
					TemplateStr:  source2.HostAuditLogTemplate,
				},
				Tunings: tunings,
			},
			source2.AuditLog{
				AuditLogLiteral: source2.AuditLogLiteral{
					Desc:         "k8s audit logs",
					OutLabel:     "MEASURE",
					TemplateName: "inputSourceK8sAuditTemplate",
					TemplateStr:  source2.K8sAuditLogTemplate,
				},
				Tunings: tunings,
			},
			source2.AuditLog{
				AuditLogLiteral: source2.AuditLogLiteral{
					Desc:         "Openshift audit logs",
					OutLabel:     "MEASURE",
					TemplateName: "inputSourceOpenShiftAuditTemplate",
					TemplateStr:  source2.OpenshiftAuditLogTemplate,
				},
				Tunings: tunings,
			},
			source2.AuditLog{
				AuditLogLiteral: source2.AuditLogLiteral{
					Desc:         "Openshift Virtual Network (OVN) audit logs",
					OutLabel:     "MEASURE",
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
	return fmt.Sprintf("%q", "/var/log/pods/**/*.log")
}

func ExcludeContainerPaths() string {
	paths := []string{}
	for _, comp := range []string{constants.CollectorName, constants.ElasticsearchName, constants.KibanaName} {
		paths = append(paths, fmt.Sprintf("\"/var/log/pods/%s_%s-*/*/*.log\"", constants.OpenshiftNS, comp))
	}
	return fmt.Sprintf("[%s]", strings.Join(paths, ", "))
}
