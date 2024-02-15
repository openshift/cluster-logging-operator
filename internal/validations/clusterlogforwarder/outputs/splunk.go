package outputs

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/status"
	"github.com/openshift/cluster-logging-operator/internal/validations/clusterlogforwarder/conditions"
	"k8s.io/api/core/v1"
	"regexp"
)

const pattern = `^([a-zA-Z0-9_]+|"(?:\\.|[^"\\])*")(?:\.([a-zA-Z0-9_]+|"(?:\\.|[^"\\])*"))*$`

func VerifySplunk(name string, splunk *loggingv1.Splunk) (bool, status.Condition) {
	if splunk != nil {
		if splunk.IndexKey != "" && splunk.IndexName != "" {
			log.V(3).Info("verifyOutputsFailed", "reason", "splunk output allows only one of indexKey or indexName, not both.")
			return false, conditions.CondInvalid("output %q: Only one of indexKey or indexName can be set, not both.", name)
		}

		if splunk.IndexKey != "" && !isIndexKeyMatch(splunk.IndexKey) {
			return false, conditions.CondInvalid("output %q: IndexKey can only contain letters, numbers, and underscores (a-zA-Z0-9_). "+
				"Segments that contain characters outside of this range must be quoted.", name)
		}
	}
	return true, conditions.CondReady
}

func VerifySecretKeysForSplunk(output *loggingv1.OutputSpec, conds loggingv1.NamedConditions, secret *v1.Secret) bool {
	fail := func(c status.Condition) bool {
		conds.Set(output.Name, c)
		return false
	}

	if len(secret.Data[constants.SplunkHECTokenKey]) > 0 {
		return true
	} else {
		return fail(conditions.CondMissing("A non-empty " + constants.SplunkHECTokenKey + " entry is required"))
	}
}

func isIndexKeyMatch(indexKey string) bool {
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(indexKey)
}
