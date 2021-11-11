package elasticsearch

import (
	"testing"

	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/security"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Generate Vector config", func() {
	inputPipeline := []string{"application"}
	var f = func(clspec logging.ClusterLoggingSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		return Conf(clfspec.Outputs[0], inputPipeline, secrets[clfspec.Outputs[0].Name], generator.NoOptions)
	}
	DescribeTable("For Elasticsearch output", generator.TestGenerateConfWith(f),
		Entry("with username,password", generator.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeElasticsearch,
						Name: "es-1",
						URL:  "https://es.svc.infra.cluster:9200",
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
inputs = ["application"]
endpoint = "https://es.svc.infra.cluster:9200"
index = "{{ log_type }}-write"
request.timeout_secs = 2147483648
bulk_action = "create"
# Basic Auth Config
[sinks.es_1.auth]
strategy = "basic"
user = ""
password = ""
`,
		}),
		Entry("with tls key,cert,ca-bundle", generator.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeElasticsearch,
						Name: "es-1",
						URL:  "https://es.svc.infra.cluster:9200",
						Secret: &logging.OutputSecretSpec{
							Name: "es-1",
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
			},
			ExpectedConf: `
[sinks.es_1]
type = "elasticsearch"
inputs = ["application"]
endpoint = "https://es.svc.infra.cluster:9200"
index = "{{ log_type }}-write"
request.timeout_secs = 2147483648
bulk_action = "create"
# TLS Config
[sinks.es_1.tls]
key_file = "/var/run/ocp-collector/secrets/es-1/tls.key"
crt_file = "/var/run/ocp-collector/secrets/es-1/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/es-1/ca-bundle.crt"
`,
		}),
		Entry("without security", generator.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type:   logging.OutputTypeElasticsearch,
						Name:   "es-1",
						URL:    "http://es.svc.infra.cluster:9200",
						Secret: nil,
					},
				},
			},
			Secrets: security.NoSecrets,
			ExpectedConf: `
[sinks.es_1]
type = "elasticsearch"
inputs = ["application"]
endpoint = "http://es.svc.infra.cluster:9200"
index = "{{ log_type }}-write"
request.timeout_secs = 2147483648
bulk_action = "create"
`,
		}),
	)
})

func TestVectorConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vector Conf Generation")
}
