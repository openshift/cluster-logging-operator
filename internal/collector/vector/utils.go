package vector

import (
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"path"
)

const (
	RunVectorFile    = "run-vector.sh"
	DefaultDataPath  = "/var/lib/vector"
	ConfigFile       = "vector.toml"
	vectorConfigPath = "/etc/vector"
	entrypointValue  = "/usr/bin/run-vector.sh"
)

func GetDataPath(namespace, forwarderName string) string {
	//legacy installation
	if constants.OpenshiftNS == namespace && constants.SingletonName == forwarderName {
		return DefaultDataPath
	}
	//make data path unique to avoid collision in Multi CLF installation
	return path.Join(DefaultDataPath, namespace, forwarderName)
}
