package vector

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	corev1 "k8s.io/api/core/v1"
)

//nolint:govet // using declarative style
func Conf(clspec *logging.ClusterLoggingSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, op generator.Options) []generator.Section {
	return []generator.Section{
		{
			Sources(clfspec, op),
			"Set of all input sources",
		},
		{
			NormalizeLogs(clfspec, op),
			"set 'level' field, add metadata",
		},
		{
			SourcesToInputs(clfspec, op),
			"",
		},
		{
			InputsToPipelines(clfspec, op),
			"",
		},
		{
			Outputs(clspec, secrets, clfspec, op),
			"vector outputs",
		},
	}
}
