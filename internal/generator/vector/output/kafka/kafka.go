package kafka

import (
	"fmt"
	"net/url"
	"strings"

	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	urlhelper "github.com/openshift/cluster-logging-operator/internal/generator/url"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/normalize"
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
	common.RootMixin
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
{{.Compression}}
{{end}}
`
}

func (k *Kafka) SetCompression(algo string) {
	k.Compression.Value = algo
}

func New(id string, o logging.OutputSpec, inputs []string, secret *corev1.Secret, strategy common.ConfigStrategy, op Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Debug(id, vectorhelpers.MakeInputs(inputs...)),
		}
	}

	dedottedID := vectorhelpers.MakeID(id, "dedot")
	brokers, genTlsConf := Brokers(o)
	sink := Output(id, o, []string{dedottedID}, secret, op, brokers)
	if strategy != nil {
		strategy.VisitSink(sink)
	}
	return MergeElements(
		[]Element{
			normalize.DedotLabels(dedottedID, inputs),
			sink,
			Encoding(id, op),
			common.NewAcknowledgments(id, strategy),
			common.NewBatch(id, strategy),
			common.NewBuffer(id, strategy),
		},
		TLSConf(id, o, secret, op, genTlsConf),
		SASLConf(id, o, secret),
	)
}

func Output(id string, o logging.OutputSpec, inputs []string, secret *corev1.Secret, op Options, brokers string) *Kafka {
	return &Kafka{
		Desc:             "Kafka config",
		ComponentID:      id,
		Inputs:           vectorhelpers.MakeInputs(inputs...),
		Topic:            fmt.Sprintf("%q", Topics(o)),
		BootstrapServers: fmt.Sprintf("%q", brokers),
		RootMixin:        common.NewRootMixin(nil),
	}
}

// Brokers returns the list of broker endpoints of a Kafka cluster.
// The list represents only the initial set used by the collector's Kafka client for the
// first connection only. The collector's Kafka client fetches constantly an updated list
// from Kafka. These updates are not reconciled back to the collector configuration.
// The list of brokers are populated from the Kafka OutputSpec `Brokers` field, a list of
// valid URLs. If none provided the target URL from the OutputSpec is used as fallback.
// Finally, if neither approach works the current collector process will be terminated.
func Brokers(o logging.OutputSpec) (string, bool) {
	genTLSConf := false // is there a TLS endpoint among the brokers
	brokerUrls := []string{}
	if o.URL != "" {
		brokerUrls = append(brokerUrls, o.URL)
	}
	if o.Kafka != nil { // Add optional extra broker URLs.
		brokerUrls = append(brokerUrls, o.Kafka.Brokers...)
	}
	brokerHosts := []string{}
	for _, s := range brokerUrls { // Convert URLs to hostnames
		u, _ := url.Parse(s)
		if u != nil {
			if !genTLSConf {
				genTLSConf = urlhelper.IsTLSScheme(u.Scheme)
			}
			brokerHosts = append(brokerHosts, u.Host)
		}
	}
	return strings.Join(brokerHosts, ","), genTLSConf
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

func Encoding(id string, op Options) Element {
	return ConfLiteral{
		ComponentID:  id,
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

func TLSConf(id string, o logging.OutputSpec, secret *corev1.Secret, op Options, genTLSConf bool) []Element {
	isTls := isTlsBrokers(o)
	if isTls && (o.Secret != nil || (o.TLS != nil && o.TLS.InsecureSkipVerify)) {
		conf := []Element{}

		if tlsConf := common.GenerateTLSConfWithID(id, o, secret, op, genTLSConf); tlsConf != nil {
			// KafkaInsecure (InsecureTLS)
			if o.TLS != nil && o.TLS.InsecureSkipVerify {
				conf = append(conf, InsecureTLS{
					ComponentID: id,
				})
			}

			tlsConf.InsecureSkipVerify = false
			conf = append(conf, tlsConf)
		}
		// Kafka does not use the verify_certificate or verify_hostname options, see insecureTLS
		return conf
	}
	return []Element{}
}

func isTlsBrokers(o logging.OutputSpec) bool {
	isTls := true
	if o.Kafka != nil {
		for _, b := range o.Kafka.Brokers {
			if !strings.HasPrefix(b, "tls:") {
				isTls = false
				break
			}
		}
	}
	return isTls
}

func SASLConf(id string, o logging.OutputSpec, secret *corev1.Secret) []Element {
	conf := []Element{}
	// Try the preferred and deprecated names.
	_, ok := common.TryKeys(secret, constants.SASLEnable, constants.DeprecatedSaslOverSSL)
	if o.Secret != nil && ok {
		if common.HasUsernamePassword(secret) {
			sasl := SASL{
				Desc:        "SASL Config",
				ComponentID: id,
				Username:    common.GetFromSecret(secret, constants.ClientUsername),
				Password:    common.GetFromSecret(secret, constants.ClientPassword),
				Mechanism:   SASLMechanismPlain,
			}
			if m := common.GetFromSecret(secret, constants.SASLMechanisms); m != "" {
				// librdkafka does not support multiple values for sasl mechanism
				sasl.Mechanism = m
			}
			conf = append(conf, sasl)
		}
	}
	return conf
}
