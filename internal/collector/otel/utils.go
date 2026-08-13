package otel

import (
	"path"

	"github.com/openshift/cluster-logging-operator/internal/constants"
)

const (
	ConfigFile      = "config.yaml"
	DefaultDataPath = "/var/lib/otelcol"
	configPath      = "/etc/otelcol"
)

func GetDataPath(namespace, forwarderName string) string {
	if constants.OpenshiftNS == namespace && constants.SingletonName == forwarderName {
		return DefaultDataPath
	}
	return path.Join(DefaultDataPath, namespace, forwarderName)
}
