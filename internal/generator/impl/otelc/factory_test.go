package otelc_test

import (
	_ "embed"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	"github.com/openshift/cluster-logging-operator/internal/utils/yaml"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Testing Complete Config Generation", func() {
	var (
		clusterOptions = framework.Options{framework.ClusterTLSProfileSpec: tls.GetClusterTLSProfileSpec(nil)}
		secretName     = "kafka-receiver-1"
		secrets        = map[string]*corev1.Secret{
			secretName: {
				Data: map[string][]byte{
					"tls.key":       []byte("junk"),
					"tls.crt":       []byte("junk"),
					"ca-bundle.crt": []byte("junk"),
				},
			},
		}
		outputName = "default-lokistack"
		outputSpec = obs.OutputSpec{
			Type: obs.OutputTypeLokiStack,
			Name: outputName,
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
			TLS: &obs.OutputTLSSpec{
				TLSSpec: obs.TLSSpec{
					CA: &obs.ValueReference{
						Key:           constants.TrustedCABundleKey,
						ConfigMapName: "openshift-service-ca.crt",
					},
				},
			},
		}
	)

	DescribeTable("Generate full vector config", func(expFile string, op framework.Options, spec obs.ClusterLogForwarderSpec) {
		format.MaxLength = 0
		exp, err := expContent.ReadFile(expFile)
		if err != nil {
			Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", expFile, err))
		}
		if op == nil {
			op = clusterOptions
		}
		conf := otelc.NewConfig(secrets, spec, constants.OpenshiftNS, "my-forwarder", factory.ForwarderResourceNames{CommonName: constants.CollectorName}, op)
		os.WriteFile("/tmp/act.yaml", []byte(yaml.MustMarshal(conf)), 0600)
		Expect(yaml.MustMarshal(conf)).To(MatchYAML(exp))
	},
		Entry("with complex spec",
			"complex.yaml",
			nil,
			obs.ClusterLogForwarderSpec{
				Inputs: []obs.InputSpec{
					{
						Name: "mytestapp",
						Type: obs.InputTypeApplication,
						Application: &obs.Application{
							Includes: []obs.NamespaceContainerSpec{
								{Namespace: "test-ns"},
							},
						},
					},
					{
						Name:           string(obs.InputTypeInfrastructure),
						Type:           obs.InputTypeInfrastructure,
						Infrastructure: &obs.Infrastructure{},
					},
					{
						Name:  string(obs.InputTypeAudit),
						Type:  obs.InputTypeAudit,
						Audit: &obs.Audit{},
					},
				},
				Pipelines: []obs.PipelineSpec{
					{
						InputRefs: []string{
							"mytestapp",
							string(obs.InputTypeInfrastructure),
							string(obs.InputTypeAudit),
						},
						OutputRefs: []string{outputName},
						Name:       "pipeline",
						FilterRefs: []string{"my-labels"},
					},
				},
				Filters: []obs.FilterSpec{
					{
						Name:            "my-labels",
						Type:            obs.FilterTypeOpenshiftLabels,
						OpenshiftLabels: map[string]string{"key1": "value1", "key2": "value2"},
					},
				},
				Outputs: []obs.OutputSpec{
					outputSpec,
				},
			},
		),
	)
})
