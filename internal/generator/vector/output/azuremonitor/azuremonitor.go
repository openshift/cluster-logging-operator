package azuremonitor

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
)

type AzureMonitorSink struct {
	Type            string   `toml:"type"`
	Inputs          []string `toml:"inputs"`
	CustomerId      string   `toml:"customer_id"`
	LogType         string   `toml:"log_type"`
	SharedKey       string   `toml:"shared_key"`
	AzureResourceId string   `toml:"azure_resource_id,omitempty"`
	Host            string   `toml:"host,omitempty"`
	Encoding        common.Encoding `toml:"encoding,omitempty"`
}

type AzureMonitor struct {
	ID   string
	Sink AzureMonitorSink
}

func (azm AzureMonitor) Config() any {
	return map[string]interface{}{
		"sinks": map[string]interface{}{
			azm.ID: azm.Sink,
		},
	}
}

func (azm AzureMonitor) Name() string {
	return "AzureMonitorVectorTemplate"
}

func (azm AzureMonitor) Template() string {
	return ""
}

func New(id string, o obs.OutputSpec, inputs []string, secrets observability.Secrets, strategy common.ConfigStrategy, op framework.Options) []framework.Element {
	if genhelper.IsDebugOutput(op) {
		return []framework.Element{
			Debug(vectorhelpers.MakeID(id, "debug"), vectorhelpers.MakeInputs(inputs...)),
		}
	}
	azm := o.AzureMonitor
	sink := AzureMonitorSink{
		Type:       "azure_monitor_logs",
		Inputs:     inputs,
		CustomerId: azm.CustomerId,
		LogType:    azm.LogType,
		Host:       azm.Host,
		Encoding:   common.NewEncoding(id, ""),
	}
	if azm.Authentication != nil && azm.Authentication.SharedKey != nil {
		sink.SharedKey = vectorhelpers.SecretFrom(azm.Authentication.SharedKey)
	}
	//confTLS := tls.New(id, o.TLS, secrets, op)
	return []framework.Element{
		AzureMonitor{ID: id, Sink: sink},
		//common.NewEncoding(id, ""),
		//common.NewAcknowledgments(id, strategy),
		//common.NewBatch(id, strategy),
		//common.NewBuffer(id, strategy),
		//common.NewRequest(id, strategy),
		//confTLS,
	}
}
