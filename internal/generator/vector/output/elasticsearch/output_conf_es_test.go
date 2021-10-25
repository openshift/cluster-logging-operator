package elasticsearch_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/elasticsearch"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Generating vector config blocks", func() {

	var (
		outputs  []logging.OutputSpec
		pipeline logging.PipelineSpec
		clfspec  logging.ClusterLogForwarderSpec
		g        generator.Generator
		secrets  map[string]*corev1.Secret
	)
	BeforeEach(func() {
		g = generator.MakeGenerator()
	})

	Context("for a secure endpoint", func() {
		BeforeEach(func() {
			outputs = []logging.OutputSpec{
				{
					Type: logging.OutputTypeElasticsearch,
					Name: "oncluster-elasticsearch",
					URL:  "https://es.svc.messaging.cluster.local:9654",
					Secret: &logging.OutputSecretSpec{
						Name: "my-es-secret",
					},
				},
			}
			pipeline = logging.PipelineSpec{
				Name:       "my-secure-pipeline",
				InputRefs:  []string{logging.InputNameApplication},
				OutputRefs: []string{"oncluster-elasticsearch"},
			}
			secrets = map[string]*corev1.Secret{
				"oncluster-elasticsearch": {
					ObjectMeta: v1.ObjectMeta{
						Name: "my-es-secret",
					},
					Data: map[string][]byte{
						"tls.key":       []byte("test-key"),
						"tls.crt":       []byte("test-crt"),
						"ca-bundle.crt": []byte("test-bundle"),
					},
				},
			}
		})

		PIt("should produce well formed configuration with input application and output as elastic search", func() {
			inputPipeline := []string{"transform_application"}
			pipeline.OutputRefs = append(pipeline.OutputRefs, "other-elasticsearch")
			clfspec.Pipelines = []logging.PipelineSpec{pipeline}
			clfspec.Outputs = outputs
			g := generator.MakeGenerator()
			e := elasticsearch.Conf(nil, clfspec.Outputs[0], generator.NoOptions, inputPipeline)
			results, err := g.GenerateConf(e...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(`
    [sinks.oncluster_elasticsearch]
      type = "elasticsearch"
      inputs = ["transform_application"]
      endpoint = "es.svc.messaging.cluster.local:9654"
      mode = "normal"
      pipeline = "pipeline-name"
      compression = "none"
`))
		})
		It("should produce well formed output label config with username/password", func() {
			inputPipeline := []string{"transform_application"}
			results, err := g.GenerateConf(elasticsearch.Conf(secrets[outputs[0].Name], outputs[0], nil, inputPipeline)...)
			Expect(err).To(BeNil())
			Expect(results).To(EqualTrimLines(`
  [sinks.oncluster_elasticsearch]
  type = "elasticsearch"
  inputs = ["transform_application"]
  endpoint = "es.svc.messaging.cluster.local:9654"
  mode = "normal"
  pipeline = "pipeline-name"
  compression = "none"
  tls.key_file = '/var/run/ocp-collector/secrets/my-es-secret/tls.key'
tls.crt_file = '/var/run/ocp-collector/secrets/my-es-secret/tls.crt'
tls.ca_file = '/var/run/ocp-collector/secrets/my-es-secret/ca-bundle.crt'
`))
		})
	})
})
