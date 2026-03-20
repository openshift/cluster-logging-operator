package conf

import (
	"github.com/openshift/cluster-logging-operator/internal/collector/vector"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

func Global(c *api.Config, namespace, forwarderName string) *api.Config {
	dataDir := vector.GetDataPath(namespace, forwarderName)
	if dataDir == vector.DefaultDataPath {
		dataDir = ""
	}
	c.Api = &api.Api{Enabled: true}
	c.LogSchema = &api.LogSchema{HostKey: "hostname"}
	c.Global = api.Global{
		ExpireMetricsSec: 60,
		DataDir:          dataDir,
	}
	c.Secret = map[string]*api.Secret{
		helpers.VectorSecretID: api.NewDirectorySecret(constants.CollectorSecretsDir),
	}
	return c
}
