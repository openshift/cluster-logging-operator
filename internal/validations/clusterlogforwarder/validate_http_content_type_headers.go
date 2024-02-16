package clusterlogforwarder

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/validations/errors"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
func validateHttpContentTypeHeaders(clf v1.ClusterLogForwarder, k8sClient client.Client, extras map[string]bool) (error, *v1.ClusterLogForwarderStatus) {
	for i, output := range clf.Spec.Outputs {
		_, output := i, output // Don't bind range variable.
		if output.Type == v1.OutputTypeHttp && output.Http != nil {
			if contentType, found := output.Http.Headers["Content-Type"]; found && validContentTypes[strings.ToLower(contentType)] == "" {
				validKeys := reflect.ValueOf(validContentTypes).MapKeys()
				log.V(3).Info("validateHttpContentTypeHeaders failed", "reason", "not valid content type set in headers",
					"content type", contentType, "supported types: ", validKeys)
				return errors.NewValidationError("not valid content type set in headers: %s , supported types: %s",
					contentType, validKeys), nil
			}
		}
	}
	return nil, nil
}
