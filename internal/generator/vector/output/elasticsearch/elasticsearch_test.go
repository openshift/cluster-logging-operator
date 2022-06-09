package elasticsearch

import (
	"github.com/openshift/cluster-logging-operator/test/helpers"
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
		e := []generator.Element{}
		for _, o := range clfspec.Outputs {
			e = generator.MergeElements(e, Conf(o, inputPipeline, secrets[o.Name], op))
		}
		return e
	}
	DescribeTable("For Elasticsearch output", helpers.TestGenerateConfWith(f),
		Entry("with username,password", helpers.ConfGenerateTest{
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
# Set Elasticsearch index
[transforms.es_1_add_es_index]
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
  .write_index = index + "-write"
  ._id = encode_base64(uuid_v4())
  del(.file)
  del(.tag)
  del(.source_type)
"""

[transforms.es_1_dedot_and_flatten]
type = "lua"
inputs = ["es_1_add_es_index"]
version = "2"
hooks.process = "process"
source = """
    function process(event, emit)
        if event.log.kubernetes == nil then
            emit(event)
            return
        end
        if event.log.kubernetes.labels == nil then
            emit(event)
            return
        end
        dedot(event.log.kubernetes.labels)
        flatten_labels(event)
        prune_labels(event)
        emit(event)
    end

    function flatten_labels(event)
        -- create "flat_labels" key
        event.log.kubernetes.flat_labels = {}
        i = 1
        -- flatten the labels
        for k,v in pairs(event.log.kubernetes.labels) do
          event.log.kubernetes.flat_labels[i] = k.."="..v
          i=i+1
        end
    end 

	function prune_labels(event)
		local exclusions = {"app_kubernetes_io/name", "app_kubernetes_io/instance", "app_kubernetes_io/version", "app_kubernetes_io/component", "app_kubernetes_io/part-of", "app_kubernetes_io/managed-by", "app_kubernetes_io/created-by"}
		local keys = {}
		for k,v in pairs(event.log.kubernetes.labels) do
			for index, e in pairs(exclusions) do
				if k == e then
					keys[k] = v
				end
			end
		end
		event.log.kubernetes.labels = keys
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
bulk.index = "{{ write_index }}"
bulk.action = "create"
request.timeout_secs = 2147483648
id_key = "_id"

# Basic Auth Config
[sinks.es_1.auth]
strategy = "basic"
user = "testuser"
password = "testpass"
`,
		}),
		Entry("with tls key,cert,ca-bundle", helpers.ConfGenerateTest{
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
# Set Elasticsearch index
[transforms.es_1_add_es_index]
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
  .write_index = index + "-write"
  ._id = encode_base64(uuid_v4())
  del(.file)
  del(.tag)
  del(.source_type)
"""

[transforms.es_1_dedot_and_flatten]
type = "lua"
inputs = ["es_1_add_es_index"]
version = "2"
hooks.process = "process"
source = """
    function process(event, emit)
        if event.log.kubernetes == nil then
            emit(event)
            return
        end
        if event.log.kubernetes.labels == nil then
            emit(event)
            return
        end
        dedot(event.log.kubernetes.labels)
        flatten_labels(event)
        prune_labels(event)
        emit(event)
    end

    function flatten_labels(event)
        -- create "flat_labels" key
        event.log.kubernetes.flat_labels = {}
        i = 1
        -- flatten the labels
        for k,v in pairs(event.log.kubernetes.labels) do
          event.log.kubernetes.flat_labels[i] = k.."="..v
          i=i+1
        end
    end 

	function prune_labels(event)
		local exclusions = {"app_kubernetes_io/name", "app_kubernetes_io/instance", "app_kubernetes_io/version", "app_kubernetes_io/component", "app_kubernetes_io/part-of", "app_kubernetes_io/managed-by", "app_kubernetes_io/created-by"}
		local keys = {}
		for k,v in pairs(event.log.kubernetes.labels) do
			for index, e in pairs(exclusions) do
				if k == e then
					keys[k] = v
				end
			end
		end
		event.log.kubernetes.labels = keys
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
bulk.index = "{{ write_index }}"
bulk.action = "create"
request.timeout_secs = 2147483648
id_key = "_id"

# TLS Config
[sinks.es_1.tls]
enabled = true
key_file = "/var/run/ocp-collector/secrets/es-1/tls.key"
crt_file = "/var/run/ocp-collector/secrets/es-1/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/es-1/ca-bundle.crt"
`,
		}),
		Entry("without security", helpers.ConfGenerateTest{
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
# Set Elasticsearch index
[transforms.es_1_add_es_index]
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
  .write_index = index + "-write"
  ._id = encode_base64(uuid_v4())
  del(.file)
  del(.tag)
  del(.source_type)
"""

[transforms.es_1_dedot_and_flatten]
type = "lua"
inputs = ["es_1_add_es_index"]
version = "2"
hooks.process = "process"
source = """
    function process(event, emit)
        if event.log.kubernetes == nil then
            emit(event)
            return
        end
        if event.log.kubernetes.labels == nil then
            emit(event)
            return
        end
        dedot(event.log.kubernetes.labels)
        flatten_labels(event)
        prune_labels(event)
        emit(event)
    end

    function flatten_labels(event)
        -- create "flat_labels" key
        event.log.kubernetes.flat_labels = {}
        i = 1
        -- flatten the labels
        for k,v in pairs(event.log.kubernetes.labels) do
          event.log.kubernetes.flat_labels[i] = k.."="..v
          i=i+1
        end
    end 

	function prune_labels(event)
		local exclusions = {"app_kubernetes_io/name", "app_kubernetes_io/instance", "app_kubernetes_io/version", "app_kubernetes_io/component", "app_kubernetes_io/part-of", "app_kubernetes_io/managed-by", "app_kubernetes_io/created-by"}
		local keys = {}
		for k,v in pairs(event.log.kubernetes.labels) do
			for index, e in pairs(exclusions) do
				if k == e then
					keys[k] = v
				end
			end
		end
		event.log.kubernetes.labels = keys
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
bulk.index = "{{ write_index }}"
bulk.action = "create"
request.timeout_secs = 2147483648
id_key = "_id"
`,
		}),
		Entry("with multiple pipelines for elastic-search", helpers.ConfGenerateTest{
			CLSpec: logging.ClusterLoggingSpec{
				Forwarder: &logging.ForwarderSpec{},
			},
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameApplication,
							logging.InputNameInfrastructure,
						},
						OutputRefs: []string{"es-1"},
						Name:       "pipeline1",
					},
					{
						InputRefs: []string{
							logging.InputNameAudit},
						OutputRefs: []string{"es-1", "es-2"},
						Name:       "pipeline2",
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
			ExpectedConf: `
# Set Elasticsearch index
[transforms.es_1_add_es_index]
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
  .write_index = index + "-write"
  ._id = encode_base64(uuid_v4())
  del(.file)
  del(.tag)
  del(.source_type)
"""

[transforms.es_1_dedot_and_flatten]
type = "lua"
inputs = ["es_1_add_es_index"]
version = "2"
hooks.process = "process"
source = """
    function process(event, emit)
        if event.log.kubernetes == nil then
            emit(event)
            return
        end
        if event.log.kubernetes.labels == nil then
            emit(event)
            return
        end
        dedot(event.log.kubernetes.labels)
        flatten_labels(event)
        prune_labels(event)
        emit(event)
    end

    function flatten_labels(event)
        -- create "flat_labels" key
        event.log.kubernetes.flat_labels = {}
        i = 1
        -- flatten the labels
        for k,v in pairs(event.log.kubernetes.labels) do
          event.log.kubernetes.flat_labels[i] = k.."="..v
          i=i+1
        end
    end 

	function prune_labels(event)
		local exclusions = {"app_kubernetes_io/name", "app_kubernetes_io/instance", "app_kubernetes_io/version", "app_kubernetes_io/component", "app_kubernetes_io/part-of", "app_kubernetes_io/managed-by", "app_kubernetes_io/created-by"}
		local keys = {}
		for k,v in pairs(event.log.kubernetes.labels) do
			for index, e in pairs(exclusions) do
				if k == e then
					keys[k] = v
				end
			end
		end
		event.log.kubernetes.labels = keys
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
endpoint = "https://es-1.svc.messaging.cluster.local:9200"
bulk.index = "{{ write_index }}"
bulk.action = "create"
request.timeout_secs = 2147483648
id_key = "_id"

# TLS Config
[sinks.es_1.tls]
enabled = true
key_file = "/var/run/ocp-collector/secrets/es-1/tls.key"
crt_file = "/var/run/ocp-collector/secrets/es-1/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/es-1/ca-bundle.crt"

# Set Elasticsearch index
[transforms.es_2_add_es_index]
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
  .write_index = index + "-write"
  ._id = encode_base64(uuid_v4())
  del(.file)
  del(.tag)
  del(.source_type)
"""

[transforms.es_2_dedot_and_flatten]
type = "lua"
inputs = ["es_2_add_es_index"]
version = "2"
hooks.process = "process"
source = """
    function process(event, emit)
        if event.log.kubernetes == nil then
            emit(event)
            return
        end
        if event.log.kubernetes.labels == nil then
            emit(event)
            return
        end
        dedot(event.log.kubernetes.labels)
        flatten_labels(event)
        prune_labels(event)
        emit(event)
    end

    function flatten_labels(event)
        -- create "flat_labels" key
        event.log.kubernetes.flat_labels = {}
        i = 1
        -- flatten the labels
        for k,v in pairs(event.log.kubernetes.labels) do
          event.log.kubernetes.flat_labels[i] = k.."="..v
          i=i+1
        end
    end 

	function prune_labels(event)
		local exclusions = {"app_kubernetes_io/name", "app_kubernetes_io/instance", "app_kubernetes_io/version", "app_kubernetes_io/component", "app_kubernetes_io/part-of", "app_kubernetes_io/managed-by", "app_kubernetes_io/created-by"}
		local keys = {}
		for k,v in pairs(event.log.kubernetes.labels) do
			for index, e in pairs(exclusions) do
				if k == e then
					keys[k] = v
				end
			end
		end
		event.log.kubernetes.labels = keys
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

[sinks.es_2]
type = "elasticsearch"
inputs = ["es_2_dedot_and_flatten"]
endpoint = "https://es-2.svc.messaging.cluster.local:9200"
bulk.index = "{{ write_index }}"
bulk.action = "create"
request.timeout_secs = 2147483648
id_key = "_id"

# TLS Config
[sinks.es_2.tls]
enabled = true
key_file = "/var/run/ocp-collector/secrets/es-2/tls.key"
crt_file = "/var/run/ocp-collector/secrets/es-2/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/es-2/ca-bundle.crt"
`,
		}),
		Entry("with StructuredTypeKey", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type:   logging.OutputTypeElasticsearch,
						Name:   "es-1",
						URL:    "http://es.svc.infra.cluster:9200",
						Secret: nil,
						OutputTypeSpec: logging.OutputTypeSpec{
							Elasticsearch: &logging.Elasticsearch{
								StructuredTypeKey: "kubernetes.labels.mylabel",
							},
						},
					},
				},
			},
			Secrets: security.NoSecrets,
			ExpectedConf: `
# Set Elasticsearch index
[transforms.es_1_add_es_index]
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
  .write_index = index + "-write"
  ._id = encode_base64(uuid_v4())
  del(.file)
  del(.tag)
  del(.source_type)
  if .log_type == "application" && .structured != null {
    val = .kubernetes.labels.mylabel
    if val != null {
      .write_index, err = "app-" + val + "-write"
    }
  }
"""

[transforms.es_1_dedot_and_flatten]
type = "lua"
inputs = ["es_1_add_es_index"]
version = "2"
hooks.process = "process"
source = """
    function process(event, emit)
        if event.log.kubernetes == nil then
            emit(event)
            return
        end
        if event.log.kubernetes.labels == nil then
            emit(event)
            return
        end
        dedot(event.log.kubernetes.labels)
        flatten_labels(event)
        prune_labels(event)
        emit(event)
    end

    function flatten_labels(event)
        -- create "flat_labels" key
        event.log.kubernetes.flat_labels = {}
        i = 1
        -- flatten the labels
        for k,v in pairs(event.log.kubernetes.labels) do
          event.log.kubernetes.flat_labels[i] = k.."="..v
          i=i+1
        end
    end 

	function prune_labels(event)
		local exclusions = {"app_kubernetes_io/name", "app_kubernetes_io/instance", "app_kubernetes_io/version", "app_kubernetes_io/component", "app_kubernetes_io/part-of", "app_kubernetes_io/managed-by", "app_kubernetes_io/created-by"}
		local keys = {}
		for k,v in pairs(event.log.kubernetes.labels) do
			for index, e in pairs(exclusions) do
				if k == e then
					keys[k] = v
				end
			end
		end
		event.log.kubernetes.labels = keys
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
bulk.index = "{{ write_index }}"
bulk.action = "create"
request.timeout_secs = 2147483648
id_key = "_id"
`,
		}),
		Entry("with StructuredTypeName", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type:   logging.OutputTypeElasticsearch,
						Name:   "es-1",
						URL:    "http://es.svc.infra.cluster:9200",
						Secret: nil,
						OutputTypeSpec: logging.OutputTypeSpec{
							Elasticsearch: &logging.Elasticsearch{
								StructuredTypeName: "myindex",
							},
						},
					},
				},
			},
			Secrets: security.NoSecrets,
			ExpectedConf: `
# Set Elasticsearch index
[transforms.es_1_add_es_index]
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
  .write_index = index + "-write"
  ._id = encode_base64(uuid_v4())
  del(.file)
  del(.tag)
  del(.source_type)
  if .log_type == "application" && .structured != null {
    .write_index = "app-myindex-write"
  }
"""

[transforms.es_1_dedot_and_flatten]
type = "lua"
inputs = ["es_1_add_es_index"]
version = "2"
hooks.process = "process"
source = """
    function process(event, emit)
        if event.log.kubernetes == nil then
            emit(event)
            return
        end
        if event.log.kubernetes.labels == nil then
            emit(event)
            return
        end
        dedot(event.log.kubernetes.labels)
        flatten_labels(event)
        prune_labels(event)
        emit(event)
    end

    function flatten_labels(event)
        -- create "flat_labels" key
        event.log.kubernetes.flat_labels = {}
        i = 1
        -- flatten the labels
        for k,v in pairs(event.log.kubernetes.labels) do
          event.log.kubernetes.flat_labels[i] = k.."="..v
          i=i+1
        end
    end 

	function prune_labels(event)
		local exclusions = {"app_kubernetes_io/name", "app_kubernetes_io/instance", "app_kubernetes_io/version", "app_kubernetes_io/component", "app_kubernetes_io/part-of", "app_kubernetes_io/managed-by", "app_kubernetes_io/created-by"}
		local keys = {}
		for k,v in pairs(event.log.kubernetes.labels) do
			for index, e in pairs(exclusions) do
				if k == e then
					keys[k] = v
				end
			end
		end
		event.log.kubernetes.labels = keys
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
bulk.index = "{{ write_index }}"
bulk.action = "create"
request.timeout_secs = 2147483648
id_key = "_id"
`,
		}),
		Entry("with both StructuredTypeKey and StructuredTypeName", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type:   logging.OutputTypeElasticsearch,
						Name:   "es-1",
						URL:    "http://es.svc.infra.cluster:9200",
						Secret: nil,
						OutputTypeSpec: logging.OutputTypeSpec{
							Elasticsearch: &logging.Elasticsearch{
								StructuredTypeKey:  "kubernetes.labels.mylabel",
								StructuredTypeName: "myindex",
							},
						},
					},
				},
			},
			Secrets: security.NoSecrets,
			ExpectedConf: `
# Set Elasticsearch index
[transforms.es_1_add_es_index]
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
  .write_index = index + "-write"
  ._id = encode_base64(uuid_v4())
  del(.file)
  del(.tag)
  del(.source_type)
  if .log_type == "application" && .structured != null {
    val = .kubernetes.labels.mylabel
    if val != null {
      .write_index, err = "app-" + val + "-write"
    } else {
      .write_index = "app-myindex-write"
    }
  }
"""

[transforms.es_1_dedot_and_flatten]
type = "lua"
inputs = ["es_1_add_es_index"]
version = "2"
hooks.process = "process"
source = """
    function process(event, emit)
        if event.log.kubernetes == nil then
            emit(event)
            return
        end
        if event.log.kubernetes.labels == nil then
            emit(event)
            return
        end
        dedot(event.log.kubernetes.labels)
        flatten_labels(event)
        prune_labels(event)
        emit(event)
    end

    function flatten_labels(event)
        -- create "flat_labels" key
        event.log.kubernetes.flat_labels = {}
        i = 1
        -- flatten the labels
        for k,v in pairs(event.log.kubernetes.labels) do
          event.log.kubernetes.flat_labels[i] = k.."="..v
          i=i+1
        end
    end 

	function prune_labels(event)
		local exclusions = {"app_kubernetes_io/name", "app_kubernetes_io/instance", "app_kubernetes_io/version", "app_kubernetes_io/component", "app_kubernetes_io/part-of", "app_kubernetes_io/managed-by", "app_kubernetes_io/created-by"}
		local keys = {}
		for k,v in pairs(event.log.kubernetes.labels) do
			for index, e in pairs(exclusions) do
				if k == e then
					keys[k] = v
				end
			end
		end
		event.log.kubernetes.labels = keys
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
bulk.index = "{{ write_index }}"
bulk.action = "create"
request.timeout_secs = 2147483648
id_key = "_id"
`,
		}),
		Entry("with StructuredTypeKey, StructuredTypeName, container annotations enabled", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type:   logging.OutputTypeElasticsearch,
						Name:   "es-1",
						URL:    "http://es.svc.infra.cluster:9200",
						Secret: nil,
						OutputTypeSpec: logging.OutputTypeSpec{
							Elasticsearch: &logging.Elasticsearch{
								StructuredTypeKey:             "kubernetes.labels.mylabel",
								StructuredTypeName:            "myindex",
								EnableStructuredContainerLogs: true,
							},
						},
					},
				},
			},
			Secrets: security.NoSecrets,
			ExpectedConf: `
# Set Elasticsearch index
[transforms.es_1_add_es_index]
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
  .write_index = index + "-write"
  
  ._id = encode_base64(uuid_v4())
  del(.file)
  del(.tag)
  del(.source_type)
  if .log_type == "application" && .structured != null {
    val = .kubernetes.labels.mylabel
    if val != null {
      .write_index, err = "app-" + val + "-write"
    } else {
      .write_index = "app-myindex-write"
    }
  }

  if .log_type == "application"  && .structured != null && .kubernetes.container_name != null && .kubernetes.annotations != null && length!(.kubernetes.annotations) > 0{
	key = join!(["containerType.logging.openshift.io", .kubernetes.container_name], separator: "/")
    index, err = get(value: .kubernetes.annotations, path: [key])
    if index != null && err == null {
      .write_index = join!(["app-",index,"-write"])
    } else {
       log(err, level: "error")
    }
  }
"""

[transforms.es_1_dedot_and_flatten]
type = "lua"
inputs = ["es_1_add_es_index"]
version = "2"
hooks.process = "process"
source = """
    function process(event, emit)
        if event.log.kubernetes == nil then
            emit(event)
            return
        end
        if event.log.kubernetes.labels == nil then
            emit(event)
            return
        end
        dedot(event.log.kubernetes.labels)
        flatten_labels(event)
        prune_labels(event)
        emit(event)
    end

    function flatten_labels(event)
        -- create "flat_labels" key
        event.log.kubernetes.flat_labels = {}
        i = 1
        -- flatten the labels
        for k,v in pairs(event.log.kubernetes.labels) do
          event.log.kubernetes.flat_labels[i] = k.."="..v
          i=i+1
        end
    end 

	function prune_labels(event)
		local exclusions = {"app_kubernetes_io/name", "app_kubernetes_io/instance", "app_kubernetes_io/version", "app_kubernetes_io/component", "app_kubernetes_io/part-of", "app_kubernetes_io/managed-by", "app_kubernetes_io/created-by"}
		local keys = {}
		for k,v in pairs(event.log.kubernetes.labels) do
			for index, e in pairs(exclusions) do
				if k == e then
					keys[k] = v
				end
			end
		end
		event.log.kubernetes.labels = keys
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
bulk.index = "{{ write_index }}"
bulk.action = "create"
request.timeout_secs = 2147483648
id_key = "_id"
`,
		}),
	)
})

func TestVectorConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vector Conf Generation")
}
