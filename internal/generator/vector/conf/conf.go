package conf

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/input"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
	corev1 "k8s.io/api/core/v1"
)

//nolint:govet // using declarative style
func Conf(clspec *logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec *logging.ClusterLogForwarderSpec, namespace, forwarderName string, op framework.Options) []framework.Section {
	return []framework.Section{
		{
			Global(namespace, forwarderName),
			`vector global options`,
		},
		{
			source.Sources(clfspec, namespace, op),
			`
			Set of all input sources, as defined in CLF spec
			 - kubernetes_logs
			 - journald
			 - file
			 - internal_metrics
			`,
		},
		{
			normalize.NormalizeLogs(clfspec, op),
			`
			- set 'level' field
			- rename fields as per data model
			- remove unused fields
			`,
		},
		{
			input.Inputs(clfspec, op),
			`
			- Route logs by log types (app, infra, audit)
			- Handle user defined inputs
			`,
		},
		{
			Pipelines(clfspec, op),
			`
			- Add pipeline labels
			`,
		},
		{
			Outputs(clspec, secrets, clfspec, op),
			`Set of all output sinks, as defined by CLF spec
			- elasticsearch
			- loki
			- kafka
			`,
		},
	}
}
