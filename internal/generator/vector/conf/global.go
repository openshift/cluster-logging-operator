package conf

import (
	"github.com/openshift/cluster-logging-operator/internal/collector/vector"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

func Global(namespace, forwarderName string) []framework.Element {
	dataDir := vector.GetDataPath(namespace, forwarderName)
	if dataDir == vector.DefaultDataPath {
		dataDir = ""
	}
	return []framework.Element{
		api.Config{
			Api: &api.Api{Enabled: true},
			Global: api.Global{
				ExpireMetricsSec: 60,
				DataDir:          dataDir,
			},
			Secret: map[string]interface{}{
				helpers.VectorSecretID: api.NewDirectorySecret(constants.CollectorSecretsDir),
			},
		},
	}
}
