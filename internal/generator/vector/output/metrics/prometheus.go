package metrics

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

const (
	AddNodenameToMetricTransformName = "add_nodename_to_metric"
	PrometheusOutputSinkName         = "prometheus_output"
	PrometheusExporterListenPort     = `24231`
)

type PrometheusExporter struct {
	ID            string
	Inputs        string
	Address       string
	TlsMinVersion string
	CipherSuites  string
}

func (p PrometheusExporter) Name() string {
	return "PrometheusExporterTemplate"
}

func (p PrometheusExporter) Template() string {
	return `{{define "` + p.Name() + `" -}}
[sinks.{{.ID}}]
type = "prometheus_exporter"
inputs = {{.Inputs}}
address = "{{.Address}}"
default_namespace = "collector"

[sinks.{{.ID}}.tls]

key_file = "/etc/collector/metrics/tls.key"
crt_file = "/etc/collector/metrics/tls.crt"
min_tls_version = "{{.TlsMinVersion}}"
ciphersuites = "{{.CipherSuites}}"
{{end}}`
}

type AddNodenameToMetric struct {
	ID     string
	Inputs string
}

func (a AddNodenameToMetric) Name() string {
	return AddNodenameToMetricTransformName
}

func (a AddNodenameToMetric) Template() string {
	return `{{define "` + a.Name() + `" -}}
[transforms.{{.ID}}]
type = "remap"
inputs = {{.Inputs}}
source = '''
.tags.hostname = get_env_var!("VECTOR_SELF_NODE_NAME")
'''
{{end}}`
}

func PrometheusOutput(id string, inputs []string, minTlsVersion string, cipherSuites string) framework.Element {
	return PrometheusExporter{
		ID:            id,
		Inputs:        helpers.MakeInputs(inputs...),
		Address:       helpers.ListenOnAllLocalInterfacesAddress() + `:` + PrometheusExporterListenPort,
		TlsMinVersion: minTlsVersion,
		CipherSuites:  cipherSuites,
	}
}

func AddNodeNameToMetric(id string, inputs []string) framework.Element {
	return AddNodenameToMetric{
		ID:     id,
		Inputs: helpers.MakeInputs(inputs...),
	}
}
