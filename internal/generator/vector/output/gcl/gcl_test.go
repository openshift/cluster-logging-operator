package gcl_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/gcl"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("Generate Vector config", func() {
	const (
		secretName = "gcl-1"
	)
	var (
		adapter *observability.Output
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
		initOutput = func() obs.OutputSpec {
			return obs.OutputSpec{
				Type: obs.OutputTypeGoogleCloudLogging,
				Name: "gcl_1",
				GoogleCloudLogging: &obs.GoogleCloudLogging{
					ID: obs.GoogleCloudLoggingId{
						Type:  obs.GoogleCloudLoggingIdTypeBillingAccount,
						Value: "billing-1",
					},
					LogId: "vector-1",
					Authentication: &obs.GoogleCloudLoggingAuthentication{
						Credentials: &obs.SecretReference{
							Key:        gcl.GoogleApplicationCredentialsKey,
							SecretName: secretName,
						},
					},
				},
			}
		}
		secrets = map[string]*corev1.Secret{
			secretName: {
				Data: map[string][]byte{
					gcl.GoogleApplicationCredentialsKey: []byte("dummy-credentials"),
					constants.ClientPrivateKey:          []byte("dummy"),
				},
			},
		}

		baseTune = &obs.BaseOutputTuningSpec{
			DeliveryMode:     obs.DeliveryModeAtLeastOnce,
			MaxWrite:         utils.GetPtr(resource.MustParse("10M")),
			MaxRetryDuration: utils.GetPtr(time.Duration(35)),
			MinRetryDuration: utils.GetPtr(time.Duration(20)),
		}
	)
	DescribeTable("For GoogleCloudLogging output", func(visit func(spec *obs.OutputSpec), op utils.Options, expFile string) {
		exp, err := tomlContent.ReadFile(expFile)
		if err != nil {
			Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", expFile, err))
		}
		outputSpec := initOutput()
		if visit != nil {
			visit(&outputSpec)
		}
		adapter = observability.NewOutput(outputSpec)
		conf := gcl.New(outputSpec.Name, adapter, []string{"application"}, secrets, op)
		Expect(string(exp)).To(EqualConfigFrom(conf))
	},
		Entry("with service account token", nil, framework.NoOptions, "gcl_with_token.toml"),
		Entry("with TLS config", func(spec *obs.OutputSpec) {
			spec.TLS = tlsSpec
		}, framework.NoOptions, "gcl_with_tls.toml"),
		Entry("with custom logId", func(spec *obs.OutputSpec) {
			spec.GoogleCloudLogging.LogId = `my-id{.log_type||"none"}`
		}, framework.NoOptions, "gcl_with_custom_logid.toml"),
		Entry("with tuning", func(spec *obs.OutputSpec) {
			spec.GoogleCloudLogging.Tuning = &obs.GoogleCloudLoggingTuningSpec{
				BaseOutputTuningSpec: *baseTune,
			}
		}, framework.NoOptions, "gcl_with_tuning.toml"),
	)
})
