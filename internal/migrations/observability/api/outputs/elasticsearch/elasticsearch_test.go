package elasticsearch

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("#MapElasticsearch", func() {
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
	It("should map logging.Elasticsearch to obs.Elasticsearch", func() {
		secret.Data[constants.ClientUsername] = []byte("user")
		secret.Data[constants.ClientPassword] = []byte("pass")

		loggingOutSpec := logging.OutputSpec{
			URL: url,
			OutputTypeSpec: logging.OutputTypeSpec{
				Elasticsearch: &logging.Elasticsearch{
					Version: 8,
					ElasticsearchStructuredSpec: logging.ElasticsearchStructuredSpec{
						StructuredTypeKey:  "namespace",
						StructuredTypeName: "structName",
					},
				},
			},
			Tuning: &logging.OutputTuningSpec{
				Delivery:    logging.OutputDeliveryModeAtLeastOnce,
				Compression: "gzip",
			},
		}

		expObsElastic := &obs.Elasticsearch{
			URLSpec: obs.URLSpec{
				URL: url,
			},
			Version: 8,
			Index:   `{.namespace||"structName"}`,
			Authentication: &obs.HTTPAuthentication{
				Username: &obs.SecretReference{
					Key:        constants.ClientUsername,
					SecretName: secretName,
				},
				Password: &obs.SecretReference{
					Key:        constants.ClientPassword,
					SecretName: secretName,
				},
			},
			Tuning: &obs.ElasticsearchTuningSpec{
				BaseOutputTuningSpec: obs.BaseOutputTuningSpec{
					Delivery: obs.DeliveryModeAtLeastOnce,
				},
				Compression: "gzip",
			},
		}

		Expect(MapElasticsearch(loggingOutSpec, secret)).To(Equal(expObsElastic))
	})
})
