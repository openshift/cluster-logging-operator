package otlp

import (
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/generator/forwarder/adapters"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/exporters"
	corev1 "k8s.io/api/core/v1"
)

// New creates an OTLP HTTP exporter for an OTLP output. The exporter sends
// logs to the URL configured on the output spec using the OTLP HTTP protocol.
//
// Returns:
//   - exporters: map of exporter ID -> configured OTLP HTTP exporter
//   - exts: map of extension ID -> configured extension (currently none)
func New(o *adapters.Output, secrets map[string]*corev1.Secret) (exportersMap api.Exporters, exts api.Extensions) {
	exportersMap = api.Exporters{}
	exts = api.Extensions{}
	if o.OTLP == nil {
		return exportersMap, exts
	}

	// The otlphttp exporter appends its own /v1/logs request path, so a URL
	// carrying the OTLP-spec recommended path suffix must be reduced to its
	// base endpoint (e.g. http://host:4318/v1/logs -> http://host:4318)
	endpoint := strings.TrimSuffix(o.OTLP.URL, "/v1/logs")
	otlphttp := exporters.NewOtlpHttp(o.Name, endpoint)
	if o.OTLP.Tuning != nil && o.OTLP.Tuning.Compression != "" {
		otlphttp.Compression = o.OTLP.Tuning.Compression
	}

	// TODO: map o.OTLP.Authentication (basic/bearer) to headers or the
	// bearertokenauth extension, and o.OutputSpec.TLS to types.TlsClientConfig
	exportersMap.Add(otlphttp.ID(), otlphttp)
	return exportersMap, exts
}
