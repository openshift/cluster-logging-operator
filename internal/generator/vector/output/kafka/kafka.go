package kafka

import (
	"fmt"
	"net/url"
	"strings"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/adapters"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	urlhelper "github.com/openshift/cluster-logging-operator/internal/generator/url"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sinks"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	commontemplate "github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/template"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

const (
	defaultKafkaTopic  = "topic"
	SASLMechanismPlain = "PLAIN"
)

func New(id string, o *adapters.Output, inputs []string, secrets observability.Secrets, op utils.Options) []framework.Element {
	if genhelper.IsDebugOutput(op) {
		return []framework.Element{
			elements.Debug(id, vectorhelpers.MakeInputs(inputs...)),
		}
	}
	componentID := vectorhelpers.MakeID(id, "topic")
	elements := []framework.Element{
		commontemplate.TemplateRemap(componentID, inputs, topic(o.Kafka), componentID, "Kafka Topic"),
		api.NewConfig(func(c *api.Config) {
			c.Sinks[id] = sinks.NewKafka(func(s *sinks.Kafka) {
				s.BootstrapServers = brokers(o.Kafka)
				s.Topic = fmt.Sprintf("{{ _internal.%s }}", componentID)
				s.Compression = sinks.CompressionType(o.GetTuning().Compression)
				s.Encoding = common.NewApiEncoding(api.CodecTypeJSON)
				s.Encoding.TimestampFormat = "rfc3339"
				s.Batch = common.NewApiBatch(o)
				s.Buffer = common.NewApiBuffer(o)
				kafkaTls(s, o, secrets, op)
				s.HealthCheck = &sinks.HealthCheck{
					Enabled: false,
				}
				sasl(s, o.Kafka.Authentication)
				librdKafkaOptions(s, o)
			}, componentID)
		}),
	}
	return elements
}

func kafkaTls(s *sinks.Kafka, o *adapters.Output, secrets observability.Secrets, op utils.Options) {
	var additionalOptions []framework.Option
	if o.TLS != nil && isTlsBrokers(o.Kafka) {
		additionalOptions = []framework.Option{
			{Name: tls.IncludeEnabled, Value: ""},
			{Name: tls.ExcludeInsecureSkipVerify, Value: ""},
		}
	}
	s.TLS = tls.NewTls(o, secrets, op, additionalOptions...)
}

func librdKafkaOptions(s *sinks.Kafka, o *adapters.Output) {
	s.LibrdKafka_Options = map[string]string{}
	if o.TLS != nil && isTlsBrokers(o.Kafka) && o.TLS.InsecureSkipVerify {
		s.LibrdKafka_Options["enable.ssl.certificate.verification"] = "false"
	}
	if o.GetTuning() != nil && o.GetTuning().MaxWrite != nil {
		s.LibrdKafka_Options["message.max.bytes"] = fmt.Sprintf("%v", o.GetTuning().MaxWrite.Value())
	}
}

func isTlsBrokers(o *obs.Kafka) bool {
	isTls := true
	if o != nil {
		for _, b := range o.Brokers {
			if !strings.HasPrefix(string(b), "tls:") {
				isTls = false
				break
			}
		}
	}
	return isTls
}

func sasl(s *sinks.Kafka, spec *obs.KafkaAuthentication) {
	if spec != nil {
		saslAuth := spec.SASL
		if saslAuth != nil && saslAuth.Username != nil && saslAuth.Password != nil {
			s.Sasl = &sinks.Sasl{
				Enabled:   true,
				Username:  vectorhelpers.SecretFrom(saslAuth.Username),
				Password:  vectorhelpers.SecretFrom(saslAuth.Password),
				Mechanism: SASLMechanismPlain,
			}
			if saslAuth.Mechanism != "" {
				s.Sasl.Mechanism = saslAuth.Mechanism
			}
		}
	}
}

// brokers returns the list of broker endpoints of a Kafka cluster.
// The list represents only the initial set used by the collector's Kafka client for the
// first connection only. The collector's Kafka client fetches constantly an updated list
// from Kafka. These updates are not reconciled back to the collector configuration.
// The list of brokers are populated from the Kafka OutputSpec `Brokers` field, a list of
// valid URLs. If none provided the target URL from the OutputSpec is used as fallback.
// Finally, if neither approach works the current collector process will be terminated.
func brokers(o *obs.Kafka) string {
	brokerUrls := []string{}
	if o.URL != "" {
		brokerUrls = append(brokerUrls, o.URL)
	}
	for _, b := range o.Brokers {
		brokerUrls = append(brokerUrls, string(b))
	}
	brokerHosts := []string{}
	for _, s := range brokerUrls { // Convert URLs to hostnames
		u, _ := url.Parse(s)
		if u != nil {
			brokerHosts = append(brokerHosts, u.Host)
		}
	}
	return strings.Join(brokerHosts, ",")
}

// topic returns the name of an existing kafka topic.
// The kafka topic is either extracted from the kafka OutputSpec `Topic` field in a multiple broker
// setup or as a fallback from the OutputSpec URL if provided as a host path. Defaults to `topic`.
func topic(o *obs.Kafka) string {
	if o != nil && o.Topic != "" {
		return o.Topic
	}

	url, _ := urlhelper.Parse(o.URL)
	topic := strings.TrimLeft(url.Path, "/")
	if topic != "" {
		return topic
	}

	// Fallback to default topic
	return defaultKafkaTopic
}
