package azuremonitor

import (
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers/security"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	corev1 "k8s.io/api/core/v1"
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

func New(id string, o logging.OutputSpec, inputs []string, secret *corev1.Secret, op framework.Options) []framework.Element {
	dedottedID := vectorhelpers.MakeID(id, "dedot")
	if genhelper.IsDebugOutput(op) {
		return []framework.Element{
			Debug(vectorhelpers.MakeID(id, "debug"), vectorhelpers.MakeInputs(inputs...)),
		}
	}
	return framework.MergeElements(
		[]framework.Element{
			normalize.DedotLabels(dedottedID, inputs),
			Output(id, o, []string{dedottedID}, secret, op),
		},
		TLSConf(id, o, secret, op),
	)
}

func Output(id string, o logging.OutputSpec, inputs []string, secret *corev1.Secret, op framework.Options) framework.Element {
	azm := o.AzureMonitor
	return AzureMonitor{
		ComponentID:     id,
		Inputs:          vectorhelpers.MakeInputs(inputs...),
		CustomerId:      azm.CustomerId,
		LogType:         azm.LogType,
		AzureResourceId: azm.AzureResourceId,
		SharedKey:       security.GetFromSecret(secret, constants.SharedKey),
		Host:            azm.Host,
	}
}

func TLSConf(id string, o logging.OutputSpec, secret *corev1.Secret, op framework.Options) []framework.Element {
	if tlsConf := common.GenerateTLSConfWithID(id, o, secret, op, false); tlsConf != nil {
		tlsConf.NeedsEnabled = false
		return []framework.Element{tlsConf}
	}
	return []framework.Element{}
}
