package metrics

import (
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/transforms"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types/transport"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/common/tls"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

const (
	AddNodenameToMetricTransformName = "add_nodename_to_metric"
	PrometheusOutputSinkName         = "prometheus_output"
	prometheusExporterListenPort     = `24231`
)

type PrometheusExporter struct {
	TLS *obs.TLSSpec
}

func (p PrometheusExporter) GetTlsSpec() *obs.TLSSpec {
	return p.TLS
}

func (p PrometheusExporter) IsInsecureSkipVerify() bool {
	return false
}

func (p PrometheusExporter) GetTlsSecurityProfile() *configv1.TLSSecurityProfile {
	return nil
}

func PrometheusOutput(inputs []string, op utils.Options) (sink types.Sink) {
	inputs = append(inputs, op.GetStringSet(framework.OptionLogsToMetricInputs)...)
	address := fmt.Sprintf("%s:%s", helpers.ListenOnAllLocalInterfacesAddress(), prometheusExporterListenPort)
	return sinks.NewPrometheusExporter(address, func(s *sinks.PrometheusExporter) {
		s.DefaultNamespace = "collector"
		s.TLS = tls.NewTlsEnabled(nil, nil, op)
		if s.TLS == nil {
			s.TLS = &transport.TlsEnabled{}
		}
		s.TLS.Enabled = true
		s.TLS.KeyFile = "/etc/collector/metrics/tls.key"
		s.TLS.CRTFile = "/etc/collector/metrics/tls.crt"
		tls.SetTLSProfile(&s.TLS.TLS, op)
		s.Auth = &sinks.PrometheusExporterAuth{
			Strategy: sinks.PrometheusExporterAuthStrategySar,
			Path:     "/metrics",
			Verb:     "get",
		}
	}, inputs...)
}

func AddNodeNameToMetric(inputs []string) types.Transform {
	return transforms.NewRemap(`.tags.hostname = get_env_var!("VECTOR_SELF_NODE_NAME")`, inputs...)
}
