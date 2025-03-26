package outputs

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"strings"
)

func ValidateSplunk(spec obs.OutputSpec) (results []string) {
	if spec.Splunk.PayloadKey != "" && spec.Splunk.IndexedFields != nil {
		payload := string(spec.Splunk.PayloadKey)
		for _, v := range spec.Splunk.IndexedFields {
			if !strings.HasPrefix(string(v), payload) {
				results = append(results, fmt.Sprintf("Indexed field: %s not part of payload: %s", v, payload))
			}
		}
	}
	return results
}
