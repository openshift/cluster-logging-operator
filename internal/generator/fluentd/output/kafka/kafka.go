package kafka

import (
	"net/url"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	corev1 "k8s.io/api/core/v1"

	. "github.com/openshift/cluster-logging-operator/internal/generator"
	. "github.com/openshift/cluster-logging-operator/internal/generator/fluentd/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/normalize"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	urlhelper "github.com/openshift/cluster-logging-operator/internal/generator/url"
)

const (
	defaultKafkaTopic = "topic"
)

type Kafka struct {
	Desc           string
	StoreID        string
	Brokers        string
	Topics         string
	SecurityConfig []Element
	BufferConfig   []Element
}

func (k Kafka) Name() string {
	return "kafkaTemplate"
}

func (k Kafka) Template() string {
	return `{{define "` + k.Name() + `" -}}
@type kafka2
@id {{.StoreID}}
brokers {{.Brokers}}
default_topic {{.Topics}}
use_event_time true
{{- with $x := compose .SecurityConfig }}
{{$x}}
{{- end}}
<format>
  @type json
</format>
{{compose .BufferConfig}}
{{- end}}
`
}

func Conf(bufspec *logging.FluentdBufferSpec, secret *corev1.Secret, o logging.OutputSpec, op Options) []Element {
	return []Element{
		FromLabel{
			InLabel: helpers.LabelName(o.Name),
			SubElements: []Element{
				normalize.DedotLabels(),
				Output(bufspec, secret, o, op),
			},
		},
	}
}

func Output(bufspec *logging.FluentdBufferSpec, secret *corev1.Secret, o logging.OutputSpec, op Options) Element {
	if genhelper.IsDebugOutput(op) {
		return genhelper.DebugOutput
	}

	topics := Topics(o)
	storeID := helpers.StoreID("", o.Name, "")
	return Match{
		MatchTags: "**",
		MatchElement: Kafka{
			StoreID:        strings.ToLower(helpers.Replacer.Replace(o.Name)),
			Topics:         topics,
			Brokers:        Brokers(o),
			SecurityConfig: SecurityConfig(o, secret),
			BufferConfig:   output.Buffer([]string{"_" + topics}, bufspec, storeID, &o),
		},
	}
}

// Brokers returns the list of broker endpoints of a kafka cluster.
// The list represents only the initial set used by the collector's kafka client for the
// first connention only. The collector's kafka client fetches constantly an updated list
// from kafka. These updates are not reconciled back to the collector configuration.
// The list of brokers are populated from the Kafka OutputSpec `Brokers` field, a list of
// valid URLs. If none provided the target URL from the OutputSpec is used as fallback.
// Finally, if neither approach works the current collector process will be terminated.
func Brokers(o logging.OutputSpec) string {
	parseBroker := func(b string) string {
		url, _ := url.Parse(b)
		return url.Host
	}

	if o.Kafka != nil {
		if o.Kafka.Brokers != nil {
			brokers := []string{}
			for _, broker := range o.Kafka.Brokers {
				b := parseBroker(broker)
				if b != "" {
					brokers = append(brokers, b)
				}
			}

			if len(brokers) > 0 {
				return strings.Join(brokers, ",")
			}
		}
	}

	// Fallback to parse a single broker from target's URL
	fallback := parseBroker(o.URL)
	if fallback == "" {
		panic("Failed to parse any Kafka broker from output spec")
	}

	return fallback
}

// Topic returns the name of an existing kafka topic.
// The kafka topic is either extracted from the kafka OutputSpec `Topic` field in a multiple broker
// setup or as a fallback from the OutputSpec URL if provided as a host path. Defaults to `topic`.
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

func SecurityConfig(o logging.OutputSpec, secret *corev1.Secret) []Element {
	conf := []Element{}
	if o.Secret != nil {
		if security.HasUsernamePassword(secret) {
			up := UserNamePass{
				UsernamePath: security.SecretPath(o.Secret.Name, constants.ClientUsername),
				PasswordPath: security.SecretPath(o.Secret.Name, constants.ClientPassword),
			}
			conf = append(conf, up)
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
		// Try the preferred and deprecated names.
		_, ok := security.TryKeys(secret, constants.SASLEnable, constants.DeprecatedSaslOverSSL)
		sasl := SASL{
			SaslOverSSL: ok,
		}
		if security.HasPassphrase(secret) {
			sasl.SaslKeyPassword = security.SecretPath(o.Secret.Name, constants.Passphrase)
		}
		if scramMechanism := security.GetFromSecret(secret, constants.SASLMechanisms); scramMechanism != "" {
			sasl.ScramMechanism = scramMechanism
		}
		conf = append(conf, sasl)
	}
	return conf
}
