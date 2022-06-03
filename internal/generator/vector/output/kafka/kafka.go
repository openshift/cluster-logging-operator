package kafka

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/ViaQ/logerr/v2/log"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	. "github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/helpers"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	urlhelper "github.com/openshift/cluster-logging-operator/internal/generator/url"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/security"
	corev1 "k8s.io/api/core/v1"
)

const (
	defaultKafkaTopic = "topic"
)

type Kafka struct {
	Desc             string
	ComponentID      string
	Inputs           string
	BootstrapServers string
	Topic            string
}

func (k Kafka) Name() string {
	return "kafkaTemplate"
}

func (k Kafka) Template() string {
	return `{{define "` + k.Name() + `" -}}
# {{.Desc}}
[sinks.{{.ComponentID}}]
type = "kafka"
inputs = {{.Inputs}}
bootstrap_servers = {{.BootstrapServers}}
topic = {{.Topic}}
{{end}}
`
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
			Encoding(o, op),
		},
		TLSConf(o, secret),
		SASLConf(o, secret),
	)
}

func Output(o logging.OutputSpec, inputs []string, secret *corev1.Secret, op Options) Element {
	if genhelper.IsDebugOutput(op) {
		return genhelper.DebugOutput
	}
	return Kafka{
		Desc:             "Kafka config",
		ComponentID:      strings.ToLower(helpers.Replacer.Replace(o.Name)),
		Inputs:           vectorhelpers.MakeInputs(inputs...),
		Topic:            fmt.Sprintf("%q", Topics(o)),
		BootstrapServers: fmt.Sprintf("%q", Brokers(o)),
	}
}

//Brokers returns the list of broker endpoints of a kafka cluster.
//The list represents only the initial set used by the collector's kafka client for the
//first connention only. The collector's kafka client fetches constantly an updated list
//from kafka. These updates are not reconciled back to the collector configuration.
//The list of brokers are populated from the Kafka OutputSpec `Brokers` field, a list of
//valid URLs. If none provided the target URL from the OutputSpec is used as fallback.
//Finally, if neither approach works the current collector process will be terminated.
func Brokers(o logging.OutputSpec) string {
	brokers := []string{o.URL} // Put o.URL first in the list.
	if o.Kafka != nil {        // Add optional extra broker URLs.
		brokers = append(brokers, o.Kafka.Brokers...)
	}
	for i, s := range brokers { // Convert URLs to hostnames
		// FIXME URL parse error is being ignored.
		u, _ := url.Parse(s)
		brokers[i] = u.Host
	}
	return strings.Join(brokers, ",")
}

//Topic returns the name of an existing kafka topic.
//The kafka topic is either extracted from the kafka OutputSpec `Topic` field in a multiple broker
//setup or as a fallback from the OutputSpec URL if provided as a host path. Defaults to `topic`.
func Topics(o logging.OutputSpec) string {
	if o.Kafka != nil && o.Kafka.Topic != "" {
		return o.Kafka.Topic
	}

	url, _ := urlhelper.Parse(o.URL)
	topic := strings.TrimLeft(url.Path, "/")
	if topic != "" {
		return topic
	}

	// Fallback to default topic
	return defaultKafkaTopic
}

func Encoding(o logging.OutputSpec, op Options) Element {
	return ConfLiteral{
		ComponentID:  strings.ToLower(helpers.Replacer.Replace(o.Name)),
		TemplateName: "kafkaEncoding",
		TemplateStr: `
{{define "kafkaEncoding" -}}
[sinks.{{.ComponentID}}.encoding]
codec = "json"
timestamp_format = "rfc3339"
{{end}}
			`,
	}
}

func TLSConf(o logging.OutputSpec, secret *corev1.Secret) []Element {
	conf := []Element{}
	// FIXME ignored URL parse error.
	u, _ := url.Parse(o.URL)
	if !urlhelper.IsTLSScheme(u.Scheme) { // Not a TLS URL
		// FIXME error if not TLS but has secret, probable misconfiguration.
		return conf
	}
	conf = append(conf, security.TLSConf{
		Desc:        "TLS Config",
		ComponentID: strings.ToLower(helpers.Replacer.Replace(o.Name)),
	})
	if o.Secret != nil {
		if security.HasPassphrase(secret) {
			pp := Passphrase{
				PassphrasePath: security.SecretPath(o.Secret.Name, constants.Passphrase),
			}
			conf = append(conf, pp)
		}
		if security.HasTLSCertAndKey(secret) {
			kc := TLSKeyCert{
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
		if security.HasKeys(secret, constants.TLSInsecure) {
			conf = append(conf, TLSInsecure(true))
			log.NewLogger("").Info("Insecure TLS selected for output %q", o.Name)
		}
	}
	conf = append(conf, TLS(true))
	return conf
}

func SASLConf(o logging.OutputSpec, secret *corev1.Secret) []Element {
	conf := []Element{}
	// Try the preferred and deprecated names.
	_, ok := security.TryKeys(secret, constants.SASLEnable, constants.DeprecatedSaslOverSSL)
	if o.Secret != nil && ok {
		hasSASL := false
		conf = append(conf, SaslConf{
			Desc:        "Sasl Config",
			ComponentID: strings.ToLower(helpers.Replacer.Replace(o.Name)),
		})
		if security.HasUsernamePassword(secret) {
			hasSASL = true
			up := UserNamePass{
				Username: security.GetFromSecret(secret, constants.ClientUsername),
				Password: security.GetFromSecret(secret, constants.ClientPassword),
			}
			conf = append(conf, up)
		}
		conf = append(conf, Sasl(true))
		if !hasSASL {
			return []Element{}
		}
	}
	return conf
}
