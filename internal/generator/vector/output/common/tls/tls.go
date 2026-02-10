package tls

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	typehelpers "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/url"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

const (
	Component                 = "component"
	IncludeEnabled            = "IncludeEnabled"
	ExcludeInsecureSkipVerify = "ExcludeInsecureSkipVerify"
)

var (
	IncludeEnabledOption = framework.Option{Name: IncludeEnabled, Value: ""}
)

type TLSConf struct {
	Component string
	ID        string
	// Enabled add enabled config (not required for all conf)
	Enabled            typehelpers.OptionalPair
	NeedsEnabled       bool
	InsecureSkipVerify bool
	TlsMinVersion      string
	CipherSuites       string
	CAFilePath         string
	CertPath           string
	KeyPath            string
	PassPhrase         string
}

func NewTls(endpoint string, spec *obs.OutputTLSSpec, secrets observability.Secrets, op utils.Options, options ...framework.Option) (conf *api.TLS) {
	if !url.IsSecure(endpoint) {
		return
	}
	conf = &api.TLS{}
	if _, found := framework.HasOption(IncludeEnabled, options); found && spec != nil {
		conf.Enabled = true
	}
	if spec != nil {
		conf.CAFile = ValuePath(spec.CA, "%s")
		conf.CRTFile = ValuePath(spec.Certificate, "%s")
		conf.KeyFile = SecretPath(spec.Key, "%s")
		conf.KeyPass = secrets.AsString(spec.KeyPassphrase)
		if _, found := framework.HasOption(ExcludeInsecureSkipVerify, options); !found && spec.InsecureSkipVerify {
			conf.VerifyCertificate = false
			conf.VerifyHostname = false
		}
	}
	conf.SetTLSProfile(op)
	return conf
}

func New(id string, spec *obs.OutputTLSSpec, secrets observability.Secrets, op framework.Options, options ...framework.Option) framework.Element {
	if outURL, found := framework.HasOption(framework.URL, options); found {
		if !url.IsSecure(outURL.(string)) {
			return framework.Nil
		}
	}
	conf := TLSConf{
		Component: "sinks",
		ID:        id,
	}
	if comp, found := framework.HasOption(Component, options); found {
		conf.Component = comp.(string)
	}
	if _, found := framework.HasOption(IncludeEnabled, options); found && spec != nil {
		conf.Enabled = typehelpers.NewOptionalPair("enabled", true)
	}

	if spec != nil {
		conf.CAFilePath = ValuePath(spec.CA)
		conf.CertPath = ValuePath(spec.Certificate)
		conf.KeyPath = SecretPath(spec.Key)
		conf.PassPhrase = secrets.AsString(spec.KeyPassphrase)
		if _, found := framework.HasOption(ExcludeInsecureSkipVerify, options); found {
			conf.InsecureSkipVerify = false
		} else {
			conf.InsecureSkipVerify = spec.InsecureSkipVerify
		}
	}
	setTLSProfileFromOptions(&conf, op)
	if conf.CipherSuites != "" || conf.TlsMinVersion != "" || spec != nil {
		conf.NeedsEnabled = true
	}
	return conf
}

func ValuePath(resource *obs.ValueReference, formatter ...string) string {
	if resource == nil {
		return ""
	}
	if resource.SecretName != "" {
		return helpers.SecretPath(resource.SecretName, resource.Key, formatter...)
	} else if resource.ConfigMapName != "" {
		return helpers.ConfigPath(resource.ConfigMapName, resource.Key, formatter...)
	}
	return ""
}

func SecretPath(resource *obs.SecretReference, formatter ...string) string {
	if resource == nil || resource.SecretName == "" {
		return ""
	}
	return helpers.SecretPath(resource.SecretName, resource.Key, formatter...)
}

func setTLSProfileFromOptions(t *TLSConf, op framework.Options) {
	if version, found := op[framework.MinTLSVersion]; found {
		t.TlsMinVersion = version.(string)
	}
	if ciphers, found := op[framework.Ciphers]; found {
		t.CipherSuites = ciphers.(string)
	}
}

func (t TLSConf) Name() string {
	return "vectorTLS"
}

func (t TLSConf) Template() string {
	if !t.NeedsEnabled {
		return `{{define "vectorTLS" -}}{{end}}`
	}
	return `
{{define "vectorTLS" -}}
[{{.Component}}.{{.ID}}.tls]
{{ .Enabled }}
{{- if ne .TlsMinVersion "" }}
min_tls_version = "{{ .TlsMinVersion }}"
{{- end }}
{{- if ne .CipherSuites "" }}
ciphersuites = "{{ .CipherSuites }}"
{{- end }}
{{- if .InsecureSkipVerify }}
verify_certificate = false
verify_hostname = false
{{- end }}
{{- if and .KeyPath .CertPath }}
key_file = {{ .KeyPath }}
crt_file = {{ .CertPath }}
{{- end }}
{{- if .CAFilePath }}
ca_file = {{ .CAFilePath }}
{{- end }}
{{- if .PassPhrase }}
key_pass = "{{ .PassPhrase }}"
{{- end }}
{{ end }}`
}
