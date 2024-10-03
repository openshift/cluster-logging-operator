package otlp

import (
	"fmt"
	"time"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/helpers/outputs/adapter/fake"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("Generate vector config", func() {
	const (
		secretName = "otlp-receiver"
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

		adapter    fake.Output
		initOutput = func() obs.OutputSpec {
			return obs.OutputSpec{
				Type: obs.OutputTypeOTLP,
				Name: "otel-collector",
				OTLP: &obs.OTLP{
					URL: "http://localhost:4318/v1/logs",
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

	DescribeTable("for OTLP output", func(secret observability.Secrets, op framework.Options, tune bool, visit func(spec *obs.OutputSpec), expFile string) {
		exp, err := tomlContent.ReadFile(expFile)
		if err != nil {
			Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", expFile, err))
		}
		outputSpec := initOutput()
		if visit != nil {
			visit(&outputSpec)
		}
		var conf []framework.Element
		if tune {
			adapter = *fake.NewOutput(outputSpec, secret, framework.NoOptions)
			conf = New(helpers.MakeOutputID(outputSpec.Name), outputSpec, []string{"pipeline_my_pipeline_viaq_0"}, secret, adapter, op)
		} else {
			conf = New(helpers.MakeOutputID(outputSpec.Name), outputSpec, []string{"pipeline_my_pipeline_viaq_0"}, secret, nil, op)
		}
		Expect(string(exp)).To(EqualConfigFrom(conf))
	},
		Entry("with only URL spec'd",
			nil,
			framework.NoOptions,
			false,
			nil,
			"otlp_all.toml",
		),
		Entry("with tuning",
			nil,
			framework.NoOptions,
			true,
			func(spec *obs.OutputSpec) {
				spec.OTLP.Tuning = &obs.OTLPTuningSpec{
					BaseOutputTuningSpec: *baseTune,
				}
			},
			"otlp_tuning.toml",
		),
		Entry("with token auth from secret",
			secrets,
			framework.NoOptions,
			false,
			func(spec *obs.OutputSpec) {
				spec.OTLP.Authentication = &obs.HTTPAuthentication{
					Token: &obs.BearerToken{
						From: obs.BearerTokenFromSecret,
						Secret: &obs.BearerTokenSecretKey{
							Key:  constants.TokenKey,
							Name: secretName,
						},
					},
				}
			},
			"otlp_with_auth_token.toml",
		),
		Entry("with token auth from SA",
			secrets,
			framework.Options{
				framework.OptionServiceAccountTokenSecretName: "my-service-account-token",
			},
			false,
			func(spec *obs.OutputSpec) {
				spec.OTLP.Authentication = &obs.HTTPAuthentication{
					Token: &obs.BearerToken{
						From: obs.BearerTokenFromServiceAccount,
					},
				}
			},
			"otlp_with_auth_sa_token.toml",
		),
		Entry("with basic auth",
			secrets,
			framework.NoOptions,
			false,
			func(spec *obs.OutputSpec) {
				spec.OTLP.Authentication = &obs.HTTPAuthentication{
					Username: &obs.SecretReference{
						Key:        constants.ClientUsername,
						SecretName: secretName,
					},
					Password: &obs.SecretReference{
						Key:        constants.ClientPassword,
						SecretName: secretName,
					},
				}
			},
			"otlp_with_auth_basic.toml",
		),
	)
})
