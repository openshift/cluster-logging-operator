package kafka

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common/tls"
	"net/url"
	"strings"

	. "github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"

	genhelper "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	urlhelper "github.com/openshift/cluster-logging-operator/internal/generator/url"
	. "github.com/openshift/cluster-logging-operator/internal/generator/vector/elements"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

const (
	defaultKafkaTopic = "topic"
)

type Kafka struct {
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

func New(id string, o obs.OutputSpec, inputs []string, secrets vectorhelpers.Secrets, strategy common.ConfigStrategy, op Options) []Element {
	if genhelper.IsDebugOutput(op) {
		return []Element{
			Debug(id, vectorhelpers.MakeInputs(inputs...)),
		}
	}

	brokers := Brokers(o)
	sink := sink(id, o, inputs, op, brokers)
	if strategy != nil {
		strategy.VisitSink(sink)
	}
	tlsConfig := []Element{Nil}
	if o.TLS != nil {
		skipVerify := o.TLS.InsecureSkipVerify
		o.TLS.InsecureSkipVerify = false
		tlsConfig = []Element{tls.New(id, o.TLS, secrets, op)}
		if skipVerify {
			tlsConfig = append(tlsConfig, InsecureTLS{
				ComponentID: id,
			})
		}
	}
	elements := []Element{
		sink,
		Encoding(id, op),
		common.NewAcknowledgments(id, strategy),
		common.NewBatch(id, strategy),
		common.NewBuffer(id, strategy),
		SASLConf(id, o.Kafka.Authentication, secrets),
	}
	elements = append(elements,
		tlsConfig...,
	)
	return elements
}

func sink(id string, o obs.OutputSpec, inputs []string, op Options, brokers string) *Kafka {
	return &Kafka{
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
func Brokers(o obs.OutputSpec) string {
	brokerUrls := []string{}
	if o.Kafka.URL != "" {
		brokerUrls = append(brokerUrls, o.Kafka.URL)
	}
	brokerUrls = append(brokerUrls, o.Kafka.Brokers...)
	brokerHosts := []string{}
	for _, s := range brokerUrls { // Convert URLs to hostnames
		u, _ := url.Parse(s)
		if u != nil {
			brokerHosts = append(brokerHosts, u.Host)
		}
	}
	return strings.Join(brokerHosts, ",")
}

// Topic returns the name of an existing kafka topic.
// The kafka topic is either extracted from the kafka OutputSpec `Topic` field in a multiple broker
// setup or as a fallback from the OutputSpec URL if provided as a host path. Defaults to `topic`.
func Topics(o obs.OutputSpec) string {
	if o.Kafka != nil && o.Kafka.Topic != "" {
		return o.Kafka.Topic
	}

	url, _ := urlhelper.Parse(o.Kafka.URL)
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
