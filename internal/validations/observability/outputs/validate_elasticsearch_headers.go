package outputs

import (
	"fmt"
	"net/http"
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

// validateElasticsearchHeaders will validate Elasticsearch custom headers
// it's not allowed to pass "Authorization" and "Content-Type" headers
// it's not allowed to have duplicate case-variant headers (e.g., 'Accept' and 'accept')
func validateElasticsearchHeaders(output obs.OutputSpec) (results []string) {
	if output.Type == obs.OutputTypeElasticsearch && output.Elasticsearch != nil && len(output.Elasticsearch.Headers) > 0 {
		forbiddenHeaders := map[string]bool{
			"Authorization": true,
			"Content-Type":  true,
		}
		var invalidHeaders []string
		canonicalHeaders := make(map[string][]string)

		for headerName := range output.Elasticsearch.Headers {
			canonicalName := http.CanonicalHeaderKey(headerName)
			if forbiddenHeaders[canonicalName] {
				invalidHeaders = append(invalidHeaders, headerName)
			}
			canonicalHeaders[canonicalName] = append(canonicalHeaders[canonicalName], headerName)
		}
		if len(invalidHeaders) > 0 {
			log.V(3).Info("validateElasticsearchHeaders failed", "reason", "invalid headers found: ", strings.Join(invalidHeaders, ","))
			results = append(results, fmt.Sprintf("invalid headers found: %s", strings.Join(invalidHeaders, ",")))
		}
		for canonicalName, originals := range canonicalHeaders {
			if len(originals) > 1 {
				log.V(3).Info("validateElasticsearchHeaders failed", "reason", "duplicate case-variant headers", "headers", originals)
				results = append(results, fmt.Sprintf("duplicate case-variant headers '%s' found, use canonical form '%s'", strings.Join(originals, "', '"), canonicalName))
			}
		}
	}
	return results
}
