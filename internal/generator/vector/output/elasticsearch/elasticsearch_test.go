package elasticsearch

import (
	"testing"

	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Generate Vector config", func() {
	inputPipeline := []string{"transform_application"}
	var f = func(clspec logging.ClusterLoggingSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		return Conf(secrets[clfspec.Outputs[0].Name], clfspec.Outputs[0], generator.NoOptions, inputPipeline)
	}
	DescribeTable("for Elasticsearch output", generator.TestGenerateConfWith(f),
		Entry("with username,password", generator.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeElasticsearch,
						Name: "es-1",
						URL:  "https://es.svc.infra.cluster:9999",
						Secret: &logging.OutputSecretSpec{
							Name: "es-1",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"es-1": {
					Data: map[string][]byte{
						"username": []byte("junk"),
						"password": []byte("junk"),
					},
				},
			},
			ExpectedConf: `
[sinks.es_1]
  type = "elasticsearch"
  inputs = ["transform_application"]
  endpoint = "es.svc.infra.cluster:9999"
  mode = "normal"
  pipeline = "pipeline-name"
  compression = "none"
  auth.strategy = "basic"
auth.user = "#{File.exists?('/var/run/ocp-collector/secrets/es-1/username') ? open('/var/run/ocp-collector/secrets/es-1/username','r') do |f|f.read end : ''}"
auth.password = "#{File.exists?('/var/run/ocp-collector/secrets/es-1/password') ? open('/var/run/ocp-collector/secrets/es-1/password','r') do |f|f.read end : ''}"
`,
		}),
		Entry("with tls key,cert,ca-bundle", generator.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeElasticsearch,
						Name: "es-1",
						URL:  "https://es.svc.infra.cluster:9999",
						Secret: &logging.OutputSecretSpec{
							Name: "es-1",
						},
					},
				},
			},
			Secrets: map[string]*corev1.Secret{
				"es-1": {
					Data: map[string][]byte{
						"tls.key": []byte("junk"),
						"tls.crt": []byte("junk"),
					},
				},
			},
			ExpectedConf: `
[sinks.es_1]
  type = "elasticsearch"
  inputs = ["transform_application"]
  endpoint = "es.svc.infra.cluster:9999"
  mode = "normal"
  pipeline = "pipeline-name"
  compression = "none"
  tls.key_file = '/var/run/ocp-collector/secrets/es-1/tls.key'
tls.crt_file = '/var/run/ocp-collector/secrets/es-1/tls.crt'
`,
		}),
		Entry("without security", generator.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type:   logging.OutputTypeElasticsearch,
						Name:   "es-1",
						URL:    "http://es.svc.infra.cluster:9999",
						Secret: nil,
					},
				},
			},
			Secrets: security.NoSecrets,
			ExpectedConf: `
[sinks.es_1]
  type = "elasticsearch"
  inputs = ["transform_application"]
  endpoint = "es.svc.infra.cluster:9999"
  mode = "normal"
  pipeline = "pipeline-name"
  compression = "none"
`,
		}),
	)
})

func TestFluendConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fluend Conf Generation")
}
