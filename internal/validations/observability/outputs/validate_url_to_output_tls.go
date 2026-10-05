package outputs

import (
	"fmt"
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/url"
)

// validateURLAccordingToTLS validate that if Output has TLS configuration Output URL scheme must be secure e.g. https, tls etc
func validateURLAccordingToTLS(output obs.OutputSpec) (results []string) {
	var specURLs []string
	switch output.Type {
	case obs.OutputTypeCloudwatch:
		specURLs = append(specURLs, output.Cloudwatch.URL)
	case obs.OutputTypeElasticsearch:
		if output.Elasticsearch.URL != "" {
			specURLs = append(specURLs, output.Elasticsearch.URL)
		}
		for _, e := range output.Elasticsearch.Endpoints {
			specURLs = append(specURLs, string(e))
		}
	case obs.OutputTypeHTTP:
		specURLs = append(specURLs, output.HTTP.URL)
	case obs.OutputTypeKafka:
		specURLs = append(specURLs, output.Kafka.URL)
	case obs.OutputTypeLoki:
		specURLs = append(specURLs, output.Loki.URL)
	case obs.OutputTypeSplunk:
		specURLs = append(specURLs, output.Splunk.URL)
	case obs.OutputTypeSyslog:
		specURLs = append(specURLs, output.Syslog.URL)
	case obs.OutputTypeOTLP:
		specURLs = append(specURLs, output.OTLP.URL)
	}

	if output.TLS != nil {
		for _, specURL := range specURLs {
			if specURL == "" {
				continue
			}
			u, _ := url.Parse(specURL)
			scheme := strings.ToLower(u.Scheme)
			if !url.IsTLSScheme(scheme) && (output.TLS.InsecureSkipVerify || output.TLS.TLSSecurityProfile != nil) {
				log.V(3).Info("validateURLAccordingToTLS failed", "reason", "URL not secure but output has TLS configuration parameters",
					"output URL", specURL, "output Name", output.Name)
				results = append(results, fmt.Sprintf("URL scheme not secure: %v, but output has TLS configuration parameters", scheme))
			}
		}
	}
	return results
}
