package forwarding

import (
	"fmt"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/generators/forwarding/fluentbit"
	"github.com/openshift/cluster-logging-operator/pkg/generators/forwarding/fluentd"
	"k8s.io/apimachinery/pkg/util/sets"
)

//SourceStrategy provides a mechanism to change how a given config generator
//produces source configuration
type SourceStrategy func(sources sets.String) (results []string, err error)

//NewConfigGenerator create a config generator for a given collector type
func NewConfigGenerator(collector logging.LogCollectionType, includeLegacyForwardConfig, includeLegacySyslogConfig, useOldRemoteSyslogPlugin bool, topology string) (ConfigGenerator, error) {
	switch collector {
	case logging.LogCollectionTypeFluentd:
		return fluentd.NewConfigGenerator(includeLegacyForwardConfig, includeLegacySyslogConfig, useOldRemoteSyslogPlugin, topology)
	case logging.LogCollectionTypeFluentbit:
		return fluentbit.NewConfigGenerator(includeLegacyForwardConfig, includeLegacySyslogConfig, useOldRemoteSyslogPlugin)
	default:
		return nil, fmt.Errorf("Config generation not supported for collectors of type %s", collector)
	}
}

//ConfigGenerator is a config generator for a given ClusterLogForwarderSpec
type ConfigGenerator interface {

	//Generate the config
	Generate(clfSpec *logging.ClusterLogForwarderSpec, fwSpec *logging.ForwarderSpec) (string, error)
}
