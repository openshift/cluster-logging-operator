package forwarding

import (
	"fmt"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/generators/forwarding/fluentd"
	"github.com/openshift/cluster-logging-operator/pkg/generators/forwarding/sources"
)

//NewConfigGenerator create a config generator for a given collector type
func NewConfigGenerator(collector logging.LogCollectionType, includeLegacyForwardConfig, includeLegacySyslogConfig, useOldRemoteSyslogPlugin bool) (ConfigGenerator, error) {
	switch collector {
	case logging.LogCollectionTypeFluentd:
		return fluentd.NewConfigGenerator(includeLegacyForwardConfig, includeLegacySyslogConfig, useOldRemoteSyslogPlugin)
	case logging.LogCollectionTypeFluentbit:
		return fluentd.NewConfigGeneratorWithSourceGenerator(includeLegacyForwardConfig, includeLegacySyslogConfig, useOldRemoteSyslogPlugin, sources.GenerateLocalFluentForwardSource)
	default:
		return nil, fmt.Errorf("Config generation not supported for collectors of type %s", collector)
	}
}

//ConfigGenerator is a config generator for a given ClusterLogForwarderSpec
type ConfigGenerator interface {

	//Generate the config
	Generate(clfSpec *logging.ClusterLogForwarderSpec, fwSpec *logging.ForwarderSpec) (string, error)
}
