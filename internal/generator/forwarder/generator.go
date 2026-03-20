package forwarder

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/conf"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/utils/toml"

	corev1 "k8s.io/api/core/v1"
)

type ConfigGenerator struct {
	conf func(secrets map[string]*corev1.Secret, clfspec obs.ClusterLogForwarderSpec, namespace, forwarderName string, resNames factory.ForwarderResourceNames, op utils.Options) *api.Config
}

func New() *ConfigGenerator {
	g := &ConfigGenerator{
		conf: conf.Conf,
	}
	return g
}

func (cg *ConfigGenerator) GenerateConf(secrets map[string]*corev1.Secret, clfspec obs.ClusterLogForwarderSpec, namespace, forwarderName string, resNames factory.ForwarderResourceNames, op framework.Options) (string, error) {
	config := cg.conf(secrets, clfspec, namespace, forwarderName, resNames, op)
	return toml.Marshal(config)
}
