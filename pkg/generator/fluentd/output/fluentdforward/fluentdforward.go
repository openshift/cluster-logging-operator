package fluentdforward

import (
	"strings"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	. "github.com/openshift/cluster-logging-operator/pkg/generator"
	. "github.com/openshift/cluster-logging-operator/pkg/generator/fluentd/elements"
	"github.com/openshift/cluster-logging-operator/pkg/generator/fluentd/helpers"
	"github.com/openshift/cluster-logging-operator/pkg/generator/fluentd/output"
	"github.com/openshift/cluster-logging-operator/pkg/generator/fluentd/output/security"
	genhelper "github.com/openshift/cluster-logging-operator/pkg/generator/helpers"
	"github.com/openshift/cluster-logging-operator/pkg/generator/url"
	urlhelper "github.com/openshift/cluster-logging-operator/pkg/generator/url"
	corev1 "k8s.io/api/core/v1"
)

const (
	defaultFluentdForwardPort = "24224"
)

type FluentdForward struct {
	Desc           string
	StoreID        string
	Host           string
	Port           string
	BufferConfig   []Element
	SecurityConfig []Element
}

func (ff FluentdForward) Name() string {
	return "fluentdForwardTemplate"
}

func (ff FluentdForward) Template() string {
	return `{{define "` + ff.Name() + `" -}}
{{if .Desc -}}
# {{.Desc}}
{{end -}}
@type forward
@id {{.StoreID}}
<server>
  host {{.Host}}
  port {{.Port}}
</server>
heartbeat_type none
keepalive true
{{compose .SecurityConfig}}
{{compose .BufferConfig}}
{{- end}}
`
}

func Conf(bufspec *logging.FluentdBufferSpec, secret *corev1.Secret, o logging.OutputSpec, op Options) []Element {
	return []Element{
		FromLabel{
			InLabel: helpers.LabelName(o.Name),
			SubElements: []Element{
				Output(bufspec, secret, o, op),
			},
		},
	}
}

func Output(bufspec *logging.FluentdBufferSpec, secret *corev1.Secret, o logging.OutputSpec, op Options) Element {
	if genhelper.IsDebugOutput(op) {
		return genhelper.DebugOutput
	}
	// URL is parasable, checked at input sanitization
	u, _ := urlhelper.Parse(o.URL)
	port := u.Port()
	if port == "" {
		port = defaultFluentdForwardPort
	}
	storeID := strings.ToLower(helpers.Replacer.Replace(o.Name))
	return Match{
		MatchTags: "**",
		MatchElement: FluentdForward{
			StoreID:        storeID,
			Host:           u.Hostname(),
			Port:           port,
			SecurityConfig: SecurityConfig(o, secret),
			BufferConfig:   output.Buffer(output.NOKEYS, bufspec, storeID, &o),
		},
	}
}

func SecurityConfig(o logging.OutputSpec, secret *corev1.Secret) []Element {
	// URL is parasable, checked at input sanitization
	u, _ := urlhelper.Parse(o.URL)
	conf := []Element{}
	if url.IsTLSScheme(u.Scheme) {
		conf = []Element{
			TLS(true),
		}
		if secret == nil {
			conf = append(conf, TLS(false))
		}
	}
	if o.Secret != nil {
		if security.HasSharedKey(secret) {
			sk := SharedKey{
				Key: security.GetFromSecret(secret, constants.SharedKey),
			}
			conf = append(conf, sk)
		}
	}
	if url.IsTLSScheme(u.Scheme) {
		if security.HasTLSCertAndKey(secret) {
			kc := TLSCertKey{
				CertPath: security.SecretPath(o.Secret.Name, constants.ClientCertKey),
				KeyPath:  security.SecretPath(o.Secret.Name, constants.ClientPrivateKey),
			}
			conf = append(conf, kc)
		}
		if security.HasCABundle(secret) {
			ca := CAFile{
				CAFilePath: security.SecretPath(o.Secret.Name, constants.TrustedCABundleKey),
			}
			conf = append(conf, ca)
		}
		if security.HasPassphrase(secret) {
			p := Passphrase{
				PassphrasePath: security.SecretPath(o.Secret.Name, constants.Passphrase),
			}
			conf = append(conf, p)
		}
	}

	return conf
}
