package loki

import (
	_ "embed"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/logstore/lokistack"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	"github.com/openshift/cluster-logging-operator/test/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	v1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("outputLabelConf", func() {
	defer GinkgoRecover()
	Skip("TODO: Enable me after rewire")
	var (
		loki *logging.Loki
	)
	BeforeEach(func() {
		loki = &logging.Loki{}
	})
	Context("#lokiLabelKeys when LabelKeys", func() {
		Context("are not spec'd", func() {
			It("should provide a default set of labels including the required ones", func() {
				exp := append(defaultLabelKeys, requiredLabelKeys...)
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

//go:embed with_default_labels.toml
var withDefaultLabels string

//go:embed with_custom_labels.toml
var withCustomLabels string

//go:embed with_tenant_id.toml
var withTenantId string

//go:embed with_custom_bearer_token.toml
var withCustomBearerToken string

//go:embed with_insecure.toml
var withInsecure string

//go:embed with_insecure_nocert.toml
var withInsecureNoCert string

//go:embed with_default_tls.toml
var withDefaultTls string

//go:embed with_default_logcollector_bearer_token.toml
var withDefaultLogcollectorToken string

var _ = Describe("Generate vector config", func() {
	defer GinkgoRecover()
	Skip("TODO: Enable me after rewire")
	defaultTLS := "VersionTLS12"
	defaultCiphers := "TLS_AES_128_GCM_SHA256,TLS_AES_256_GCM_SHA384,TLS_CHACHA20_POLY1305_SHA256,ECDHE-ECDSA-AES128-GCM-SHA256,ECDHE-RSA-AES128-GCM-SHA256,ECDHE-ECDSA-AES256-GCM-SHA384,ECDHE-RSA-AES256-GCM-SHA384,ECDHE-ECDSA-CHACHA20-POLY1305,ECDHE-RSA-CHACHA20-POLY1305,DHE-RSA-AES128-GCM-SHA256,DHE-RSA-AES256-GCM-SHA384"
	inputPipeline := []string{"application"}
	var f = func(clspec logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op framework.Options) []framework.Element {
		elements := New(vectorhelpers.FormatComponentID(clfspec.Outputs[0].Name), clfspec.Outputs[0], inputPipeline, secrets[clfspec.Outputs[0].Name], nil, op)

		return elements
	}
	DescribeTable("for Loki output", helpers.TestGenerateConfWith(f),
		Entry("with default labels", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeLoki,
						Name: "loki-receiver",
						URL:  "https://logs-us-west1.grafana.net",
						Secret: &logging.OutputSecretSpec{
							Name: "loki-receiver",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"loki-receiver": {
					Data: map[string][]byte{
						"username": []byte("username"),
						"password": []byte("password"),
					},
				},
			},
			ExpectedConf: withDefaultLabels,
		}),
		Entry("with custom labels", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeLoki,
						Name: "loki-receiver",
						URL:  "https://logs-us-west1.grafana.net",
						Secret: &logging.OutputSecretSpec{
							Name: "loki-receiver",
						},
						OutputTypeSpec: v1.OutputTypeSpec{Loki: &v1.Loki{
							LabelKeys: []string{"kubernetes.labels.app", "kubernetes.container_name"},
						}},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"loki-receiver": {
					Data: map[string][]byte{
						"username": []byte("username"),
						"password": []byte("password"),
					},
				},
			},
			ExpectedConf: withCustomLabels,
		}),
		Entry("with tenant id", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeLoki,
						Name: "loki-receiver",
						URL:  "https://logs-us-west1.grafana.net",
						Secret: &logging.OutputSecretSpec{
							Name: "loki-receiver",
						},
						OutputTypeSpec: v1.OutputTypeSpec{Loki: &v1.Loki{
							TenantKey: "foo.bar.baz",
						}},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"loki-receiver": {
					Data: map[string][]byte{
						"username": []byte("username"),
						"password": []byte("password"),
					},
				},
			},
			ExpectedConf: withTenantId,
		}),
		Entry("with custom bearer token", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeLoki,
						Name: "loki-receiver",
						URL:  "http://lokistack-dev-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
						Secret: &logging.OutputSecretSpec{
							Name: "custom-loki-secret",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"loki-receiver": {
					Data: map[string][]byte{
						"token": []byte("token-for-custom-loki"),
					},
				},
			},
			ExpectedConf: withCustomBearerToken,
		}),
		Entry("with TLS insecureSkipVerify=true", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeLoki,
						Name: "loki-receiver",
						URL:  "https://lokistack-dev-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
						Secret: &logging.OutputSecretSpec{
							Name: "custom-loki-secret",
						},
						TLS: &logging.OutputTLSSpec{
							InsecureSkipVerify: true,
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"loki-receiver": {
					Data: map[string][]byte{
						"ca-bundle.crt": []byte("junk"),
					},
				},
			},
			ExpectedConf: withInsecure,
		}),
		Entry("with TLS insecureSkipVerify=true, no certificate in secret", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeLoki,
						Name: "loki-receiver",
						URL:  "https://lokistack-dev-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
						TLS: &logging.OutputTLSSpec{
							InsecureSkipVerify: true,
						},
					},
				},
			},
			ExpectedConf: withInsecureNoCert,
		}),
		Entry("with TLS config with default minTLSVersion & ciphers", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeLoki,
						Name: "loki-receiver",
						URL:  "https://lokistack-dev-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
						Secret: &logging.OutputSecretSpec{
							Name: "custom-loki-secret",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"loki-receiver": {
					Data: map[string][]byte{
						"token": []byte("token-for-custom-loki"),
					},
				},
			},
			Options: framework.Options{
				framework.MinTLSVersion: string(tls.DefaultMinTLSVersion),
				framework.Ciphers:       strings.Join(tls.DefaultTLSCiphers, ","),
			},
			ExpectedConf: fmt.Sprintf(withDefaultTls, defaultTLS, defaultCiphers),
		}),
	)
})

var _ = Describe("Generate vector config for in cluster loki", func() {
	inputPipeline := []string{"application"}
	var f = func(clspec logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op framework.Options) []framework.Element {
		return New(vectorhelpers.FormatComponentID(clfspec.Outputs[0].Name), clfspec.Outputs[0], inputPipeline, secrets[constants.LogCollectorToken], nil, framework.NoOptions)
	}
	DescribeTable("for Loki output", helpers.TestGenerateConfWith(f),
		Entry("with default logcollector bearer token", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeLoki,
						Name: lokistack.FormatOutputNameFromInput(logging.InputNameApplication),
						URL:  "http://lokistack-dev-gateway-http.openshift-logging.svc:8080/api/logs/v1/application",
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				constants.LogCollectorToken: {
					Data: map[string][]byte{
						"token": []byte("token-for-internal-loki"),
					},
				},
			},
			ExpectedConf: withDefaultLogcollectorToken,
		}),
	)
})

func TestVectorConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vector Conf Generation")
}
