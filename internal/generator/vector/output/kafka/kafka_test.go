package kafka

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Generate vector config", func() {
	const (
		secretName = "kafka-receiver-1"
	)

	var (
		tlsSpec = &obs.OutputTLSSpec{
			TLSSpec: obs.TLSSpec{
				CA: &obs.ValueReference{
					Key:        constants.TrustedCABundleKey,
					SecretName: secretName,
				},
				Certificate: &obs.ValueReference{
					Key:        constants.ClientCertKey,
					SecretName: secretName,
				},
				Key: &obs.SecretReference{
					Key:        constants.ClientPrivateKey,
					SecretName: secretName,
				},
			},
		}
		saslAuth = &obs.SASLAuthentication{
			Username: &obs.SecretReference{
				Key:        constants.ClientUsername,
				SecretName: secretName,
			},
			Password: &obs.SecretReference{
				Key:        constants.ClientPassword,
				SecretName: secretName,
			},
		}
		initOutput = func() obs.OutputSpec {
			return obs.OutputSpec{
				Type: obs.OutputTypeKafka,
				Name: "kafka-receiver",
				Kafka: &obs.Kafka{
					URL:   "tcp://broker1-kafka.svc.messaging.cluster.local:9092/topic",
					Topic: "build_complete",
				},
			}
		}

		secrets = map[string]*corev1.Secret{
			secretName: {
				Data: map[string][]byte{
					constants.ClientUsername:     []byte("testuser"),
					constants.ClientPassword:     []byte("testpass"),
					constants.ClientPrivateKey:   []byte("akey"),
					constants.ClientCertKey:      []byte("acert"),
					constants.TrustedCABundleKey: []byte("aca"),
				},
			},
		}
	)

	DescribeTable("for kafka output", func(expFile string, op framework.Options, tlsSpec *obs.OutputTLSSpec, visit func(spec *obs.OutputSpec)) {
		exp, err := tomlContent.ReadFile(expFile)
		if err != nil {
			Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", expFile, err))
		}
		outputSpec := initOutput()
		if visit != nil {
			visit(&outputSpec)
		}
		conf := New(helpers.MakeID(outputSpec.Name), outputSpec, []string{"pipeline_1", "pipeline_2"}, secrets, nil, op)
		Expect(string(exp)).To(EqualConfigFrom(conf))
	},
		Entry("with plaintext sasl, to single topic", "kafka_sasl_plaintext_single_topic.toml", framework.NoOptions, nil, func(spec *obs.OutputSpec) {
			spec.Kafka.Authentication = &obs.KafkaAuthentication{
				SASL: saslAuth,
			}
		}),
		Entry("with plaintext sasl, to single topic", "kafka_sasl_with_tls_single_topic.toml", framework.NoOptions, nil, func(spec *obs.OutputSpec) {
			spec.Kafka.Authentication = &obs.KafkaAuthentication{
				SASL: saslAuth,
			}
			spec.TLS = tlsSpec
		}),
		Entry("with tls sasl, with SCRAM-SHA-256 mechanism to single topic", "kafka_insecure_skipverify.toml", framework.NoOptions, nil, func(spec *obs.OutputSpec) {
			spec.Kafka.URL = "tls://broker1-kafka.svc.messaging.cluster.local:9092/mytopic"
			spec.Kafka.Topic = ""
			spec.TLS = tlsSpec
			tlsSpec.InsecureSkipVerify = true

		}),
		Entry("without security", "kafka_no_security.toml", framework.NoOptions, nil, func(spec *obs.OutputSpec) {
			spec.Kafka.URL = "tcp://broker1-kafka.svc.messaging.cluster.local:9092/topic"
			spec.Kafka.Topic = ""
		}),
		Entry("without custom topic template", "kafka_custom_topic.toml", framework.NoOptions, nil, func(spec *obs.OutputSpec) {
			spec.Kafka.URL = "tcp://broker1-kafka.svc.messaging.cluster.local:9092"
			spec.Kafka.Topic = `foo-bar{.log_type||"none"}`
		}),
		Entry("with NOT tls brokers", "kafka_not_tls_brokers.toml", framework.NoOptions, nil, func(spec *obs.OutputSpec) {
			spec.Kafka.URL = ""
			spec.Kafka.Topic = ""
			spec.Kafka.Brokers = []obs.URL{`tcp://broker1:9092`, `tcp://broker2:9092`, `tcp://broker3:9092`}
		}),
	)
})
