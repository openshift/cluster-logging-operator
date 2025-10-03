package otlp

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/helpers/outputs/adapter/fake"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
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

		initOptions = func() utils.Options {
			output := initOutput()
			return utils.Options{
				helpers.CLFSpec: observability.ClusterLogForwarderSpec(obs.ClusterLogForwarderSpec{
					Outputs: []obs.OutputSpec{output},
					Pipelines: []obs.PipelineSpec{
						{
							Name:       "lokistack",
							InputRefs:  []string{obs.InputTypeApplication.String(), obs.InputTypeInfrastructure.String(), obs.InputTypeAudit.String()},
							OutputRefs: []string{output.Name},
						},
					},
					Inputs: []obs.InputSpec{
						{Name: obs.InputTypeApplication.String(), Type: obs.InputTypeApplication},
						{Name: obs.InputTypeInfrastructure.String(), Type: obs.InputTypeInfrastructure, Infrastructure: &obs.Infrastructure{Sources: obs.InfrastructureSources}},
						{Name: obs.InputTypeAudit.String(), Type: obs.InputTypeAudit, Audit: &obs.Audit{Sources: obs.AuditSources}},
					},
				}),
			}
		}
	)

	DescribeTable("for OTLP output", func(secret observability.Secrets, op utils.Options, tune bool, visit func(spec *obs.OutputSpec), expFile string) {
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
			initOptions(),
			false,
			nil,
			"otlp_all.toml",
		),
		Entry("with only some audit sources",
			nil,
			utils.Options{
				OtlpLogSourcesOption: []string{obs.AuditSourceKube.String(), obs.AuditSourceOpenShift.String()},
			},
			false,
			nil,
			"otlp_audit_two_sources.toml",
		),
		Entry("with base tuning and compression",
			nil,
			initOptions(),
			true,
			func(spec *obs.OutputSpec) {
				spec.OTLP.Tuning = &obs.OTLPTuningSpec{
					BaseOutputTuningSpec: *baseTune,
					Compression:          "gzip",
				}
			},
			"otlp_tuning.toml",
		),
		Entry("with token auth from secret",
			secrets,
			initOptions(),
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
			initOptions().Set(framework.OptionServiceAccountTokenSecretName, "my-service-account-token"),
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
			initOptions(),
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
