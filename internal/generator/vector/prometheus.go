package vector

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/elements"
)

const PrometheusMonitorTemplate = `
{{define "PrometheusMonitor" -}}
# {{.Desc}}
[sources.prometheus_metrics]
  type                          = "prometheus_scrape"
  endpoints                     = ["${POD_IP}"]
  scrape_interval_secs          = 15
  tls.crt_file = '/etc/fluent/metrics/tls.crt'
  tls.key_file = '/etc/fluent/metrics/tls.key'
 
[sources.prometheus_monitor]
  type = "prometheus_remote_write"
  address = "${POD_IP}"
`

type PrometheusMonitor = generator.ConfLiteral

func PrometheusMetrics(spec *logging.ClusterLogForwarderSpec, o generator.Options) []generator.Element {
	return []generator.Element{
		elements.Pipeline{
			Desc: "Increment Prometheus metrics",
			SubElements: []generator.Element{
				generator.ConfLiteral{
					TemplateName: "EmitMetrics",
					TemplateStr:  EmitMetrics,
				},
				generator.ConfLiteral{
					TemplateName: "Forward Journal, Audit, Kubernetes Logs",
					TemplateStr:  LogsForward,
				},
			},
		},
	}
}

const EmitMetrics string = `
{{define "EmitMetrics" -}}
[source.total_bytes]
type = "internal_metrics"
scrape_interval_secs = 2
`
const LogsForward string = `
{{define "LogsForward"  -}}
{{if .Desc -}}
# {{.Desc}}
{{end -}}
[transforms.route_journal_audit]
type = "route"
inputs = [*]
route.ingress = '.tags == "journal"'
route.ingress = 'ends_with!(.tags, "audit.log")'
route.concat = 'starts_with!(.tags, "kubernetes.")'
{{end}}`
