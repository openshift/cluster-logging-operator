package fluentd

import (
	"net/url"
	"strings"

	"github.com/ViaQ/logerr/log"
)

const defaultKafkaTopic = "topic"

//Brokers returns the list of broker endpoints of a kafka cluster.
//The list represents only the initial set used by the collector's kafka client for the
//first connention only. The collector's kafka client fetches constantly an updated list
//from kafka. These updates are not reconciled back to the collector configuration.
//The list of brokers are populated from the Kafka OutputSpec `Brokers` field, a list of
//valid URLs. If none provided the target URL from the OutputSpec is used as fallback.
//Finally, if neither approach works the current collector process will be terminated.
func (conf *outputLabelConf) Brokers() string {
	parseBroker := func(b string) string {
		url, err := url.Parse(b)
		if err != nil {
			log.Error(err, "Failed to parse Kafka broker from output spec", "spec", b)
			return ""
		}
		return url.Host
	}

	if conf.Target.Kafka != nil {
		if conf.Target.Kafka.Brokers != nil {
			brokers := []string{}
			for _, broker := range conf.Target.Kafka.Brokers {
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
	fallback := parseBroker(conf.Target.URL)
	if fallback == "" {
		panic("Failed to parse any Kafka broker from output spec")
	}

	return fallback
}

//Topic returns the name of an existing kafka topic.
//The kafka topic is either extracted from the kafka OutputSpec `Topic` field in a multiple broker
//setup or as a fallback from the OutputSpec URL if provided as a host path. Defaults to `topic`.
func (conf *outputLabelConf) Topic() string {
	if conf.Target.Kafka != nil && conf.Target.Kafka.Topic != "" {
		return conf.Target.Kafka.Topic
	}

	url, err := url.Parse(conf.Target.URL)
	if err != nil {
		log.Error(err, "Failed to extract Kafka topic from output spec url", "url", conf.Target.URL)
	}

	topic := strings.TrimLeft(url.Path, "/")
	if topic != "" {
		return topic
	}

	// Fallback to default topic
	return defaultKafkaTopic
}
