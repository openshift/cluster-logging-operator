package kafka

import (
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers/security"
	corev1 "k8s.io/api/core/v1"
)

func MapKafka(loggingOutSpec logging.OutputSpec, secret *corev1.Secret) *obs.Kafka {
	obsKafka := &obs.Kafka{
		URL: loggingOutSpec.URL,
	}

	if secret != nil {
		obsKafka.Authentication = &obs.KafkaAuthentication{
			SASL: &obs.SASLAuthentication{Mechanism: "PLAIN"},
		}
		if security.HasUsernamePassword(secret) {
			obsKafka.Authentication.SASL.Username = &obs.SecretReference{
				Key:        constants.ClientUsername,
				SecretName: secret.Name,
			}
			obsKafka.Authentication.SASL.Password = &obs.SecretReference{
				Key:        constants.ClientPassword,
				SecretName: secret.Name,
			}
		}
		if security.HasSASLMechanism(secret) {
			if m := security.GetFromSecret(secret, constants.SASLMechanisms); m != "" {
				obsKafka.Authentication.SASL.Mechanism = m
			}
		}
	}

	if loggingOutSpec.Tuning != nil {
		obsKafka.Tuning = &obs.KafkaTuningSpec{
			MaxWrite:    loggingOutSpec.Tuning.MaxWrite,
			Compression: loggingOutSpec.Tuning.Compression,
		}

		switch loggingOutSpec.Tuning.Delivery {
		case logging.OutputDeliveryModeAtLeastOnce:
			obsKafka.Tuning.Delivery = obs.DeliveryModeAtLeastOnce
		case logging.OutputDeliveryModeAtMostOnce:
			obsKafka.Tuning.Delivery = obs.DeliveryModeAtMostOnce
		}
	}

	loggingKafka := loggingOutSpec.Kafka
	if loggingKafka == nil {
		return obsKafka
	}

	obsKafka.Topic = loggingKafka.Topic
	for _, b := range loggingKafka.Brokers {
		obsKafka.Brokers = append(obsKafka.Brokers, obs.URL(b))
	}

	return obsKafka
}
