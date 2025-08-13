package http_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/http"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/helpers/outputs/adapter/fake"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("Generate vector config", func() {

	Context("#New", func() {

		const (
			secretName = "http-receiver"
			aUserName  = "username"
			aPassword  = "password"
			aToken     = "atoken"
		)
		var (
			adapter fake.Output
			secrets = map[string]*corev1.Secret{
				secretName: {
					Data: map[string][]byte{
						constants.ClientUsername: []byte(aUserName),
						constants.ClientPassword: []byte(aPassword),
						constants.TokenKey:       []byte(aToken),
					},
				},
			}
			tlsSpec = &obs.OutputTLSSpec{
				InsecureSkipVerify: true,
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
					Type: obs.OutputTypeHTTP,
					Name: "http-receiver",
					HTTP: &obs.HTTP{
						URLSpec: obs.URLSpec{URL: "https://my-logstore.com"},
						Headers: map[string]string{
							"h2": "v2",
							"h1": "v1",
						},
						Method: "POST",
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

		DescribeTable("for HTTP output", func(visit func(spec *obs.OutputSpec), secrets map[string]*corev1.Secret, tune bool, op framework.Options, expFile string) {
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
			conf := http.New(helpers.MakeID(outputSpec.Name), outputSpec, []string{"application"}, secrets, adapter, op)
			Expect(string(exp)).To(EqualConfigFrom(conf))
		},
			Entry("with Basic auth", nil, secrets, false, framework.NoOptions, "http_with_auth_basic.toml"),
			Entry("with token auth", func(spec *obs.OutputSpec) {
				spec.HTTP.Authentication.Token = &obs.BearerToken{
					From: obs.BearerTokenFromSecret,
					Secret: &obs.BearerTokenSecretKey{
						Key:  constants.TokenKey,
						Name: secretName,
					},
				}
			}, secrets, false, framework.NoOptions, "http_with_auth_token.toml"),
			Entry("with token auth", func(spec *obs.OutputSpec) {
				spec.HTTP.Authentication = nil
				spec.TLS = tlsSpec
			}, secrets, false, framework.NoOptions, "http_with_tls.toml"),
			Entry("with token auth", func(spec *obs.OutputSpec) {
				spec.HTTP.Authentication = nil
				spec.TLS = &obs.OutputTLSSpec{
					TLSSpec: obs.TLSSpec{
						CA: &obs.ValueReference{
							Key:           "ca.crt",
							ConfigMapName: secretName,
						},
						Certificate: &obs.ValueReference{
							Key:           "my.crt",
							ConfigMapName: secretName,
						},
						Key: &obs.SecretReference{
							Key:        constants.ClientPrivateKey,
							SecretName: secretName,
						},
					},
				}
			}, secrets, false, framework.NoOptions, "http_with_tls_using_configmaps.toml"),
			Entry("with tuning", func(spec *obs.OutputSpec) {
				spec.HTTP.Tuning = &obs.HTTPTuningSpec{
					BaseOutputTuningSpec: *baseTune,
				}
			}, secrets, true, framework.NoOptions, "http_with_tuning.toml"),
			Entry("with ndjson", func(spec *obs.OutputSpec) {
				spec.HTTP.LinePerEvent = true
			}, secrets, true, framework.NoOptions, "http_with_ndjson.toml"),
			Entry("with proxy", func(spec *obs.OutputSpec) {
				spec.HTTP.ProxyURL = "http://somewhere.org/proxy"
				spec.HTTP.Headers = nil
				spec.HTTP.Authentication = nil
			}, secrets, true, framework.NoOptions, "http_with_proxy.toml"),
		)
	})

})
