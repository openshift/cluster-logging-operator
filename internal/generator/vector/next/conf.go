package next

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	common "github.com/openshift/cluster-logging-operator/internal/generator/vector"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/next/outputs"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/next/source"
	corev1 "k8s.io/api/core/v1"
)

//nolint:govet // using declarative style
func Conf(clspec *logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, namespace, forwarderName string, op generator.Options) []generator.Section {
	return []generator.Section{
		{
			common.Global(namespace, forwarderName),
			`vector global options`,
		},
		{
			source.Sources(*clfspec, namespace, op),
			`
			Set of all input sources, as defined in CLF spec
			 - kubernetes_logs
			 - journald
			 - file
			 - internal_metrics
			`,
		},
		{
			Pipelines(clfspec, op),
			`
			- Add pipeline labels
			`,
		},
		{
			outputs.New(clspec, secrets, clfspec, op),
			`Set of all output sinks, as defined by CLF spec
			- elasticsearch
			- loki
			- kafka
			`,
		},
	}
}
