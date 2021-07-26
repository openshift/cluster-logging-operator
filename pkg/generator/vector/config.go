package vector

import (
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/pkg/generator"
	corev1 "k8s.io/api/core/v1"
)

//nolint:govet // using declarative style
func Conf(clspec *logging.ClusterLoggingSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, op Options) []Section {
	return []Section{
		{
			Sources(clfspec, op),
			"Set of all input sources",
		},
	}
}
