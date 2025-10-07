package azuremonitor

import (
	_ "embed"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/test/helpers/outputs/adapter/fake"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Generating vector config for Azure Monitor Logs output:", func() {

	const (
		sharedKeyValue = "z9ndQSFH1RLDnS6WR35m84u326p3"
		azureId        = "/subscriptions/11111111-1111-1111-1111-111111111111/resourceGroups/otherResourceGroup/providers/Microsoft.Storage/storageAccounts/examplestorage"
		hostCN         = "ods.opinsights.azure.cn"
		customerId     = "6vzw6sHc-0bba-6sHc-4b6c-8bz7sr5eggRt"
		secretName     = "azure-monitor-secret"
		secretTlsName  = "azure-monitor-secret-tls"
		outputName     = "azure_monitor_logs"
		logType        = "myLogType"
		sharedKey      = "shared_key"
	)

	var (
		adapter fake.Output
		secrets = map[string]*corev1.Secret{
			secretName: {
				Data: map[string][]byte{
					sharedKey:            []byte(sharedKeyValue),
					constants.Passphrase: []byte("foo"),
				},
			},
		}

		//tlsSpec = &obs.OutputTLSSpec{
		//	InsecureSkipVerify: true,
		//	TLSSpec: obs.TLSSpec{
		//		CA: &obs.ValueReference{
		//			Key:        constants.TrustedCABundleKey,
		//			SecretName: secretTlsName,
		//		},
		//		Certificate: &obs.ValueReference{
		//			Key:        constants.ClientCertKey,
		//			SecretName: secretTlsName,
		//		},
		//		Key: &obs.SecretReference{
		//			Key:        constants.ClientPrivateKey,
		//			SecretName: secretTlsName,
		//		},
		//		KeyPassphrase: &obs.SecretReference{
		//			Key:        constants.Passphrase,
		//			SecretName: secretName,
		//		},
		//	},
		//}
		initOutput = func() obs.OutputSpec {
			return obs.OutputSpec{
				Type: obs.OutputTypeAzureMonitor,
				Name: outputName,
				AzureMonitor: &obs.AzureMonitor{
					CustomerId: customerId,
					LogType:    logType,
					Authentication: &obs.AzureMonitorAuthentication{
						SharedKey: &obs.SecretReference{
							Key:        "shared_key",
							SecretName: secretName,
						},
					},
				},
			}
		}

		//baseTune = &obs.BaseOutputTuningSpec{
		//	DeliveryMode:     obs.DeliveryModeAtLeastOnce,
		//	MaxWrite:         utils.GetPtr(resource.MustParse("10M")),
		//	MaxRetryDuration: utils.GetPtr(time.Duration(35)),
		//	MinRetryDuration: utils.GetPtr(time.Duration(20)),
		//}
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

		if tune {
			adapter = *fake.NewOutput(outputSpec, secrets, framework.NoOptions)
		}
		conf := New(vectorhelpers.MakeOutputID(outputSpec.Name), outputSpec, []string{"pipelineName"}, secrets, adapter, nil)
		Expect(string(exp)).To(EqualConfigFrom(conf))
	},
		Entry("for common case", nil, false, "azm_common.toml"),
		//Entry("for advance case", func(output *obs.OutputSpec) {
		//	output.AzureMonitor.AzureResourceId = azureId
		//	output.AzureMonitor.Host = hostCN
		//}, false, "azm_advance.toml"),
		//Entry("for common with tls case", func(output *obs.OutputSpec) {
		//	output.TLS = tlsSpec
		//}, false, "azm_tls.toml"),
		//Entry("for common with tls case", func(output *obs.OutputSpec) {
		//	output.AzureMonitor.Tuning = baseTune
		//}, true, "azm_tuning.toml"),
	)
})
