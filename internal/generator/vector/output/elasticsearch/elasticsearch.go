package elasticsearch

import (
	"strings"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	. "github.com/openshift/cluster-logging-operator/internal/generator"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/security"
	corev1 "k8s.io/api/core/v1"
)

type Elasticsearch struct {
	ID_Key      string
	Desc        string
	ComponentID string
	Inputs      string
	Index       string
	Endpoint    string
}

func (e Elasticsearch) Name() string {
	return "elasticsearchTemplate"
}

func (e Elasticsearch) Template() string {
	return `{{define "` + e.Name() + `" -}}

[sinks.{{.ComponentID}}]
type = "elasticsearch"
inputs = {{.Inputs}}
endpoint = "{{.Endpoint}}"
index = "{{ "{{ log_type }}-write" }}"
request.timeout_secs = 2147483648
bulk_action = "create"
{{- end}}
`
}

func Conf(o logging.OutputSpec, inputs []string, secret *corev1.Secret, op Options) []Element {
	outputs := []Element{}

	outputs = MergeElements(outputs,
		[]Element{
			Output(o, inputs, secret, op),
		},
		TLSConf(o, secret),
		BasicAuth(o, secret),
	)

	return outputs
}

func Output(o logging.OutputSpec, inputs []string, secret *corev1.Secret, op Options) Element {

	return Elasticsearch{
		ComponentID: strings.ToLower(vectorhelpers.Replacer.Replace(o.Name)),
		Endpoint:    o.URL,
		Inputs:      vectorhelpers.MakeInputs(inputs...),
	}
}

func TLSConf(o logging.OutputSpec, secret *corev1.Secret) []Element {
	conf := []Element{}
	if o.Secret != nil {
		hasTLS := false
		conf = append(conf, security.TLSConf{
			Desc:        "TLS Config",
			ComponentID: strings.ToLower(vectorhelpers.Replacer.Replace(o.Name)),
		})

		if o.Name == logging.OutputNameDefault || security.HasTLSCertAndKey(secret) {
			hasTLS = true
			kc := TLSKeyCert{
				CertPath: security.SecretPath(o.Secret.Name, constants.ClientCertKey),
				KeyPath:  security.SecretPath(o.Secret.Name, constants.ClientPrivateKey),
			}
			conf = append(conf, kc)
		}
		if o.Name == logging.OutputNameDefault || security.HasCABundle(secret) {
			hasTLS = true
			ca := CAFile{
				CAFilePath: security.SecretPath(o.Secret.Name, constants.TrustedCABundleKey),
			}
			conf = append(conf, ca)
		}
		if !hasTLS {
			return []Element{}
		}
	}
	return conf
}

func BasicAuth(o logging.OutputSpec, secret *corev1.Secret) []Element {
	conf := []Element{}

	if o.Secret != nil {
		hasBasicAuth := false
		conf = append(conf, BasicAuthConf{
			Desc:        "Basic Auth Config",
			ComponentID: strings.ToLower(vectorhelpers.Replacer.Replace(o.Name)),
		})
		if security.HasUsernamePassword(secret) {
			hasBasicAuth = true
			up := UserNamePass{
				UsernamePath: security.SecretPath(o.Secret.Name, constants.ClientUsername),
				PasswordPath: security.SecretPath(o.Secret.Name, constants.ClientPassword),
			}
			conf = append(conf, up)
		}
		if !hasBasicAuth {
			return []Element{}
		}
	}

	return conf
}
