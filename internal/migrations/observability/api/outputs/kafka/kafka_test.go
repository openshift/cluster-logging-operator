package kafka

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("#MapKafka", func() {
	const secretName = "my-secret"
	var (
		secret *corev1.Secret
		url    = "0.0.0.0:9200"
	)
	BeforeEach(func() {
		secret = &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      secretName,
				Namespace: "foo-space",
			},
			Data: map[string][]byte{
				constants.ClientCertKey:      []byte("cert"),
				constants.ClientPrivateKey:   []byte("privatekey"),
				constants.TrustedCABundleKey: []byte("cabundle"),
				constants.Passphrase:         []byte("pass"),
			},
		}
	})
	It("should map logging.Kafka to obs.Kafka", func() {
		secret.Data[constants.ClientUsername] = []byte("user")
		secret.Data[constants.ClientPassword] = []byte("pass")
		secret.Data[constants.SASLMechanisms] = []byte("SCRAM-SHA-256")

		loggingOutSpec := logging.OutputSpec{
			URL: url,
			OutputTypeSpec: logging.OutputTypeSpec{
				Kafka: &logging.Kafka{
					Topic:   "foo",
					Brokers: []string{"foo", "bar"},
				},
			},
			Tuning: &logging.OutputTuningSpec{
				Delivery:    logging.OutputDeliveryModeAtLeastOnce,
				MaxWrite:    utils.GetPtr(resource.MustParse("100m")),
				Compression: "zstd",
			},
		}

		expObsKafka := &obs.Kafka{
			URL:     url,
			Topic:   "foo",
			Brokers: []obs.URL{"foo", "bar"},
			Authentication: &obs.KafkaAuthentication{
				SASL: &obs.SASLAuthentication{
					Username: &obs.SecretReference{
						Key:        constants.ClientUsername,
						SecretName: secretName,
					},
					Password: &obs.SecretReference{
						Key:        constants.ClientPassword,
						SecretName: secretName,
					},
					Mechanism: "SCRAM-SHA-256",
				},
			},
			Tuning: &obs.KafkaTuningSpec{
				Delivery:    obs.DeliveryModeAtLeastOnce,
				MaxWrite:    utils.GetPtr(resource.MustParse("100m")),
				Compression: "zstd",
			},
		}

		Expect(MapKafka(loggingOutSpec, secret)).To(Equal(expObsKafka))
	})
})
