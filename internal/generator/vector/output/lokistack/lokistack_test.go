package lokistack

import (
	_ "embed"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/helpers/outputs/adapter/fake"
	. "github.com/openshift/cluster-logging-operator/test/matchers"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Generate vector config", func() {
	const (
		configMapName     = "openshift-service-ca.crt"
		saTokenSecretName = "test-sa-token"
	)

	var (
		adapter fake.Output
		//- lokiStack:
		//authentication:
		//token:
		//from: serviceAccount
		//target:
		//name: logging-loki
		//namespace: openshift-logging
		//name: default-lokistack
		//tls:
		//ca:
		//configMapName: openshift-service-ca.crt
		//key: service-ca.crt
		//type: lokiStack
		secrets = map[string]*corev1.Secret{
			configMapName: {
				Data: map[string][]byte{
					"service-ca.crt": []byte("testuser"),
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
					Key:           constants.TrustedCABundleKey,
					ConfigMapName: "openshift-service-ca.crt",
				},
			},
		}
		initOutput = func() obs.OutputSpec {
			return obs.OutputSpec{
				Type: obs.OutputTypeLokiStack,
				Name: "default-lokistack",
				LokiStack: &obs.LokiStack{
					Authentication: &obs.LokiStackAuthentication{
						Token: &obs.BearerToken{
							From: obs.BearerTokenFromServiceAccount,
						},
					},
					Target: obs.LokiStackTarget{
						Namespace: "openshift-logging",
						Name:      "logging-loki",
					},
				},
				TLS: tlsSpec,
			}
		}
		initOptions = func() utils.Options {
			output := initOutput()
			return utils.Options{
				framework.OptionServiceAccountTokenSecretName: saTokenSecretName,
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

		initReceiverOptions = func() utils.Options {
			output := initOutput()
			return utils.Options{
				framework.OptionServiceAccountTokenSecretName: saTokenSecretName,
				helpers.CLFSpec: observability.ClusterLogForwarderSpec(obs.ClusterLogForwarderSpec{
					Outputs: []obs.OutputSpec{output},
					Pipelines: []obs.PipelineSpec{
						{
							Name:       "lokistack-receivers",
							InputRefs:  []string{"http-receiver", "syslog-receiver"},
							OutputRefs: []string{output.Name},
						},
					},
					Inputs: []obs.InputSpec{
						{
							Name: "http-receiver",
							Type: obs.InputTypeReceiver,
							Receiver: &obs.ReceiverSpec{
								Type: obs.ReceiverTypeHTTP,
							},
						},
						{
							Name: "syslog-receiver",
							Type: obs.InputTypeReceiver,
							Receiver: &obs.ReceiverSpec{
								Type: obs.ReceiverTypeSyslog,
							},
						},
					},
				}),
			}
		}
	)
	DescribeTable("for LokiStack output", func(expFile string, op framework.Options, tune bool, visit func(spec *obs.OutputSpec)) {
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
		conf := New(helpers.MakeOutputID(outputSpec.Name), outputSpec, []string{"pipeline_fake"}, secrets, adapter, op)
		Expect(string(exp)).To(EqualConfigFrom(conf))
	},
		Entry("with ViaQ datamodel", "lokistack_viaq.toml", initOptions(), false, func(spec *obs.OutputSpec) {}),
		Entry("with Otel datamodel", "lokistack_otel.toml", initOptions(), false, func(spec *obs.OutputSpec) {
			spec.LokiStack.DataModel = obs.LokiStackDataModelOpenTelemetry
		}),
		Entry("with ViaQ datamodel with receiver", "lokistack_viaq_receiver.toml", initReceiverOptions(), false, func(spec *obs.OutputSpec) {}),
	)
})
