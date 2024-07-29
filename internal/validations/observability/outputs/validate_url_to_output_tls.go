package outputs

import (
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/url"
	"strings"
)

// validateURLAccordingToTLS validate that if Output has TLS configuration Output URL scheme must be secure e.g. https, tls etc
func validateURLAccordingToTLS(output obs.OutputSpec) (results []string) {
	specURL := ""
	switch output.Type {
	case obs.OutputTypeCloudwatch:
		if output.Cloudwatch.URL != nil {
			specURL = *output.Cloudwatch.URL
		}
	case obs.OutputTypeElasticsearch:
		specURL = output.Elasticsearch.URL
	case obs.OutputTypeHTTP:
		specURL = output.HTTP.URL
	case obs.OutputTypeKafka:
		if output.Kafka.URL != nil {
			specURL = *output.Kafka.URL
		}
	case obs.OutputTypeLoki:
		specURL = output.Loki.URL
	case obs.OutputTypeSplunk:
		specURL = output.Splunk.URL
	case obs.OutputTypeSyslog:
		specURL = output.Syslog.URL
	case obs.OutputTypeOTLP:
		specURL = output.OTLP.URL
	}

	// some outputs not require to have output URL (e.g. Amazon CloudWatch or Google Cloud Logging)
	if specURL != "" && output.TLS != nil {
		u, _ := url.Parse(specURL)
		scheme := strings.ToLower(u.Scheme)
		if !url.IsTLSScheme(scheme) && (output.TLS.InsecureSkipVerify || output.TLS.TLSSecurityProfile != nil) {
			log.V(3).Info("validateURLAccordingToTLS failed", "reason", "URL not secure but output has TLS configuration parameters",
				"output URL", specURL, "output Name", output.Name)
			results = append(results, fmt.Sprintf("URL scheme not secure: %v, but output has TLS configuration parameters", scheme))
		}
	}
	return results
}
