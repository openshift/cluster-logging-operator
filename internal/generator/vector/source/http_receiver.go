package source

import (
	"strings"

	configv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/tls"
)

func NewHttpSource(id, inputName string, input logging.InputSpec, op framework.Options) (framework.Element, string) {
	var minTlsVersion, cipherSuites string
	if _, ok := op[framework.ClusterTLSProfileSpec]; ok {
		tlsProfileSpec := op[framework.ClusterTLSProfileSpec].(configv1.TLSProfileSpec)
		minTlsVersion = tls.MinTLSVersion(tlsProfileSpec)
		cipherSuites = strings.Join(tls.TLSCiphers(tlsProfileSpec), `,`)
	}
	return HttpReceiver{
		ID:            id,
		InputName:     inputName,
		ListenAddress: helpers.ListenOnAllLocalInterfacesAddress(),
		ListenPort:    input.Receiver.GetHTTPPort(),
		Format:        input.Receiver.GetHTTPFormat(),
		TlsMinVersion: minTlsVersion,
		CipherSuites:  cipherSuites,
	}, helpers.MakeID(id, "items")
}

type HttpReceiver struct {
	ID            string
	InputName     string
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
key_file = "/etc/collector/receiver/{{.InputName}}/tls.key"
crt_file = "/etc/collector/receiver/{{.InputName}}/tls.crt"
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
