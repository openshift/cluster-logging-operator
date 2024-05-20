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

//go:embed complex_drop_filter.toml
var ExpectedComplexDropFilterToml string

//go:embed complex_prune_filter.toml
var ExpectedComplexPruneFilterTOML string

// TODO: Use a detailed CLF spec
var _ = Describe("Testing Complete Config Generation", func() {
	defer GinkgoRecover()
	Skip("TODO: enable me after re-wire")

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
		kafkaOutput = obs.OutputSpec{Type: obs.OutputTypeKafka,
			Name: "kafka-receiver",
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
		conf := Conf(secrets, &spec, constants.OpenshiftNS, "my-forwarder", &factory.ForwarderResourceNames{CommonName: constants.CollectorName}, op)
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
						OutputRefs: []string{"kafka-receiver"},
						Name:       "mypipeline",
						//TODO: enable pipeline filter
						//Labels:     map[string]string{"key1": "value1", "key2": "value2"},
					},
				},
				Outputs: []obs.OutputSpec{
					kafkaOutput,
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
						Receiver: &obs.ReceiverSpec{
							Type: obs.ReceiverTypeHTTP,
							Port: 7777,
							HTTP: &obs.HTTPReceiver{
								Format: obs.HTTPReceiverFormatKubeAPIAudit,
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
		//TODO: MOVE INTO UNIT TEST
		//Entry("with drop filters", testhelpers.ConfGenerateTest{
		//	Options: framework.Options{
		//		framework.ClusterTLSProfileSpec: tls.GetClusterTLSProfileSpec(nil),
		//	},
		//	CLFSpec: logging.ClusterLogForwarderSpec{
		//		Inputs: []logging.InputSpec{
		//			{
		//				Name: "mytestapp",
		//				Application: &logging.Application{
		//					Namespaces: []string{"test-ns"},
		//				},
		//			},
		//			{
		//				Name:           logging.InputNameInfrastructure,
		//				Infrastructure: &logging.Infrastructure{},
		//			},
		//			{
		//				Name:  logging.InputNameAudit,
		//				Audit: &logging.Audit{},
		//			},
		//		},
		//		Filters: []logging.FilterSpec{
		//			{
		//				Name: "drop-test",
		//				Type: logging.FilterDrop,
		//				FilterTypeSpec: logging.FilterTypeSpec{
		//					DropTestsSpec: &[]logging.DropTest{
		//						{
		//							DropConditions: []logging.DropCondition{
		//								{
		//									Field:   ".kubernetes.namespace_name",
		//									Matches: "busybox",
		//								},
		//								{
		//									Field:      ".level",
		//									NotMatches: "d.+",
		//								},
		//							},
		//						},
		//						{
		//							DropConditions: []logging.DropCondition{
		//								{
		//									Field:   ".log_type",
		//									Matches: "application",
		//								},
		//							},
		//						},
		//						{
		//							DropConditions: []logging.DropCondition{
		//								{
		//									Field:   ".kubernetes.container_name",
		//									Matches: "error|warning",
		//								},
		//								{
		//									Field:      ".kubernetes.labels.test",
		//									NotMatches: "foo",
		//								},
		//							},
		//						},
		//					},
		//				},
		//			},
		//		},
		//		Pipelines: []logging.PipelineSpec{
		//			{
		//				InputRefs: []string{
		//					"mytestapp",
		//					logging.InputNameInfrastructure,
		//					logging.InputNameAudit},
		//				OutputRefs: []string{"kafka-receiver"},
		//				Name:       "pipeline",
		//				Labels:     map[string]string{"key1": "value1", "key2": "value2"},
		//				FilterRefs: []string{"drop-test"},
		//			},
		//		},
		//		Outputs: []logging.OutputSpec{
		//			{
		//				Type: logging.OutputTypeKafka,
		//				Name: "kafka-receiver",
		//				URL:  "tls://broker1-kafka.svc.messaging.cluster.local:9092/topic",
		//				Secret: &logging.OutputSecretSpec{
		//					Name: "kafka-receiver-1",
		//				},
		//			},
		//		},
		//	},
		//	Secrets: map[string]*corev1.Secret{
		//		"kafka-receiver": {
		//			Data: map[string][]byte{
		//				"tls.key":       []byte("junk"),
		//				"tls.crt":       []byte("junk"),
		//				"ca-bundle.crt": []byte("junk"),
		//			},
		//		},
		//	},
		//	ExpectedConf: ExpectedComplexDropFilterToml,
		//}),

		// TODO: move into unit test
		//Entry("with complex spec with prune filter", testhelpers.ConfGenerateTest{
		//	Options: framework.Options{
		//		framework.ClusterTLSProfileSpec: tls.GetClusterTLSProfileSpec(nil),
		//	},
		//	CLFSpec: logging.ClusterLogForwarderSpec{
		//		Filters: []logging.FilterSpec{
		//			{
		//				Name: "my-prune",
		//				Type: logging.FilterPrune,
		//				FilterTypeSpec: logging.FilterTypeSpec{
		//					PruneFilterSpec: &logging.PruneFilterSpec{
		//						In:    []string{".log_type", ".message", ".kubernetes.container_name"},
		//						NotIn: []string{`.kubernetes.labels."foo-bar/baz"`, ".level"},
		//					},
		//				},
		//			},
		//		},
		//		Inputs: []logging.InputSpec{
		//			{
		//				Name: "mytestapp",
		//				Application: &logging.Application{
		//					Namespaces: []string{"test-ns"},
		//				},
		//			},
		//			{
		//				Name:           logging.InputNameInfrastructure,
		//				Infrastructure: &logging.Infrastructure{},
		//			},
		//			{
		//				Name:  logging.InputNameAudit,
		//				Audit: &logging.Audit{},
		//			},
		//		},
		//		Pipelines: []logging.PipelineSpec{
		//			{
		//				InputRefs: []string{
		//					"mytestapp",
		//					logging.InputNameInfrastructure,
		//					logging.InputNameAudit},
		//				OutputRefs: []string{"kafka-receiver"},
		//				Name:       "pipeline",
		//				Labels:     map[string]string{"key1": "value1", "key2": "value2"},
		//				FilterRefs: []string{"my-prune"},
		//			},
		//		},
		//		Outputs: []logging.OutputSpec{
		//			{
		//				Type: logging.OutputTypeKafka,
		//				Name: "kafka-receiver",
		//				URL:  "tls://broker1-kafka.svc.messaging.cluster.local:9092/topic",
		//				Secret: &logging.OutputSecretSpec{
		//					Name: "kafka-receiver-1",
		//				},
		//			},
		//		},
		//	},
		//	Secrets: map[string]*corev1.Secret{
		//		"kafka-receiver": {
		//			Data: map[string][]byte{
		//				"tls.key":       []byte("junk"),
		//				"tls.crt":       []byte("junk"),
		//				"ca-bundle.crt": []byte("junk"),
		//			},
		//		},
		//	},
		//	ExpectedConf: ExpectedComplexPruneFilterTOML,
		//}),
	)

	Describe("test helper functions", func() {
		It("test MakeInputs", func() {
			diff := cmp.Diff(helpers.MakeInputs("a", "b"), "[\"a\",\"b\"]")
			fmt.Println(diff)
			Expect(diff).To(Equal(""))
		})
	})
})
