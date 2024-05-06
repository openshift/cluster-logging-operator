package azuremonitor

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/openshift/viaq"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

type AzureMonitor struct {
	ComponentID     string
	Inputs          string
	CustomerId      string
	LogType         string
	AzureResourceId string
	SharedKey       string
	Host            string
}

func (azm AzureMonitor) Name() string {
	return "AzureMonitorVectorTemplate"
}

func (azm AzureMonitor) Template() string {
	return `{{define "` + azm.Name() + `" -}}
[sinks.{{.ComponentID}}]
type = "azure_monitor_logs"
inputs = {{.Inputs}}
{{ if .AzureResourceId}}
azure_resource_id = "{{.AzureResourceId}}"
{{- end }}
customer_id = "{{.CustomerId}}"
{{ if .Host }}
host = "{{.Host}}"
{{- end }}
log_type = "{{.LogType}}"
shared_key = "{{.SharedKey}}"
{{end}}`
}

func New(id string, o obs.OutputSpec, inputs []string, secrets vectorhelpers.Secrets, strategy common.ConfigStrategy, op framework.Options) []framework.Element {
	dedottedID := vectorhelpers.MakeID(id, "dedot")
	if genhelper.IsDebugOutput(op) {
		return []framework.Element{
			Debug(vectorhelpers.MakeID(id, "debug"), vectorhelpers.MakeInputs(inputs...)),
		}
	}
	azm := o.AzureMonitor
	e := AzureMonitor{
		ComponentID:     id,
		Inputs:          vectorhelpers.MakeInputs(dedottedID),
		CustomerId:      azm.CustomerId,
		LogType:         azm.LogType,
		AzureResourceId: azm.AzureResourceId,
		Host:            azm.Host,
	}
	if azm.Authentication != nil {
		e.SharedKey = secrets.AsString(azm.Authentication.SharedKey)
	}
	confTLS := tls.New(id, o.TLS, secrets, op)
	return []framework.Element{
		viaq.DedotLabels(dedottedID, inputs),
		e,
		common.NewAcknowledgments(id, strategy),
		common.NewBatch(id, strategy),
		common.NewBuffer(id, strategy),
		common.NewRequest(id, strategy),
		confTLS,
	}
}
