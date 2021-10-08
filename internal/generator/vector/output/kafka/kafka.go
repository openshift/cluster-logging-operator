package kafka

import (
	"net/url"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	corev1 "k8s.io/api/core/v1"

	. "github.com/openshift/cluster-logging-operator/internal/generator"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	urlhelper "github.com/openshift/cluster-logging-operator/internal/generator/url"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/security"
)

const (
	defaultKafkaTopic = "topic"
)

type Kafka struct {
	Desc           string
	ElementID      string
	StoreID        string
	Brokers        string
	Topics         string
	InputPipeline  []string
	SecurityConfig []Element
	BufferConfig   []Element
}

func (k Kafka) ID() string {
	return k.ElementID
}

func (k Kafka) Name() string {
	return k.ElementID
}

func (k Kafka) Template() string {
	return `{{define "` + k.Name() + `" -}}
[sinks.{{.ElementID}}]
  type = "kafka"
  input = ` + helpers.ConcatArrays(k.InputPipeline) + `
  bootstrap_servers = "{{.Brokers}}"
  topic = "{{.Topics}}"
  {{- with $x := compose .SecurityConfig }}
  {{$x}}
  {{- end}}
  {{- end}}
`
}

func Conf(bufspec *logging.FluentdBufferSpec, secret *corev1.Secret, o logging.OutputSpec, op Options, inputPipelines []string) []Element {
	return []Element{
		Output(bufspec, secret, o, op, inputPipelines),
	}
}

func Output(bufspec *logging.FluentdBufferSpec, secret *corev1.Secret, o logging.OutputSpec, op Options, inputPipelines []string) Element {
	if genhelper.IsDebugOutput(op) {
		return genhelper.DebugOutput
	}
	topics := Topics(o)
	storeID := genhelper.StoreID("", o.Name, "")
	return Kafka{
		StoreID:        strings.ToLower(genhelper.Replacer.Replace(o.Name)),
		ElementID:      storeID,
		Topics:         topics,
		InputPipeline:  inputPipelines,
		Brokers:        Brokers(o),
		SecurityConfig: SecurityConfig(o, secret),
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

// TODO: Update UserNamePass with file path
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
		conf = append(conf, SaslOverSSL(ok))
	}
	return conf
}
