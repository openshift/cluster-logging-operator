package lokistack

import (
	_ "embed"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
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
		options = utils.Options{
			framework.OptionServiceAccountTokenSecretName: saTokenSecretName,
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
		Entry("with ViaQ datamodel", "lokistack_viaq.toml", options, false, func(spec *obs.OutputSpec) {}),
		Entry("with Otel datamodel", "lokistack_otel.toml", options, false, func(spec *obs.OutputSpec) {
			spec.LokiStack.DataModel = obs.LokiStackDataModelOpenTelemetry
		}),
	)
})
