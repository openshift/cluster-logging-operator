package outputs

import (
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/status"
	"github.com/openshift/cluster-logging-operator/internal/validations/clusterlogforwarder/conditions"
	corev1 "k8s.io/api/core/v1"
	"regexp"
)

func VerifyAzureMonitorLog(name string, azml *loggingv1.AzureMonitor) (bool, status.Condition) {
	pattern := "^[a-zA-Z0-9_]{1,100}$"
	// Compile the regex pattern
	reg := regexp.MustCompile(pattern)
	if azml.LogType == "" {
		return false, conditions.CondInvalid("output %q: LogType must be set.", name)
	}
	if !reg.MatchString(azml.LogType) {
		return false, conditions.CondInvalid("output %q: LogType can only contain letters, numbers, and underscores (_), and may not exceed 100 characters.", name)
	}
	if azml.CustomerId == "" {
		return false, conditions.CondInvalid("output %q: CustomerId must be set.", name)
	}
	return true, conditions.CondReady
}

func VerifySharedKeysForAzure(output *loggingv1.OutputSpec, conds loggingv1.NamedConditions, secret *corev1.Secret) bool {
	fail := func(c status.Condition) bool {
		conds.Set(output.Name, c)
		return false
	}

	if len(secret.Data[constants.SharedKey]) > 0 {
		return true
	} else {
		return fail(conditions.CondMissing("A non-empty " + constants.SharedKey + " entry is required"))
	}
}
