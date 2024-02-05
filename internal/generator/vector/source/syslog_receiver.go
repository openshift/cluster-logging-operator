package source

import (
	"strings"

	configv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/tls"
)

func NewSyslogSource(id, inputName string, input logging.InputSpec, op framework.Options) framework.Element {
	var minTlsVersion, cipherSuites string
	if _, ok := op[framework.ClusterTLSProfileSpec]; ok {
		tlsProfileSpec := op[framework.ClusterTLSProfileSpec].(configv1.TLSProfileSpec)
		minTlsVersion = tls.MinTLSVersion(tlsProfileSpec)
		cipherSuites = strings.Join(tls.TLSCiphers(tlsProfileSpec), `,`)
	}
	return SyslogReceiver{
		ID:            id,
		InputName:     inputName,
		ListenAddress: helpers.ListenOnAllLocalInterfacesAddress(),
		ListenPort:    input.Receiver.GetSyslogPort(),
		TlsMinVersion: minTlsVersion,
		CipherSuites:  cipherSuites,
	}
}

type SyslogReceiver struct {
	ID            string
	InputName     string
	ListenAddress string
	ListenPort    int32
	TlsMinVersion string
	CipherSuites  string
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

[sources.{{.ID}}.tls]
enabled = true
key_file = "/etc/collector/receiver/{{.InputName}}/tls.key"
crt_file = "/etc/collector/receiver/{{.InputName}}/tls.crt"
{{- if ne .TlsMinVersion "" }}
min_tls_version = "{{ .TlsMinVersion }}"
{{- end }}
{{- if ne .CipherSuites "" }}
ciphersuites = "{{ .CipherSuites }}"
{{- end }}
{{end}}
`
}
