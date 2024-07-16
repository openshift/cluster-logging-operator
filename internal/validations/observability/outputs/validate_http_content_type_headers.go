package outputs

import (
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"reflect"
	"strings"
)

var validContentTypes = map[string]string{
	"application/json":     "json",
	"application/x-ndjson": "ndjson",
}

// validateHttpContentTypeHeaders will validate Content-Type header in Http Output
// valid content-type are: "application/json" and "application/x-ndjson"
// was introduced in https://github.com/openshift/cluster-logging-operator/pull/1924
// for https://issues.redhat.com/browse/LOG-3784
func validateHttpContentTypeHeaders(output obs.OutputSpec) (results []string) {
	if output.Type == obs.OutputTypeHTTP && output.HTTP != nil {
		if contentType, found := output.HTTP.Headers["Content-Type"]; found && validContentTypes[strings.ToLower(contentType)] == "" {
			validKeys := reflect.ValueOf(validContentTypes).MapKeys()
			log.V(3).Info("validateHttpContentTypeHeaders failed", "reason", "not valid content type set in headers",
				"content type", contentType, "supported types: ", validKeys)
			results = append(results, fmt.Sprintf("not valid content type set in headers: %s , supported types: %s", contentType, validKeys))
		}
	}
	return results
}
