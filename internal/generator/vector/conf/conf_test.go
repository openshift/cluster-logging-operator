package conf

import (
	_ "embed"
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"

	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	. "github.com/openshift/cluster-logging-operator/test/matchers"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
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
		outputName  = "kafka-receiver"
		kafkaOutput = obs.OutputSpec{Type: obs.OutputTypeKafka,
			Name: outputName,
			Kafka: &obs.Kafka{
				URLSpec: obs.URLSpec{
					URL: "tls://broker1-kafka.svc.messaging.cluster.local:9092/topic",
				},
			},
			TLS: &obs.OutputTLSSpec{
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
			},
		}
	)

	DescribeTable("Generate full vector.toml", func(expFile string, op framework.Options, spec obs.ClusterLogForwarderSpec) {
		exp, err := tomlContent.ReadFile(expFile)
		if err != nil {
			Fail(fmt.Sprintf("Error reading the file %q with exp config: %v", expFile, err))
		}
		if op == nil {
			op = clusterOptions
		}
		conf := Conf(secrets, spec, constants.OpenshiftNS, "my-forwarder", factory.ForwarderResourceNames{CommonName: constants.CollectorName}, op)
		Expect(string(exp)).To(EqualConfigFrom(conf))
	},
		Entry("with spec for containers", "container.toml", nil,
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
						Name: "myinfra",
						Type: obs.InputTypeInfrastructure,
						Infrastructure: &obs.Infrastructure{
							Sources: []obs.InfrastructureSource{obs.InfrastructureSourceContainer},
						},
					},
				},
				Pipelines: []obs.PipelineSpec{
					{
						InputRefs: []string{
							"myinfra",
							"mytestapp",
						},
						OutputRefs: []string{outputName},
						Name:       "mypipeline",
						FilterRefs: []string{"my-labels"},
					},
				},
				Outputs: []obs.OutputSpec{
					kafkaOutput,
				},
				Filters: []obs.FilterSpec{
					{
						Name:            "my-labels",
						Type:            obs.FilterTypeOpenshiftLabels,
						OpenShiftLabels: map[string]string{"key1": "value1", "key2": "value2"},
					},
				},
			}),
		Entry("with complex spec", "complex.toml", nil,
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
						OutputRefs: []string{"kafka-receiver"},
						Name:       "pipeline",
						FilterRefs: []string{"my-labels"},
					},
				},
				Filters: []obs.FilterSpec{
					{
						Name:            "my-labels",
						Type:            obs.FilterTypeOpenshiftLabels,
						OpenShiftLabels: map[string]string{"key1": "value1", "key2": "value2"},
					},
				},
				Outputs: []obs.OutputSpec{
					kafkaOutput,
				},
			},
		),
		Entry("with complex spec && http audit receiver as input source", "complex_http_receiver.toml", nil,
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
					{
						Name: "myreceiver",
						Type: obs.InputTypeReceiver,
						Receiver: &obs.ReceiverSpec{
							Type: obs.ReceiverTypeHTTP,
							Port: 7777,
							HTTP: &obs.HTTPReceiver{
								Format: obs.HTTPReceiverFormatKubeAPIAudit,
							},
							TLS: &obs.InputTLSSpec{
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
						},
					},
				},
				Pipelines: []obs.PipelineSpec{
					{
						InputRefs: []string{
							"mytestapp",
							logging.InputNameInfrastructure,
							logging.InputNameAudit,
							"myreceiver",
						},
						OutputRefs: []string{"kafka-receiver"},
						Name:       "pipeline",
					},
				},
				Outputs: []obs.OutputSpec{
					kafkaOutput,
				},
			}),
	)

	Describe("test helper functions", func() {
		It("test MakeInputs", func() {
			diff := cmp.Diff(helpers.MakeInputs("a", "b"), "[\"a\",\"b\"]")
			fmt.Println(diff)
			Expect(diff).To(Equal(""))
		})
	})
})
