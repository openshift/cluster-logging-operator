package vector

import (
	"path"

	"github.com/openshift/cluster-logging-operator/internal/constants"

	_ "embed"
)

const (
	RunVectorFile        = "run-vector.sh"
	DefaultDataPath      = "/var/lib/vector"
	ConfigFile           = "vector.toml"
	vectorConfigPath     = "/etc/vector"
	entrypointValue      = "/usr/bin/run-vector.sh"
	SecretDataReaderFile = "read_secret_data.sh"
	SecretDataReaderPath = "/usr/bin/" + "read_secret_data.sh"
)

// RunVectorScript is the run-vector.sh script for launching the Vector container process
// will override content of /scripts/run-vector.sh in https://github.com/ViaQ/vector
//
//go:embed run-vector.sh
var RunVectorScript string

func GetDataPath(namespace, forwarderName string) string {
	//legacy installation
	if constants.OpenshiftNS == namespace && constants.SingletonName == forwarderName {
		return DefaultDataPath
	}
	//make data path unique to avoid collision in Multi CLF installation
	return path.Join(DefaultDataPath, namespace, forwarderName)
}
