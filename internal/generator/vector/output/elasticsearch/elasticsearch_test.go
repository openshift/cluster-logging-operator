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
		return Conf(clfspec.Outputs[0], inputPipeline, secrets[clfspec.Outputs[0].Name], op)
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
						"username": []byte("testuser"),
						"password": []byte("testpass"),
					},
				},
			},
			ExpectedConf: `
# Adding _id field
[transforms.es_1_add_es_id]
type = "remap"
inputs = ["application"]
source = """
  index = "default"
  if (.log_type == "application"){
    index = "app"
  }
  if (.log_type == "infrastructure"){
    index = "infra"
  }
  if (.log_type == "audit"){
    index = "audit"
  }
  ."write-index"=index+"-write"
  ._id = encode_base64(uuid_v4())
"""

[transforms.es_1_dedot_and_flatten]
type = "lua"
inputs = ["es_1_add_es_id"]
version = "2"
hooks.process = "process"
source = """
    function process(event, emit)
        if event.log.kubernetes == nil then
            return
        end
        dedot(event.log.kubernetes.pod_labels)
        -- create "flat_labels" key
        event.log.kubernetes.flat_labels = {}
        i = 1
        -- flatten the labels
        for k,v in pairs(event.log.kubernetes.pod_labels) do
          event.log.kubernetes.flat_labels[i] = k.."="..v
          i=i+1
        end
        -- delete the "pod_labels" key
        event.log.kubernetes["pod_labels"] = nil
        emit(event)
    end

    function dedot(map)
        if map == nil then
            return
        end
        local new_map = {}
        local changed_keys = {}
        for k, v in pairs(map) do
            local dedotted = string.gsub(k, "%.", "_")
            if dedotted ~= k then
                new_map[dedotted] = v
                changed_keys[k] = true
            end
        end
        for k in pairs(changed_keys) do
            map[k] = nil
        end
        for k, v in pairs(new_map) do
            map[k] = v
        end
    end
"""

[sinks.es_1]
type = "elasticsearch"
inputs = ["es_1_dedot_and_flatten"]
endpoint = "https://es.svc.infra.cluster:9200"
index = "{{ write-index }}"
request.timeout_secs = 2147483648
bulk_action = "create"
id_key = "_id"

# Basic Auth Config
[sinks.es_1.auth]
strategy = "basic"
user = "testuser"
password = "testpass"
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
# Adding _id field
[transforms.es_1_add_es_id]
type = "remap"
inputs = ["application"]
source = """
  index = "default"
  if (.log_type == "application"){
    index = "app"
  }
  if (.log_type == "infrastructure"){
    index = "infra"
  }
  if (.log_type == "audit"){
    index = "audit"
  }
  ."write-index"=index+"-write"
  ._id = encode_base64(uuid_v4())
"""

[transforms.es_1_dedot_and_flatten]
type = "lua"
inputs = ["es_1_add_es_id"]
version = "2"
hooks.process = "process"
source = """
    function process(event, emit)
        if event.log.kubernetes == nil then
            return
        end
        dedot(event.log.kubernetes.pod_labels)
        -- create "flat_labels" key
        event.log.kubernetes.flat_labels = {}
        i = 1
        -- flatten the labels
        for k,v in pairs(event.log.kubernetes.pod_labels) do
          event.log.kubernetes.flat_labels[i] = k.."="..v
          i=i+1
        end
        -- delete the "pod_labels" key
        event.log.kubernetes["pod_labels"] = nil
        emit(event)
    end

    function dedot(map)
        if map == nil then
            return
        end
        local new_map = {}
        local changed_keys = {}
        for k, v in pairs(map) do
            local dedotted = string.gsub(k, "%.", "_")
            if dedotted ~= k then
                new_map[dedotted] = v
                changed_keys[k] = true
            end
        end
        for k in pairs(changed_keys) do
            map[k] = nil
        end
        for k, v in pairs(new_map) do
            map[k] = v
        end
    end
"""

[sinks.es_1]
type = "elasticsearch"
inputs = ["es_1_dedot_and_flatten"]
endpoint = "https://es.svc.infra.cluster:9200"
index = "{{ write-index }}"
request.timeout_secs = 2147483648
bulk_action = "create"
id_key = "_id"

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
# Adding _id field
[transforms.es_1_add_es_id]
type = "remap"
inputs = ["application"]
source = """
  index = "default"
  if (.log_type == "application"){
    index = "app"
  }
  if (.log_type == "infrastructure"){
    index = "infra"
  }
  if (.log_type == "audit"){
    index = "audit"
  }
  ."write-index"=index+"-write"
  ._id = encode_base64(uuid_v4())
"""

[transforms.es_1_dedot_and_flatten]
type = "lua"
inputs = ["es_1_add_es_id"]
version = "2"
hooks.process = "process"
source = """
    function process(event, emit)
        if event.log.kubernetes == nil then
            return
        end
        dedot(event.log.kubernetes.pod_labels)
        -- create "flat_labels" key
        event.log.kubernetes.flat_labels = {}
        i = 1
        -- flatten the labels
        for k,v in pairs(event.log.kubernetes.pod_labels) do
          event.log.kubernetes.flat_labels[i] = k.."="..v
          i=i+1
        end
        -- delete the "pod_labels" key
        event.log.kubernetes["pod_labels"] = nil
        emit(event)
    end

    function dedot(map)
        if map == nil then
            return
        end
        local new_map = {}
        local changed_keys = {}
        for k, v in pairs(map) do
            local dedotted = string.gsub(k, "%.", "_")
            if dedotted ~= k then
                new_map[dedotted] = v
                changed_keys[k] = true
            end
        end
        for k in pairs(changed_keys) do
            map[k] = nil
        end
        for k, v in pairs(new_map) do
            map[k] = v
        end
    end
"""

[sinks.es_1]
type = "elasticsearch"
inputs = ["es_1_dedot_and_flatten"]
endpoint = "http://es.svc.infra.cluster:9200"
index = "{{ write-index }}"
request.timeout_secs = 2147483648
bulk_action = "create"
id_key = "_id"
`,
		}),
	)
})

func TestVectorConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vector Conf Generation")
}
