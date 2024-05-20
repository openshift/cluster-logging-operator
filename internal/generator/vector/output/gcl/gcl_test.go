package gcl_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/gcl"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Generate Vector config", func() {
	const (
		secretName = "gcl-1"
	)
	var (
		tlsSpec = &obs.OutputTLSSpec{
			TLSSpec: obs.TLSSpec{
				CA: &obs.ConfigMapOrSecretKey{
					Secret: &corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: constants.TrustedCABundleKey,
				},
				Certificate: &obs.ConfigMapOrSecretKey{
					Secret: &corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: constants.ClientCertKey,
				},
				Key: &obs.SecretKey{
					Secret: &corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: constants.ClientPrivateKey,
				},
			},
		}
		initOutput = func() obs.OutputSpec {
			return obs.OutputSpec{
				Type: obs.OutputTypeGoogleCloudLogging,
				Name: "gcl_1",
				GoogleCloudLogging: &obs.GoogleCloudLogging{
					BillingAccountID: "billing-1",
					LogID:            "vector-1",
					Authentication: &obs.GoogleCloudLoggingAuthentication{
						Credentials: &obs.SecretKey{
							Secret: &corev1.LocalObjectReference{
								Name: secretName,
							},
							Key: gcl.GoogleApplicationCredentialsKey,
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
	)
	DescribeTable("For GoogleCloudLogging output", func(visit func(spec *obs.OutputSpec), op framework.Options, expFile string) {
		exp, err := tomlContent.ReadFile(expFile)
		if err != nil {
			Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", expFile, err))
		}
		outputSpec := initOutput()
		if visit != nil {
			visit(&outputSpec)
		}
		conf := gcl.New(outputSpec.Name, outputSpec, []string{"application"}, secrets, nil, op)
		Expect(string(exp)).To(EqualConfigFrom(conf))
	},
		Entry("with service account token", nil, framework.NoOptions, "gcl_with_token.toml"),
		Entry("with TLS config", func(spec *obs.OutputSpec) {
			spec.TLS = tlsSpec
		}, framework.NoOptions, "gcl_with_tls.toml"),
	)
})
