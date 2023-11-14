package source

import (
	configv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	"strings"
)

func HttpSources(spec *logging.ClusterLogForwarderSpec, op framework.Options) []framework.Element {
	el := []framework.Element{}
	for _, input := range spec.Inputs {
		if input.Receiver != nil && input.Receiver.HTTP != nil {
			el = append(el, NewHttpSource(input.Name, input, op))
		}
	}
	return el
}

func NewHttpSource(id string, input logging.InputSpec, op framework.Options) framework.Element {
	var minTlsVersion, cipherSuites string
	if _, ok := op[framework.ClusterTLSProfileSpec]; ok {
		tlsProfileSpec := op[framework.ClusterTLSProfileSpec].(configv1.TLSProfileSpec)
		minTlsVersion = tls.MinTLSVersion(tlsProfileSpec)
		cipherSuites = strings.Join(tls.TLSCiphers(tlsProfileSpec), `,`)
	}
	return HttpReceiver{
		ID:            id,
		ListenAddress: helpers.ListenOnAllLocalInterfacesAddress(),
		ListenPort:    input.Receiver.HTTP.Port,
		Format:        input.Receiver.HTTP.Format,
		TlsMinVersion: minTlsVersion,
		CipherSuites:  cipherSuites,
	}
}

type HttpReceiver struct {
	ID            string
	ListenAddress string
	ListenPort    int32
	Format        string
	TlsMinVersion string
	CipherSuites  string
}

func (HttpReceiver) Name() string {
	return "httpReceiver"
}

func (i HttpReceiver) Template() string {
	return `
{{define "` + i.Name() + `" -}}
[sources.{{.ID}}]
type = "http_server"
address = "{{.ListenAddress}}:{{.ListenPort}}"
decoding.codec = "json"

[sources.{{.ID}}.tls]
enabled = true
key_file = "/etc/collector/{{.ID}}/tls.key"
crt_file = "/etc/collector/{{.ID}}/tls.crt"
{{- if ne .TlsMinVersion "" }}
min_tls_version = "{{ .TlsMinVersion }}"
{{- end }}
{{- if ne .CipherSuites "" }}
ciphersuites = "{{ .CipherSuites }}"
{{- end }}

[transforms.{{.ID}}_split]
type = "remap"
inputs = ["{{.ID}}"]
source = '''
  if exists(.items) && is_array(.items) {. = unnest!(.items)} else {.}
'''

[transforms.{{.ID}}_items]
type = "remap"
inputs = ["{{.ID}}_split"]
source = '''
  if exists(.items) {. = .items} else {.}
'''
{{end}}
`
}
