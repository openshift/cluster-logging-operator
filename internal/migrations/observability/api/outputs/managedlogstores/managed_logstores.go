package managedlogstores

import (
	"fmt"

	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
)

const (
	DefaultEsName        = "default-elasticsearch"
	DefaultLokistackName = "default-lokistack"
	DefaultName          = "default-"
)

func GenerateDefaultOutput(logStoreSpec *logging.LogStoreSpec) *obs.OutputSpec {
	var output *obs.OutputSpec
	var outputName string

	switch logStoreSpec.Type {
	case logging.LogStoreTypeElasticsearch:
		outputName = DefaultEsName
		output = &obs.OutputSpec{
			Name: outputName,
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
						SecretName: constants.CollectorSecretName,
					},
					Certificate: &obs.ValueReference{
						Key:        constants.ClientCertKey,
						SecretName: constants.CollectorSecretName,
					},
					Key: &obs.SecretReference{
						Key:        constants.ClientPrivateKey,
						SecretName: constants.CollectorSecretName,
					},
				},
			},
		}
	case logging.LogStoreTypeLokiStack:
		outputName = DefaultLokistackName
		output = &obs.OutputSpec{
			Name: outputName,
			Type: obs.OutputTypeLokiStack,
			LokiStack: &obs.LokiStack{
				Target: obs.LokiStackTarget{
					Name:      logStoreSpec.LokiStack.Name,
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
	}
	return output
}
