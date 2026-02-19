package azuremonitor

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/common/tls"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/utils"

	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

func New(id string, o *observability.Output, inputs []string, secrets observability.Secrets, op utils.Options) []framework.Element {
	if genhelper.IsDebugOutput(op) {
		return []framework.Element{
			elements.Debug(vectorhelpers.MakeID(id, "debug"), vectorhelpers.MakeInputs(inputs...)),
		}
	}
	azm := o.AzureMonitor
	return []framework.Element{
		api.NewConfig(func(c *api.Config) {
			c.Sinks[id] = sinks.NewAzureMonitorLogs(func(s *sinks.AzureMonitorLogs) {
				azureSharedKey(s, azm, secrets)
				s.CustomerId = azm.CustomerId
				s.LogType = azm.LogType
				s.AzureResourceId = azm.AzureResourceId
				s.Host = azm.Host
				s.Encoding = common.NewApiEncoding("")
				s.Batch = common.NewApiBatch(o)
				s.Buffer = common.NewApiBuffer(o)
				s.Request = common.NewApiRequest(o)
				s.TLS = tls.NewTls(o, secrets, op)
			}, inputs...)
		}),
	}
}

func azureSharedKey(s *sinks.AzureMonitorLogs, azm *obs.AzureMonitor, secrets observability.Secrets) {
	if azm.Authentication != nil && azm.Authentication.SharedKey != nil {
		s.SharedKey = vectorhelpers.SecretFrom(azm.Authentication.SharedKey)
	}
}
