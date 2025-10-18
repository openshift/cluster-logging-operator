package outputs

import (
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"strings"
)

// validateElasticsearchHeaders will validate Elasticsearch custom headers
// it's not allowed to pass "Authorization" and "Content-Type" headers
func validateElasticsearchHeaders(output obs.OutputSpec) (results []string) {
	if output.Type == obs.OutputTypeElasticsearch && output.Elasticsearch != nil && len(output.Elasticsearch.Headers) > 0 {
		var invalidHeaders []string
		for headerName := range output.Elasticsearch.Headers {
			if strings.ToLower(headerName) == "authorization" || strings.ToLower(headerName) == "content-type" {
				invalidHeaders = append(invalidHeaders, headerName)
			}
		}
		if len(invalidHeaders) > 0 {
			log.V(3).Info("validateElasticsearchHeaders failed", "reason", "invalid headers found: ", strings.Join(invalidHeaders, ","))
			results = append(results, fmt.Sprintf("invalid headers found: %s", strings.Join(invalidHeaders, ",")))
		}
	}
	return results
}
