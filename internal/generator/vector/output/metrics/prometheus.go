package metrics

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/transforms/remap"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers/tls"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

const (
	AddNodenameToMetricTransformName = "add_nodename_to_metric"
	prometheusOutputSinkName         = "prometheus_output"
	prometheusExporterListenPort     = `24231`
)

func PrometheusOutput(inputs []string, op utils.Options) framework.Element {
	inputs = append(inputs, op.GetStringSet(framework.OptionLogsToMetricInputs)...)
	return api.NewConfig(func(config *api.Config) {
		address := fmt.Sprintf("%s:%s", helpers.ListenOnAllLocalInterfacesAddress(), prometheusExporterListenPort)
		config.Sinks[prometheusOutputSinkName] = sinks.NewPrometheusExporter(address, func(s *sinks.PrometheusExporter) {
			s.DefaultNamespace = "collector"
			s.TLS = tls.NewTls(nil, nil, op)
			if s.TLS == nil {
				s.TLS = &api.TLS{}
			}
			s.TLS.Enabled = true
			s.TLS.KeyFile = "/etc/collector/metrics/tls.key"
			s.TLS.CRTFile = "/etc/collector/metrics/tls.crt"
		}, inputs...)
	})
}

func AddNodeNameToMetric(id string, inputs []string) framework.Element {
	return remap.New(id, `.tags.hostname = get_env_var!("VECTOR_SELF_NODE_NAME")`, inputs...)
}
