package fluentdforward

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers/security"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/normalize"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/url"
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
	BufferConfig   []framework.Element
	SecurityConfig []framework.Element
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
keepalive_timeout 30s
expire_dns_cache 30s
{{compose .SecurityConfig}}
{{compose .BufferConfig}}
{{- end}}
`
}

func Conf(bufspec *logging.FluentdBufferSpec, secret *corev1.Secret, o logging.OutputSpec, op framework.Options) []framework.Element {
	return []framework.Element{
		elements.FromLabel{
			InLabel: helpers.LabelName(o.Name),
			SubElements: []framework.Element{
				normalize.DedotLabels(),
				Output(bufspec, secret, o, op),
			},
		},
	}
}

func Output(bufspec *logging.FluentdBufferSpec, secret *corev1.Secret, o logging.OutputSpec, op framework.Options) framework.Element {
	if genhelper.IsDebugOutput(op) {
		return genhelper.DebugOutput
	}
	// URL is parasable, checked at input sanitization
	u, _ := url.Parse(o.URL)
	port := u.Port()
	if port == "" {
		port = defaultFluentdForwardPort
	}
	storeID := strings.ToLower(helpers.Replacer.Replace(o.Name))
	return elements.Match{
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

func SecurityConfig(o logging.OutputSpec, secret *corev1.Secret) []framework.Element {
	// URL is parasable, checked at input sanitization
	u, _ := url.Parse(o.URL)
	conf := []framework.Element{}
	if url.IsTLSScheme(u.Scheme) {
		conf = []framework.Element{
			TLS{
				InsecureMode: o.TLS != nil && o.TLS.InsecureSkipVerify,
			},
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
