package splunk

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/helpers"
	urlhelper "github.com/openshift/cluster-logging-operator/internal/generator/url"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/security"
	"net/url"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/internal/generator"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	corev1 "k8s.io/api/core/v1"
)

var (
	splunkEncodingJson = fmt.Sprintf("%q", "json")
)

type Splunk struct {
	ComponentID  string
	Inputs       string
	Endpoint     string
	DefaultToken string
}

func (s Splunk) Name() string {
	return "SplunkVectorTemplate"
}

func (s Splunk) Template() string {
	return `{{define "` + s.Name() + `" -}}
[sinks.{{.ComponentID}}]
type = "splunk_hec"
inputs = {{.Inputs}}
endpoint = "{{.Endpoint}}"
compression = "none"
default_token = "{{.DefaultToken}}"
{{end}}`
}

type SplunkEncoding struct {
	ComponentID string
	Codec       string
}

func (se SplunkEncoding) Name() string {
	return "splunkEncoding"
}

func (se SplunkEncoding) Template() string {
	return `{{define "` + se.Name() + `" -}}
[sinks.{{.ComponentID}}.encoding]
codec = {{.Codec}}
{{end}}`
}

func Conf(o logging.OutputSpec, inputs []string, secret *corev1.Secret, op Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Debug(strings.ToLower(vectorhelpers.Replacer.Replace(o.Name)), vectorhelpers.MakeInputs(inputs...)),
		}
	}
	return MergeElements(
		[]Element{
			Output(o, inputs, secret, op),
			Encoding(o),
		},
		TLSConf(o, secret),
	)
}

func Output(o logging.OutputSpec, inputs []string, secret *corev1.Secret, op Options) Element {
	return Splunk{
		ComponentID:  strings.ToLower(vectorhelpers.Replacer.Replace(o.Name)),
		Inputs:       vectorhelpers.MakeInputs(inputs...),
		Endpoint:     o.URL,
		DefaultToken: security.GetFromSecret(secret, "hecToken"),
	}
}

func Encoding(o logging.OutputSpec) Element {
	return SplunkEncoding{
		ComponentID: strings.ToLower(vectorhelpers.Replacer.Replace(o.Name)),
		Codec:       splunkEncodingJson,
	}
}

func TLSConf(o logging.OutputSpec, secret *corev1.Secret) []Element {
	conf := []Element{}
	if o.Secret == nil {
		return conf
	}
	hasTLS := false
	u, _ := url.Parse(o.URL)
	if urlhelper.IsTLSScheme(u.Scheme) {
		if security.HasPassphrase(secret) {
			pp := security.Passphrase{
				PassphrasePath: security.SecretPath(o.Secret.Name, constants.Passphrase),
			}
			conf = append(conf, pp)
			hasTLS = true
		}
		if security.HasTLSCertAndKey(secret) {
			kc := TLSKeyCert{
				CertPath: security.SecretPath(o.Secret.Name, constants.ClientCertKey),
				KeyPath:  security.SecretPath(o.Secret.Name, constants.ClientPrivateKey),
			}
			conf = append(conf, kc)
			hasTLS = true
		}
		if security.HasCABundle(secret) {
			ca := CAFile{
				CAFilePath: security.SecretPath(o.Secret.Name, constants.TrustedCABundleKey),
			}
			conf = append(conf, ca)
			hasTLS = true
		}
	}
	if hasTLS {
		conf = append([]Element{security.TLSConf{
			ComponentID:        strings.ToLower(helpers.Replacer.Replace(o.Name)),
			InsecureSkipVerify: o.TLS != nil && o.TLS.InsecureSkipVerify,
		}}, conf...)
	}
	return conf
}
