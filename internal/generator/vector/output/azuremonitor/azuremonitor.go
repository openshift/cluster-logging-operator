package azuremonitor

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/adapters"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/common/tls"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/utils"

	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

func New(id string, o *adapters.Output, inputs []string, secrets observability.Secrets, op utils.Options) (_ string, sink types.Sink, tfs api.Transforms) {
	azm := o.AzureMonitor
	sink = sinks.NewAzureMonitorLogs(func(s *sinks.AzureMonitorLogs) {
		azureSharedKey(s, azm)
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
	return id, sink, tfs
}

func azureSharedKey(s *sinks.AzureMonitorLogs, azm *obs.AzureMonitor) {
	if azm.Authentication != nil && azm.Authentication.SharedKey != nil {
		s.SharedKey = vectorhelpers.SecretFrom(azm.Authentication.SharedKey)
	}
}
