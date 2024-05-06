package conf

import (
	_ "embed"
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	testhelpers "github.com/openshift/cluster-logging-operator/test/helpers"
	. "github.com/openshift/cluster-logging-operator/test/matchers"

	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	corev1 "k8s.io/api/core/v1"
)

//go:embed complex.toml
var ExpectedComplexToml string

//go:embed complex_drop_filter.toml
var ExpectedComplexDropFilterToml string

//go:embed complex_es_no_ver.toml
var ExpectedComplexEsNoVerToml string

//go:embed complex_es_v6.toml
var ExpectedComplexEsV6Toml string

//go:embed complex_otel.toml
var ExpectedComplexOTELToml string

//go:embed complex_http_receiver.toml
var ExpectedComplexHTTPReceiverTOML string

//go:embed complex_prune_filter.toml
var ExpectedComplexPruneFilterTOML string

//go:embed container.toml
var ExpectedContainerToml string

// TODO: Use a detailed CLF spec
var _ = Describe("Testing Complete Config Generation", func() {
	defer GinkgoRecover()
	Skip("TODO: enable me after re-wire")
	var (
		f = func(testcase testhelpers.ConfGenerateTest) {
			if testcase.Options == nil {
				testcase.Options = framework.Options{framework.ClusterTLSProfileSpec: tls.GetClusterTLSProfileSpec(nil)}
			}

			Expect(testcase.ExpectedConf).To(EqualConfigFrom(Conf(&testcase.CLSpec, testcase.Secrets, &testcase.CLFSpec, constants.OpenshiftNS, "my-forwarder", &factory.ForwarderResourceNames{CommonName: constants.CollectorName}, testcase.Options)))
		}
	)

	DescribeTable("Generate full vector.toml", f,
		Entry("with spec for containers", testhelpers.ConfGenerateTest{
			Options: framework.Options{
				framework.ClusterTLSProfileSpec: tls.GetClusterTLSProfileSpec(nil),
			},
			CLFSpec: logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{
						Name: "mytestapp",
						Application: &logging.Application{
							Namespaces: []string{"test-ns"},
						},
					},
					{
						Name: "myinfra",
						Infrastructure: &logging.Infrastructure{
							Sources: []string{logging.InfrastructureSourceContainer},
						},
					},
				},
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							"myinfra",
							"mytestapp",
						},
						OutputRefs: []string{"kafka-receiver"},
						Name:       "mypipeline",
						Labels:     map[string]string{"key1": "value1", "key2": "value2"},
					},
				},
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeKafka,
						Name: "kafka-receiver",
						URL:  "tls://broker1-kafka.svc.messaging.cluster.local:9092/topic",
						Secret: &logging.OutputSecretSpec{
							Name: "kafka-receiver-1",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"kafka-receiver": {
					Data: map[string][]byte{
						"tls.key":       []byte("junk"),
						"tls.crt":       []byte("junk"),
						"ca-bundle.crt": []byte("junk"),
					},
				},
			},
			ExpectedConf: ExpectedContainerToml,
		}),
		Entry("with complex spec", testhelpers.ConfGenerateTest{
			Options: framework.Options{
				framework.ClusterTLSProfileSpec: tls.GetClusterTLSProfileSpec(nil),
			},
			CLFSpec: logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{
						Name: "mytestapp",
						Application: &logging.Application{
							Namespaces: []string{"test-ns"},
						},
					},
					{
						Name:           logging.InputNameInfrastructure,
						Infrastructure: &logging.Infrastructure{},
					},
					{
						Name:  logging.InputNameAudit,
						Audit: &logging.Audit{},
					},
				},
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							"mytestapp",
							logging.InputNameInfrastructure,
							logging.InputNameAudit},
						OutputRefs: []string{"kafka-receiver"},
						Name:       "pipeline",
						Labels:     map[string]string{"key1": "value1", "key2": "value2"},
					},
				},
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeKafka,
						Name: "kafka-receiver",
						URL:  "tls://broker1-kafka.svc.messaging.cluster.local:9092/topic",
						Secret: &logging.OutputSecretSpec{
							Name: "kafka-receiver-1",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"kafka-receiver": {
					Data: map[string][]byte{
						"tls.key":       []byte("junk"),
						"tls.crt":       []byte("junk"),
						"ca-bundle.crt": []byte("junk"),
					},
				},
			},
			ExpectedConf: ExpectedComplexToml,
		}),

		Entry("with complex spec for elasticsearch, without version specified", testhelpers.ConfGenerateTest{
			CLSpec: logging.CollectionSpec{},
			CLFSpec: logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{
						Name:        logging.InputNameApplication,
						Application: &logging.Application{},
					},
					{
						Name:           logging.InputNameInfrastructure,
						Infrastructure: &logging.Infrastructure{},
					},
					{
						Name:  logging.InputNameAudit,
						Audit: &logging.Audit{},
					},
				},
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameApplication,
							logging.InputNameInfrastructure,
							logging.InputNameAudit},
						OutputRefs: []string{"es-1", "es-2"},
						Name:       "pipeline",
					},
				},
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeElasticsearch,
						Name: "es-1",
						URL:  "https://es-1.svc.messaging.cluster.local:9200",
						Secret: &logging.OutputSecretSpec{
							Name: "es-1",
						},
					},
					{
						Type: logging.OutputTypeElasticsearch,
						Name: "es-2",
						URL:  "https://es-2.svc.messaging.cluster.local:9200",
						Secret: &logging.OutputSecretSpec{
							Name: "es-2",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"es-1": {
					Data: map[string][]byte{
						"tls.key":       []byte("junk"),
						"tls.crt":       []byte("junk"),
						"ca-bundle.crt": []byte("junk"),
					},
				},
				"es-2": {
					Data: map[string][]byte{
						"tls.key":       []byte("junk"),
						"tls.crt":       []byte("junk"),
						"ca-bundle.crt": []byte("junk"),
					},
				},
			},
			ExpectedConf: ExpectedComplexEsNoVerToml,
		}),

		Entry("with complex spec for elasticsearch default v6 & latest version", testhelpers.ConfGenerateTest{
			Options: framework.Options{
				framework.ClusterTLSProfileSpec: tls.GetClusterTLSProfileSpec(&configv1.TLSSecurityProfile{Type: configv1.TLSProfileIntermediateType}),
			},
			CLSpec: logging.CollectionSpec{},
			CLFSpec: logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{
						Name:        logging.InputNameApplication,
						Application: &logging.Application{},
					},
					{
						Name:           logging.InputNameInfrastructure,
						Infrastructure: &logging.Infrastructure{},
					},
					{
						Name:  logging.InputNameAudit,
						Audit: &logging.Audit{},
					},
				},
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameApplication,
							logging.InputNameInfrastructure,
							logging.InputNameAudit},
						OutputRefs: []string{"default", "es-1", "es-2"},
						Name:       "pipeline",
					},
				},
				Outputs: []logging.OutputSpec{
					{
						Name:   logging.OutputNameDefault,
						Type:   logging.OutputTypeElasticsearch,
						URL:    constants.LogStoreURL,
						Secret: &logging.OutputSecretSpec{Name: constants.CollectorSecretName},
					},
					{
						Type: logging.OutputTypeElasticsearch,
						Name: "es-1",
						URL:  "https://es-1.svc.messaging.cluster.local:9200",
						OutputTypeSpec: logging.OutputTypeSpec{
							Elasticsearch: &logging.Elasticsearch{
								Version: logging.DefaultESVersion,
							},
						},
						Secret: &logging.OutputSecretSpec{
							Name: "es-1",
						},
						TLS: &logging.OutputTLSSpec{
							TLSSecurityProfile: &configv1.TLSSecurityProfile{Type: configv1.TLSProfileIntermediateType},
						},
					},
					{
						Type: logging.OutputTypeElasticsearch,
						Name: "es-2",
						URL:  "https://es-2.svc.messaging.cluster.local:9200",
						OutputTypeSpec: logging.OutputTypeSpec{
							Elasticsearch: &logging.Elasticsearch{
								Version: logging.FirstESVersionWithoutType,
							},
						},
						Secret: &logging.OutputSecretSpec{
							Name: "es-2",
						},
						TLS: &logging.OutputTLSSpec{
							TLSSecurityProfile: &configv1.TLSSecurityProfile{Type: configv1.TLSProfileIntermediateType},
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"es-1": {
					Data: map[string][]byte{
						"tls.key":       []byte("junk"),
						"tls.crt":       []byte("junk"),
						"ca-bundle.crt": []byte("junk"),
					},
				},
				"es-2": {
					Data: map[string][]byte{
						"tls.key":       []byte("junk"),
						"tls.crt":       []byte("junk"),
						"ca-bundle.crt": []byte("junk"),
					},
				},
			},
			ExpectedConf: ExpectedComplexEsV6Toml,
		}),
		Entry("with complex spec & AnnotationSchemaEnabled = 'enabled' & o.HTTP.schema = 'opentelemetry'", testhelpers.ConfGenerateTest{
			Options: framework.Options{
				constants.AnnotationEnableSchema: "true",
			},
			CLSpec: logging.CollectionSpec{},
			CLFSpec: logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{
						Name: "mytestapp",
						Application: &logging.Application{
							Namespaces: []string{"test-ns"},
						},
					},
					{
						Name:           logging.InputNameInfrastructure,
						Infrastructure: &logging.Infrastructure{},
					},
					{
						Name:  logging.InputNameAudit,
						Audit: &logging.Audit{},
					},
				},
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							"mytestapp",
							logging.InputNameInfrastructure,
							logging.InputNameAudit},
						OutputRefs: []string{"http-receiver"},
						Name:       "pipeline",
						Labels:     map[string]string{"key1": "value1", "key2": "value2"},
					},
				},
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeHttp,
						Name: "http-receiver",
						URL:  "https://my-logstore.com",
						Secret: &logging.OutputSecretSpec{
							Name: "http-receiver",
						},
						OutputTypeSpec: logging.OutputTypeSpec{
							Http: &logging.Http{
								Headers: map[string]string{
									"h2": "v2",
									"h1": "v1",
								},
								Method: "POST",
								Schema: constants.OTELSchema,
							},
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"http-receiver": {
					Data: map[string][]byte{
						"username": []byte("username"),
						"password": []byte("password"),
					},
				},
			},
			ExpectedConf: ExpectedComplexOTELToml,
		}),

		Entry("with complex spec && http audit receiver as input source", testhelpers.ConfGenerateTest{
			Options: framework.Options{
				framework.ClusterTLSProfileSpec: tls.GetClusterTLSProfileSpec(nil),
			},
			CLFSpec: logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{
						Name: "mytestapp",
						Application: &logging.Application{
							Namespaces: []string{"test-ns"},
						},
					},
					{
						Name:           logging.InputNameInfrastructure,
						Infrastructure: &logging.Infrastructure{},
					},
					{
						Name:  logging.InputNameAudit,
						Audit: &logging.Audit{},
					},
					{
						Name: "myreceiver",
						Receiver: &logging.ReceiverSpec{
							Type: logging.OutputTypeHttp,
							ReceiverTypeSpec: &logging.ReceiverTypeSpec{
								HTTP: &logging.HTTPReceiver{
									Port:   7777,
									Format: logging.FormatKubeAPIAudit,
								},
							},
						},
					},
				},
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							"mytestapp",
							logging.InputNameInfrastructure,
							logging.InputNameAudit,
							"myreceiver"},
						OutputRefs: []string{"kafka-receiver"},
						Name:       "pipeline",
						Labels:     map[string]string{"key1": "value1", "key2": "value2"},
					},
				},
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeKafka,
						Name: "kafka-receiver",
						URL:  "tls://broker1-kafka.svc.messaging.cluster.local:9092/topic",
						Secret: &logging.OutputSecretSpec{
							Name: "kafka-receiver-1",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"kafka-receiver": {
					Data: map[string][]byte{
						"tls.key":       []byte("junk"),
						"tls.crt":       []byte("junk"),
						"ca-bundle.crt": []byte("junk"),
					},
				},
			},
			ExpectedConf: ExpectedComplexHTTPReceiverTOML,
		}),
		Entry("with drop filters", testhelpers.ConfGenerateTest{
			Options: framework.Options{
				framework.ClusterTLSProfileSpec: tls.GetClusterTLSProfileSpec(nil),
			},
			CLFSpec: logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{
						Name: "mytestapp",
						Application: &logging.Application{
							Namespaces: []string{"test-ns"},
						},
					},
					{
						Name:           logging.InputNameInfrastructure,
						Infrastructure: &logging.Infrastructure{},
					},
					{
						Name:  logging.InputNameAudit,
						Audit: &logging.Audit{},
					},
				},
				Filters: []logging.FilterSpec{
					{
						Name: "drop-test",
						Type: logging.FilterDrop,
						FilterTypeSpec: logging.FilterTypeSpec{
							DropTestsSpec: &[]logging.DropTest{
								{
									DropConditions: []logging.DropCondition{
										{
											Field:   ".kubernetes.namespace_name",
											Matches: "busybox",
										},
										{
											Field:      ".level",
											NotMatches: "d.+",
										},
									},
								},
								{
									DropConditions: []logging.DropCondition{
										{
											Field:   ".log_type",
											Matches: "application",
										},
									},
								},
								{
									DropConditions: []logging.DropCondition{
										{
											Field:   ".kubernetes.container_name",
											Matches: "error|warning",
										},
										{
											Field:      ".kubernetes.labels.test",
											NotMatches: "foo",
										},
									},
								},
							},
						},
					},
				},
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							"mytestapp",
							logging.InputNameInfrastructure,
							logging.InputNameAudit},
						OutputRefs: []string{"kafka-receiver"},
						Name:       "pipeline",
						Labels:     map[string]string{"key1": "value1", "key2": "value2"},
						FilterRefs: []string{"drop-test"},
					},
				},
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeKafka,
						Name: "kafka-receiver",
						URL:  "tls://broker1-kafka.svc.messaging.cluster.local:9092/topic",
						Secret: &logging.OutputSecretSpec{
							Name: "kafka-receiver-1",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"kafka-receiver": {
					Data: map[string][]byte{
						"tls.key":       []byte("junk"),
						"tls.crt":       []byte("junk"),
						"ca-bundle.crt": []byte("junk"),
					},
				},
			},
			ExpectedConf: ExpectedComplexDropFilterToml,
		}),

		Entry("with complex spec with prune filter", testhelpers.ConfGenerateTest{
			Options: framework.Options{
				framework.ClusterTLSProfileSpec: tls.GetClusterTLSProfileSpec(nil),
			},
			CLFSpec: logging.ClusterLogForwarderSpec{
				Filters: []logging.FilterSpec{
					{
						Name: "my-prune",
						Type: logging.FilterPrune,
						FilterTypeSpec: logging.FilterTypeSpec{
							PruneFilterSpec: &logging.PruneFilterSpec{
								In:    []string{".log_type", ".message", ".kubernetes.container_name"},
								NotIn: []string{`.kubernetes.labels."foo-bar/baz"`, ".level"},
							},
						},
					},
				},
				Inputs: []logging.InputSpec{
					{
						Name: "mytestapp",
						Application: &logging.Application{
							Namespaces: []string{"test-ns"},
						},
					},
					{
						Name:           logging.InputNameInfrastructure,
						Infrastructure: &logging.Infrastructure{},
					},
					{
						Name:  logging.InputNameAudit,
						Audit: &logging.Audit{},
					},
				},
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							"mytestapp",
							logging.InputNameInfrastructure,
							logging.InputNameAudit},
						OutputRefs: []string{"kafka-receiver"},
						Name:       "pipeline",
						Labels:     map[string]string{"key1": "value1", "key2": "value2"},
						FilterRefs: []string{"my-prune"},
					},
				},
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeKafka,
						Name: "kafka-receiver",
						URL:  "tls://broker1-kafka.svc.messaging.cluster.local:9092/topic",
						Secret: &logging.OutputSecretSpec{
							Name: "kafka-receiver-1",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"kafka-receiver": {
					Data: map[string][]byte{
						"tls.key":       []byte("junk"),
						"tls.crt":       []byte("junk"),
						"ca-bundle.crt": []byte("junk"),
					},
				},
			},
			ExpectedConf: ExpectedComplexPruneFilterTOML,
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
