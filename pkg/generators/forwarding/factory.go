package forwarding

import (
	"fmt"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/generators/forwarding/fluentd"
	corev1 "k8s.io/api/core/v1"
)

//NewConfigGenerator create a config generator for a given collector type
func NewConfigGenerator(collector logging.LogCollectionType, includeLegacyForwardConfig, includeLegacySyslogConfig, useOldRemoteSyslogPlugin bool) (ConfigGenerator, error) {
	switch collector {
	case logging.LogCollectionTypeFluentd:
		return fluentd.NewConfigGenerator(includeLegacyForwardConfig, includeLegacySyslogConfig, useOldRemoteSyslogPlugin)
	default:
		return nil, fmt.Errorf("Config generation not supported for collectors of type %s", collector)
	}
}

//ConfigGenerator is a config generator for a given ClusterLogForwarderSpec
type ConfigGenerator interface {

	//Generate the config
	Generate(clfSpec *logging.ClusterLogForwarderSpec, secrets map[string]*corev1.Secret, fwSpec *logging.ForwarderSpec) (string, error)
}
