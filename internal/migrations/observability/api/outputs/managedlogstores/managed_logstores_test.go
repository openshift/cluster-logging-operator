package managedlogstores

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
)

var _ = Describe("#MapManagedLogStores", func() {
	It("should generate default elasticsearch output based on logstoreSpec", func() {
		logStoreSpec := logging.LogStoreSpec{
			Type: logging.LogStoreTypeElasticsearch,
		}
		expEsOut := &obs.OutputSpec{
			Name: "default-elasticsearch",
			Type: obs.OutputTypeElasticsearch,
			Elasticsearch: &obs.Elasticsearch{
				URLSpec: obs.URLSpec{
					URL: fmt.Sprintf("https://%s:%d", string(obs.OutputTypeElasticsearch), 9200),
				},
				Version: 6,
				Index:   `{.log_type||"none"}`,
			},
			TLS: &obs.OutputTLSSpec{
				TLSSpec: obs.TLSSpec{
					CA: &obs.ValueReference{
						Key:        constants.TrustedCABundleKey,
						SecretName: constants.ElasticsearchName,
					},
					Certificate: &obs.ValueReference{
						Key:        constants.ClientCertKey,
						SecretName: constants.ElasticsearchName,
					},
					Key: &obs.SecretReference{
						Key:        constants.ClientPrivateKey,
						SecretName: constants.ElasticsearchName,
					},
				},
			},
		}
		Expect(GenerateDefaultOutput(&logStoreSpec)).To(Equal(expEsOut))
	})

	It("should generate default lokistack output", func() {
		logStoreSpec := logging.LogStoreSpec{
			Type: logging.LogStoreTypeLokiStack,
			LokiStack: logging.LokiStackStoreSpec{
				Name: "my-lokistack",
			},
		}
		expLokiStackOut := &obs.OutputSpec{
			Name: "default-lokistack",
			Type: obs.OutputTypeLokiStack,
			LokiStack: &obs.LokiStack{
				Target: obs.LokiStackTarget{
					Name:      "my-lokistack",
					Namespace: constants.OpenshiftNS,
				},
				Authentication: &obs.LokiStackAuthentication{
					Token: &obs.BearerToken{
						From: obs.BearerTokenFromSecret,
						Secret: &obs.BearerTokenSecretKey{
							Name: constants.LogCollectorToken,
							Key:  constants.BearerTokenFileKey,
						},
					},
				},
			},
			TLS: &obs.OutputTLSSpec{
				TLSSpec: obs.TLSSpec{
					CA: &obs.ValueReference{
						Key:        "service-ca.crt",
						SecretName: constants.LogCollectorToken,
					},
				},
			},
		}
		Expect(GenerateDefaultOutput(&logStoreSpec)).To(Equal(expLokiStackOut))
	})
})
