package elasticsearch

import (
	"fmt"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	corev1 "k8s.io/api/core/v1"

	. "github.com/openshift/cluster-logging-operator/internal/generator"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/url"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/security"
)

const (
	defaultElasticsearchPort = "9200"
	KeyStructured            = "structured"
)

type Elasticsearch struct {
	Desc           string
	StoreID        string
	Host           string
	Port           string
	RetryTag       Element
	SecurityConfig []Element
	InputPipeline  []string
}

func (e Elasticsearch) Name() string {
	return "elasticsearchTemplate"
}

func (e Elasticsearch) Template() string {
	return `{{define "` + e.Name() + `" -}}
[sinks.{{.StoreID}}]
  type = "elasticsearch"
  inputs = ` + helpers.ConcatArrays(e.InputPipeline) + `
  endpoint = "{{.Host}}:{{.Port}}"
  mode = "normal"
  pipeline = "pipeline-name"
  compression = "none"
  {{- with $x := compose .SecurityConfig }}
  {{$x}}
  {{- end}}
  {{- end}}
`
}

func Conf(secret *corev1.Secret, o logging.OutputSpec, op Options, inputPipelines []string) []Element {
	return []Element{
		Output(secret, o, op, inputPipelines),
	}
}

func Output(secret *corev1.Secret, o logging.OutputSpec, op Options, inputPipelines []string) Elasticsearch {
	// URL is parasable, checked at input sanitization
	u, _ := url.Parse(o.URL)
	port := u.Port()
	if port == "" {
		port = defaultElasticsearchPort
	}
	storeID := genhelper.StoreID("", o.Name, "")
	return Elasticsearch{
		StoreID:        storeID,
		Host:           u.Hostname(),
		Port:           port,
		InputPipeline:  inputPipelines,
		SecurityConfig: SecurityConfig(o, secret),
	}
}

func SecurityConfig(o logging.OutputSpec, secret *corev1.Secret) []Element {
	// URL is parasable, checked at input sanitization
	conf := []Element{}
	if o.Secret != nil {
		if security.HasUsernamePassword(secret) {
			up := UserNamePass{
				UsernamePath: security.SecretPath(o.Secret.Name, constants.ClientUsername),
				PasswordPath: security.SecretPath(o.Secret.Name, constants.ClientPassword),
			}
			conf = append(conf, up)
		}
		if o.Name == logging.OutputNameDefault || security.HasTLSCertAndKey(secret) {
			kc := TLSKeyCert{
				CertPath: security.SecretPath(o.Secret.Name, constants.ClientCertKey),
				KeyPath:  security.SecretPath(o.Secret.Name, constants.ClientPrivateKey),
			}
			conf = append(conf, kc)
		}
		if o.Name == logging.OutputNameDefault || security.HasCABundle(secret) {
			ca := CAFile{
				CAFilePath: security.SecretPath(o.Secret.Name, constants.TrustedCABundleKey),
			}
			conf = append(conf, ca)
		}
	}
	return conf
}

func generateRubyDigArgs(path string) string {
	var args []string
	for _, s := range strings.Split(path, ".") {
		args = append(args, fmt.Sprintf("%q", s))
	}
	return strings.Join(args, ",")
}
