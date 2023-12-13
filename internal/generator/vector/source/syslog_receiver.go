package source

import (
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

func NewSyslogSource(id string, input logging.InputSpec, op framework.Options) framework.Element {
	return SyslogReceiver{
		ID:            id,
		ListenAddress: helpers.ListenOnAllLocalInterfacesAddress(),
		ListenPort:    input.Receiver.Syslog.Port,
		Protocol:      input.Receiver.Syslog.Protocol,
	}
}

type SyslogReceiver struct {
	ID            string
	ListenAddress string
	ListenPort    int32
	Protocol      string
}

func (SyslogReceiver) Name() string {
	return "syslogReceiver"
}

func (i SyslogReceiver) Template() string {
	return `
{{define "` + i.Name() + `" -}}
[sources.{{.ID}}]
type = "syslog"
address = "{{.ListenAddress}}:{{.ListenPort}}"
mode = "{{.Protocol}}"
{{end}}
`
}
