package http_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/http"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
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
				InsecureSkipVerify: true,
			}
			initOutput = func() obs.OutputSpec {
				return obs.OutputSpec{
					Type: obs.OutputTypeHttp,
					Name: "http-receiver",
					HTTP: &obs.HTTP{
						URLSpec: obs.URLSpec{URL: "https://my-logstore.com"},
						Headers: map[string]string{
							"h2": "v2",
							"h1": "v1",
						},
						Method: "POST",
						Authentication: &obs.HTTPAuthentication{
							Username: &obs.SecretKey{
								Secret: &corev1.LocalObjectReference{
									Name: secretName,
								},
								Key: constants.ClientUsername,
							},
							Password: &obs.SecretKey{
								Secret: &corev1.LocalObjectReference{
									Name: secretName,
								},
								Key: constants.ClientPassword,
							},
						},
					},
				}
			}
		)

		DescribeTable("for HTTP output", func(visit func(spec *obs.OutputSpec), secrets map[string]*corev1.Secret, op framework.Options, expFile string) {
			exp, err := tomlContent.ReadFile(expFile)
			if err != nil {
				Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", expFile, err))
			}
			outputSpec := initOutput()
			if visit != nil {
				visit(&outputSpec)
			}
			conf := http.New(helpers.MakeID(outputSpec.Name), outputSpec, []string{"application"}, secrets, nil, op)
			Expect(string(exp)).To(EqualConfigFrom(conf))
		},
			Entry("with Basic auth", nil, secrets, framework.NoOptions, "http_with_auth_basic.toml"),
			Entry("with token auth", func(spec *obs.OutputSpec) {
				spec.HTTP.Authentication.Token = &obs.BearerToken{
					Key: constants.TokenKey,
					Secret: &corev1.LocalObjectReference{
						Name: secretName,
					},
				}
			}, secrets, framework.NoOptions, "http_with_auth_token.toml"),
			Entry("with token auth", func(spec *obs.OutputSpec) {
				spec.HTTP.Authentication = nil
				spec.TLS = tlsSpec
			}, secrets, framework.NoOptions, "http_with_tls.toml"),
		)
	})

})
