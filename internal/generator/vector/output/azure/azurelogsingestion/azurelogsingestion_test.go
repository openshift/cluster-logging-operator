package azurelogsingestion

import (
	_ "embed"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/adapters"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("Generating vector config for Azure Log Ingestion output:", func() {

	const (
		secretName      = "azure-log-ingestion-secret"
		secretTlsName   = "azure-log-ingestion-secret-tls"
		outputName      = "azure-log-ingestion"
		endpoint        = "https://my-dce-abcdefghij.westus-1.ingest.monitor.azure.com"
		dcrImmutableId  = "dcr-a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6"
		streamName      = "Custom-MyTable_Logs_0_CL"
		tenantId        = "a0b1c2d3-e4f5-a6b7-c8d9-e0f1a2b3c4d5"
		clientId        = "b1c2d3e4-f5a6-b7c8-d9e0-f1a2b3c4d5e6"
		clientSecretVal = "AbCdE~FgH1IjKlMnOpQrStUvWxYz0123456789Ab"
		clientSecretKey = "client_secret"
	)

	var (
		adapter *adapters.Output
		secrets = map[string]*corev1.Secret{
			secretName: {
				Data: map[string][]byte{
					clientSecretKey:      []byte(clientSecretVal),
					constants.Passphrase: []byte("foo"),
				},
			},
		}

		tlsSpec = &obs.OutputTLSSpec{
			InsecureSkipVerify: true,
			TLSSpec: obs.TLSSpec{
				CA: &obs.ValueReference{
					Key:        constants.TrustedCABundleKey,
					SecretName: secretTlsName,
				},
				Certificate: &obs.ValueReference{
					Key:        constants.ClientCertKey,
					SecretName: secretTlsName,
				},
				Key: &obs.SecretReference{
					Key:        constants.ClientPrivateKey,
					SecretName: secretTlsName,
				},
				KeyPassphrase: &obs.SecretReference{
					Key:        constants.Passphrase,
					SecretName: secretName,
				},
			},
		}

		baseTune = &obs.BaseOutputTuningSpec{
			DeliveryMode:     obs.DeliveryModeAtLeastOnce,
			MaxWrite:         utils.GetPtr(resource.MustParse("10M")),
			MaxRetryDuration: utils.GetPtr(time.Duration(35)),
			MinRetryDuration: utils.GetPtr(time.Duration(20)),
		}

		initOutput = func() obs.OutputSpec {
			return obs.OutputSpec{
				Type: obs.OutputTypeAzureLogsIngestion,
				Name: outputName,
				AzureLogsIngestion: &obs.AzureLogsIngestion{
					URLSpec: obs.URLSpec{
						URL: endpoint,
					},
					DcrImmutableId: dcrImmutableId,
					StreamName:     streamName,
					Authentication: &obs.AzureLogsIngestionAuthentication{
						Type: obs.AzureLogsIngestionAuthTypeClientSecret,
						ClientSecret: &obs.AzureLogsIngestionClientSecret{
							TenantId: tenantId,
							ClientId: clientId,
							Secret: &obs.SecretReference{
								Key:        clientSecretKey,
								SecretName: secretName,
							},
						},
					},
				},
			}
		}
	)

	DescribeTable("should generate valid config", func(visit func(output *obs.OutputSpec), tune bool, expFile string) {
		exp, err := tomlContent.ReadFile(expFile)
		if err != nil {
			Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", expFile, err))
		}
		outputSpec := initOutput()

		if visit != nil {
			visit(&outputSpec)
		}

		adapter = adapters.NewOutput(outputSpec)

		id, sink, transforms := New(vectorhelpers.MakeOutputID(outputSpec.Name), adapter, []string{"pipelineName"}, secrets, nil)
		Expect(exp).To(EqualConfigFrom(api.NewConfig(func(c *api.Config) {
			c.Sinks[id] = sink
			c.AddTransforms(transforms)
		})))
	},
		Entry("for client secret auth", nil, false, "azli_common.toml"),
		Entry("for workload identity auth", func(output *obs.OutputSpec) {
			output.AzureLogsIngestion.Authentication = &obs.AzureLogsIngestionAuthentication{
				Type: obs.AzureLogsIngestionAuthTypeWorkloadIdentity,
				WorkloadIdentity: &obs.AzureLogsIngestionWorkloadIdentity{
					TenantId: tenantId,
					ClientId: clientId,
					Token: &obs.BearerToken{
						From: obs.BearerTokenFromServiceAccount,
					},
				},
			}
		}, false, "azli_workload_identity.toml"),
		Entry("with timestamp field", func(output *obs.OutputSpec) {
			output.AzureLogsIngestion.TimestampField = "Timestamp"
		}, false, "azli_timestamp_field.toml"),
		Entry("with token scope", func(output *obs.OutputSpec) {
			output.AzureLogsIngestion.TokenScope = "https://monitor.azure.cn/.default"
		}, false, "azli_token_scope.toml"),
		Entry("with tls settings", func(output *obs.OutputSpec) {
			output.TLS = tlsSpec
		}, false, "azli_tls.toml"),
		Entry("with tuning parameters", func(output *obs.OutputSpec) {
			output.AzureLogsIngestion.Tuning = &obs.AzureLogsIngestionTuningSpec{
				BaseOutputTuningSpec: *baseTune,
			}
		}, true, "azli_tuning.toml"),
	)
})
