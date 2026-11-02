package forwarder

import (
	"fmt"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/conf"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/utils/toml"
	"github.com/openshift/cluster-logging-operator/internal/utils/yaml"

	corev1 "k8s.io/api/core/v1"
)

type ConfigGenerator struct {
	collectorType string
}

func New(collectorType string) *ConfigGenerator {
	g := &ConfigGenerator{
		collectorType: collectorType,
	}
	return g
}

func (cg *ConfigGenerator) GenerateConf(secrets map[string]*corev1.Secret, clfspec obs.ClusterLogForwarderSpec, namespace, forwarderName string, resNames factory.ForwarderResourceNames, op utils.Options) (string, error) {
	switch cg.collectorType {
	case constants.ComponentNameOtelc:
		config := otelc.NewConfig(secrets, clfspec, namespace, forwarderName, resNames, op)
		content, err := yaml.Marshal(config)
		return string(content), err
	case constants.VectorName:
		config, err := conf.Conf(secrets, clfspec, namespace, forwarderName, resNames, op)
		if err != nil {
			return "", err
		}
		return toml.Marshal(config)
	default:
		return "", fmt.Errorf("unknown collector type: %s", cg.collectorType)
	}
}
