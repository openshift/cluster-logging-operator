package loki

import (
	_ "embed"
	"fmt"
	"sort"
	"time"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	testhelpers "github.com/openshift/cluster-logging-operator/test/helpers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/test/helpers/outputs/adapter/fake"
	. "github.com/openshift/cluster-logging-operator/test/matchers"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("outputLabelConf", func() {
	defer GinkgoRecover()
	var (
		loki *obs.Loki
	)
	BeforeEach(func() {
		loki = &obs.Loki{}
	})
	Context("#lokiLabelKeys when LabelKeys", func() {
		Context("are not spec'd", func() {
			It("should provide a default set of labels including the required ones", func() {
				exp := append(DefaultLabelKeys, requiredLabelKeys...)
				sort.Strings(exp)
				Expect(lokiLabelKeys(loki)).To(BeEquivalentTo(exp))
			})
		})
		Context("are spec'd", func() {
			It("should use the ones provided and add the required ones", func() {
				loki.LabelKeys = []string{"foo"}
				exp := append(loki.LabelKeys, requiredLabelKeys...)
				Expect(lokiLabelKeys(loki)).To(BeEquivalentTo(exp))
			})
		})

	})
})

var _ = Describe("Generate vector config", func() {
	const (
		secretName        = "loki-receiver"
		saTokenSecretName = "test-sa-token"
		defaultTLS        = "VersionTLS12"
		defaultCiphers    = "TLS_AES_128_GCM_SHA256,TLS_AES_256_GCM_SHA384,TLS_CHACHA20_POLY1305_SHA256,ECDHE-ECDSA-AES128-GCM-SHA256,ECDHE-RSA-AES128-GCM-SHA256,ECDHE-ECDSA-AES256-GCM-SHA384,ECDHE-RSA-AES256-GCM-SHA384,ECDHE-ECDSA-CHACHA20-POLY1305,ECDHE-RSA-CHACHA20-POLY1305,DHE-RSA-AES128-GCM-SHA256,DHE-RSA-AES256-GCM-SHA384"
	)

	var (
		adapter fake.Output
		secrets = map[string]*corev1.Secret{
			secretName: {
				Data: map[string][]byte{
					constants.ClientUsername:     []byte("testuser"),
					constants.ClientPassword:     []byte("testpass"),
					constants.ClientPrivateKey:   []byte("akey"),
					constants.ClientCertKey:      []byte("acert"),
					constants.TrustedCABundleKey: []byte("aca"),
					constants.TokenKey:           []byte("loki-token"),
				},
			},
			saTokenSecretName: {
				Data: map[string][]byte{
					constants.TokenKey: []byte("test-token"),
				},
			},
		}
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
				Type: obs.OutputTypeLoki,
				Name: "loki-receiver",
				Loki: &obs.Loki{
					URLSpec: obs.URLSpec{
						URL: "https://logs-us-west1.grafana.net",
					},
				},
			}
		}
		baseTune = &obs.BaseOutputTuningSpec{
			DeliveryMode:     obs.DeliveryModeAtLeastOnce,
			MaxWrite:         utils.GetPtr(resource.MustParse("10M")),
			MaxRetryDuration: utils.GetPtr(time.Duration(35)),
			MinRetryDuration: utils.GetPtr(time.Duration(20)),
		}
	)
	DescribeTable("for Loki output", func(expFile string, op framework.Options, tune bool, visit func(spec *obs.OutputSpec)) {
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
		conf := New(helpers.MakeID(outputSpec.Name), outputSpec, []string{"application"}, secrets, adapter, op)
		Expect(string(exp)).To(EqualConfigFrom(conf))
	},
		Entry("with default labels", "with_default_labels.toml", framework.NoOptions, false, func(spec *obs.OutputSpec) {}),
		Entry("with custom labels", "with_custom_labels.toml", framework.NoOptions, false, func(spec *obs.OutputSpec) {
			spec.Loki.LabelKeys = []string{"kubernetes.labels.app", "kubernetes.container_name"}
		}),
		Entry("with tenant id", "with_tenant_id.toml", framework.NoOptions, false, func(spec *obs.OutputSpec) {
			spec.Loki.TenantKey = `foo-{.foo.bar.baz||"none"}`
		}),
		Entry("with custom bearer token", "with_custom_bearer_token.toml", framework.NoOptions, false, func(spec *obs.OutputSpec) {
			spec.Loki.Authentication = &obs.HTTPAuthentication{
				Token: &obs.BearerToken{
					From: obs.BearerTokenFromSecret,
					Secret: &obs.BearerTokenSecretKey{
						Name: secretName,
						Key:  constants.TokenKey,
					},
				},
			}
		}),
		Entry("with custom bearer token", "with_sa_token.toml", framework.Options{
			framework.OptionServiceAccountTokenSecretName: "my-service-account-token",
		}, false, func(spec *obs.OutputSpec) {
			spec.Loki.Authentication = &obs.HTTPAuthentication{
				Token: &obs.BearerToken{
					From: obs.BearerTokenFromServiceAccount,
				},
			}
		}),
		Entry("with username/password token", "with_username_password.toml", framework.NoOptions, false, func(spec *obs.OutputSpec) {
			spec.Loki.Authentication = &obs.HTTPAuthentication{
				Username: &obs.SecretReference{
					Key:        constants.ClientUsername,
					SecretName: secretName,
				},
				Password: &obs.SecretReference{
					Key:        constants.ClientPassword,
					SecretName: secretName,
				},
			}
		}),
		Entry("with TLS insecureSkipVerify=true", "with_insecure.toml", framework.NoOptions, false, func(spec *obs.OutputSpec) {
			spec.TLS = &obs.OutputTLSSpec{
				InsecureSkipVerify: true,
				TLSSpec: obs.TLSSpec{
					CA: &obs.ValueReference{
						Key:        constants.TrustedCABundleKey,
						SecretName: secretName,
					},
				},
			}
		}),
		Entry("with TLS insecureSkipVerify=true, no certificate in secret", "with_insecure_nocert.toml", framework.NoOptions, false, func(spec *obs.OutputSpec) {
			spec.TLS = &obs.OutputTLSSpec{
				InsecureSkipVerify: true,
			}
		}),
		Entry("with TLS", "with_tls.toml", framework.NoOptions, false, func(spec *obs.OutputSpec) {
			spec.TLS = tlsSpec
		}),
		Entry("with TLS config with default minTLSVersion & ciphers", "with_default_tls.toml", testhelpers.FrameworkOptionWithDefaultTLSCiphers, false, func(spec *obs.OutputSpec) {
			spec.TLS = &obs.OutputTLSSpec{
				InsecureSkipVerify: false,
			}
		}),
		Entry("with tuning", "with_tuning.toml", framework.NoOptions, true, func(spec *obs.OutputSpec) {
			spec.Loki.Tuning = &obs.LokiTuningSpec{
				BaseOutputTuningSpec: *baseTune,
			}
		}),
		Entry("with proxy", "with_proxy.toml", framework.NoOptions, true, func(spec *obs.OutputSpec) {
			spec.Loki.ProxyURL = "http://somewhere.org/proxy"
		}),
	)
})
