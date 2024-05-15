package source

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

func NewSyslogSource(id, inputName string, input obs.InputSpec) framework.Element {
	return SyslogReceiver{
		ID:            id,
		InputName:     inputName,
		ListenAddress: helpers.ListenOnAllLocalInterfacesAddress(),
		ListenPort:    input.Receiver.Port,
	}
}

type SyslogReceiver struct {
	ID            string
	InputName     string
	ListenAddress string
	ListenPort    int32
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
mode = "tcp"
{{end}}
`
}
