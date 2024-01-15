package elasticsearch

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	vectorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/output/common"
	"testing"

	"github.com/openshift/cluster-logging-operator/test/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Generate Vector config", func() {
	inputPipeline := []string{"application"}
	var f = func(clspec logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op framework.Options) []framework.Element {
		e := []framework.Element{}
		for _, o := range clfspec.Outputs {
			e = framework.MergeElements(e, New(vectorhelpers.FormatComponentID(o.Name), o, inputPipeline, secrets[o.Name], op))
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
source = '''
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
'''

[transforms.es_1_dedot_and_flatten]
type = "remap"
inputs = ["es_1_add_es_index"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
    .kubernetes.flat_labels = []
    for_each(object!(.kubernetes.labels)) -> |k,v| {
    .kubernetes.flat_labels = push(.kubernetes.flat_labels, join!([string(k),"=",string!(v)]))
    }
  }
  if exists(.kubernetes.labels) {
    exclusions = ["app_kubernetes_io_name", "app_kubernetes_io_instance", "app_kubernetes_io_version", "app_kubernetes_io_component", "app_kubernetes_io_part-of", "app_kubernetes_io_managed-by", "app_kubernetes_io_created-by"]
    keep = {}
    for_each(object!(.kubernetes.labels))->|k,v|{
      if !includes(exclusions, k) {
        .kubernetes.labels, err = remove(object!(.kubernetes.labels),[k],true)
        if err != null {
          log(err, level: "error")
        }
      }
    }
  }
'''
[sinks.es_1]
type = "elasticsearch"
inputs = ["es_1_dedot_and_flatten"]
endpoints = ["https://es.svc.infra.cluster:9200"]
bulk.index = "{{ write_index }}"
bulk.action = "create"
encoding.except_fields = ["write_index"]
id_key = "_id"
api_version = "v6"

[sinks.es_1.buffer]
when_full = "drop_newest"

[sinks.es_1.request]
retry_attempts = 17
timeout_secs = 2147483648

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
source = '''
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
'''

[transforms.es_1_dedot_and_flatten]
type = "remap"
inputs = ["es_1_add_es_index"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
    .kubernetes.flat_labels = []
    for_each(object!(.kubernetes.labels)) -> |k,v| {
    .kubernetes.flat_labels = push(.kubernetes.flat_labels, join!([string(k),"=",string!(v)]))
    }
  }
  if exists(.kubernetes.labels) {
    exclusions = ["app_kubernetes_io_name", "app_kubernetes_io_instance", "app_kubernetes_io_version", "app_kubernetes_io_component", "app_kubernetes_io_part-of", "app_kubernetes_io_managed-by", "app_kubernetes_io_created-by"]
    keep = {}
    for_each(object!(.kubernetes.labels))->|k,v|{
      if !includes(exclusions, k) {
        .kubernetes.labels, err = remove(object!(.kubernetes.labels),[k],true)
        if err != null {
          log(err, level: "error")
        }
      }
    }
  }
'''
[sinks.es_1]
type = "elasticsearch"
inputs = ["es_1_dedot_and_flatten"]
endpoints = ["https://es.svc.infra.cluster:9200"]
bulk.index = "{{ write_index }}"
bulk.action = "create"
encoding.except_fields = ["write_index"]
id_key = "_id"
api_version = "v6"

[sinks.es_1.buffer]
when_full = "drop_newest"

[sinks.es_1.request]
retry_attempts = 17
timeout_secs = 2147483648

[sinks.es_1.tls]
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
			Secrets: common.NoSecrets,
			ExpectedConf: `
# Set Elasticsearch index
[transforms.es_1_add_es_index]
type = "remap"
inputs = ["application"]
source = '''
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
'''

[transforms.es_1_dedot_and_flatten]
type = "remap"
inputs = ["es_1_add_es_index"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
    .kubernetes.flat_labels = []
    for_each(object!(.kubernetes.labels)) -> |k,v| {
    .kubernetes.flat_labels = push(.kubernetes.flat_labels, join!([string(k),"=",string!(v)]))
    }
  }
  if exists(.kubernetes.labels) {
    exclusions = ["app_kubernetes_io_name", "app_kubernetes_io_instance", "app_kubernetes_io_version", "app_kubernetes_io_component", "app_kubernetes_io_part-of", "app_kubernetes_io_managed-by", "app_kubernetes_io_created-by"]
    keep = {}
    for_each(object!(.kubernetes.labels))->|k,v|{
      if !includes(exclusions, k) {
        .kubernetes.labels, err = remove(object!(.kubernetes.labels),[k],true)
        if err != null {
          log(err, level: "error")
        }
      }
    }
  }
'''
[sinks.es_1]
type = "elasticsearch"
inputs = ["es_1_dedot_and_flatten"]
endpoints = ["http://es.svc.infra.cluster:9200"]
bulk.index = "{{ write_index }}"
bulk.action = "create"
encoding.except_fields = ["write_index"]
id_key = "_id"
api_version = "v6"

[sinks.es_1.buffer]
when_full = "drop_newest"

[sinks.es_1.request]
retry_attempts = 17
timeout_secs = 2147483648
`,
		}),
		Entry("without secret and TLS.insecureSkipVerify=true", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type:   logging.OutputTypeElasticsearch,
						Name:   "es-1",
						URL:    "https://es.svc.infra.cluster:9200",
						Secret: nil,
						TLS: &logging.OutputTLSSpec{
							InsecureSkipVerify: true,
						},
					},
				},
			},
			Secrets: common.NoSecrets,
			ExpectedConf: `
# Set Elasticsearch index
[transforms.es_1_add_es_index]
type = "remap"
inputs = ["application"]
source = '''
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
'''

[transforms.es_1_dedot_and_flatten]
type = "remap"
inputs = ["es_1_add_es_index"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
    .kubernetes.flat_labels = []
    for_each(object!(.kubernetes.labels)) -> |k,v| {
    .kubernetes.flat_labels = push(.kubernetes.flat_labels, join!([string(k),"=",string!(v)]))
    }
  }
  if exists(.kubernetes.labels) {
    exclusions = ["app_kubernetes_io_name", "app_kubernetes_io_instance", "app_kubernetes_io_version", "app_kubernetes_io_component", "app_kubernetes_io_part-of", "app_kubernetes_io_managed-by", "app_kubernetes_io_created-by"]
    keep = {}
    for_each(object!(.kubernetes.labels))->|k,v|{
      if !includes(exclusions, k) {
        .kubernetes.labels, err = remove(object!(.kubernetes.labels),[k],true)
        if err != null {
          log(err, level: "error")
        }
      }
    }
  }
'''

[sinks.es_1]
type = "elasticsearch"
inputs = ["es_1_dedot_and_flatten"]
endpoints = ["https://es.svc.infra.cluster:9200"]
bulk.index = "{{ write_index }}"
bulk.action = "create"
encoding.except_fields = ["write_index"]
id_key = "_id"
api_version = "v6"

[sinks.es_1.buffer]
when_full = "drop_newest"

[sinks.es_1.request]
retry_attempts = 17
timeout_secs = 2147483648

[sinks.es_1.tls]
verify_certificate = false
verify_hostname = false
`,
		}),
		Entry("with multiple pipelines for elastic-search", helpers.ConfGenerateTest{
			CLSpec: logging.CollectionSpec{},
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
source = '''
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
'''

[transforms.es_1_dedot_and_flatten]
type = "remap"
inputs = ["es_1_add_es_index"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
    .kubernetes.flat_labels = []
    for_each(object!(.kubernetes.labels)) -> |k,v| {
    .kubernetes.flat_labels = push(.kubernetes.flat_labels, join!([string(k),"=",string!(v)]))
    }
  }
  if exists(.kubernetes.labels) {
    exclusions = ["app_kubernetes_io_name", "app_kubernetes_io_instance", "app_kubernetes_io_version", "app_kubernetes_io_component", "app_kubernetes_io_part-of", "app_kubernetes_io_managed-by", "app_kubernetes_io_created-by"]
    keep = {}
    for_each(object!(.kubernetes.labels))->|k,v|{
      if !includes(exclusions, k) {
        .kubernetes.labels, err = remove(object!(.kubernetes.labels),[k],true)
        if err != null {
          log(err, level: "error")
        }
      }
    }
  }
'''
[sinks.es_1]
type = "elasticsearch"
inputs = ["es_1_dedot_and_flatten"]
endpoints = ["https://es-1.svc.messaging.cluster.local:9200"]
bulk.index = "{{ write_index }}"
bulk.action = "create"
encoding.except_fields = ["write_index"]
id_key = "_id"
api_version = "v6"
[sinks.es_1.buffer]
when_full = "drop_newest"

[sinks.es_1.request]
retry_attempts = 17
timeout_secs = 2147483648
[sinks.es_1.tls]
key_file = "/var/run/ocp-collector/secrets/es-1/tls.key"
crt_file = "/var/run/ocp-collector/secrets/es-1/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/es-1/ca-bundle.crt"

# Set Elasticsearch index
[transforms.es_2_add_es_index]
type = "remap"
inputs = ["application"]
source = '''
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
'''

[transforms.es_2_dedot_and_flatten]
type = "remap"
inputs = ["es_2_add_es_index"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
    .kubernetes.flat_labels = []
    for_each(object!(.kubernetes.labels)) -> |k,v| {
    .kubernetes.flat_labels = push(.kubernetes.flat_labels, join!([string(k),"=",string!(v)]))
    }
  }
  if exists(.kubernetes.labels) {
    exclusions = ["app_kubernetes_io_name", "app_kubernetes_io_instance", "app_kubernetes_io_version", "app_kubernetes_io_component", "app_kubernetes_io_part-of", "app_kubernetes_io_managed-by", "app_kubernetes_io_created-by"]
    keep = {}
    for_each(object!(.kubernetes.labels))->|k,v|{
      if !includes(exclusions, k) {
        .kubernetes.labels, err = remove(object!(.kubernetes.labels),[k],true)
        if err != null {
          log(err, level: "error")
        }
      }
    }
  }
'''
[sinks.es_2]
type = "elasticsearch"
inputs = ["es_2_dedot_and_flatten"]
endpoints = ["https://es-2.svc.messaging.cluster.local:9200"]
bulk.index = "{{ write_index }}"
bulk.action = "create"
encoding.except_fields = ["write_index"]
id_key = "_id"
api_version = "v6"
[sinks.es_2.buffer]
when_full = "drop_newest"

[sinks.es_2.request]
retry_attempts = 17
timeout_secs = 2147483648

[sinks.es_2.tls]
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
								ElasticsearchStructuredSpec: logging.ElasticsearchStructuredSpec{
									StructuredTypeKey: "kubernetes.labels.mylabel",
								},
							},
						},
					},
				},
			},
			Secrets: common.NoSecrets,
			ExpectedConf: `
# Set Elasticsearch index
[transforms.es_1_add_es_index]
type = "remap"
inputs = ["application"]
source = '''
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

  if .structured != null && .write_index == "app-write" {
    .message = encode_json(.structured)
    del(.structured)
  }
'''

[transforms.es_1_dedot_and_flatten]
type = "remap"
inputs = ["es_1_add_es_index"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
    .kubernetes.flat_labels = []
    for_each(object!(.kubernetes.labels)) -> |k,v| {
    .kubernetes.flat_labels = push(.kubernetes.flat_labels, join!([string(k),"=",string!(v)]))
    }
  }
  if exists(.kubernetes.labels) {
    exclusions = ["app_kubernetes_io_name", "app_kubernetes_io_instance", "app_kubernetes_io_version", "app_kubernetes_io_component", "app_kubernetes_io_part-of", "app_kubernetes_io_managed-by", "app_kubernetes_io_created-by"]
    keep = {}
    for_each(object!(.kubernetes.labels))->|k,v|{
      if !includes(exclusions, k) {
        .kubernetes.labels, err = remove(object!(.kubernetes.labels),[k],true)
        if err != null {
          log(err, level: "error")
        }
      }
    }
  }
'''
[sinks.es_1]
type = "elasticsearch"
inputs = ["es_1_dedot_and_flatten"]
endpoints = ["http://es.svc.infra.cluster:9200"]
bulk.index = "{{ write_index }}"
bulk.action = "create"
encoding.except_fields = ["write_index"]
id_key = "_id"
api_version = "v6"
[sinks.es_1.buffer]
when_full = "drop_newest"

[sinks.es_1.request]
retry_attempts = 17
timeout_secs = 2147483648
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
								ElasticsearchStructuredSpec: logging.ElasticsearchStructuredSpec{
									StructuredTypeName: "myindex",
								},
							},
						},
					},
				},
			},
			Secrets: common.NoSecrets,
			ExpectedConf: `
# Set Elasticsearch index
[transforms.es_1_add_es_index]
type = "remap"
inputs = ["application"]
source = '''
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
  if .structured != null && .write_index == "app-write" {
    .message = encode_json(.structured)
    del(.structured)
  }
'''

[transforms.es_1_dedot_and_flatten]
type = "remap"
inputs = ["es_1_add_es_index"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
    .kubernetes.flat_labels = []
    for_each(object!(.kubernetes.labels)) -> |k,v| {
    .kubernetes.flat_labels = push(.kubernetes.flat_labels, join!([string(k),"=",string!(v)]))
    }
  }
  if exists(.kubernetes.labels) {
    exclusions = ["app_kubernetes_io_name", "app_kubernetes_io_instance", "app_kubernetes_io_version", "app_kubernetes_io_component", "app_kubernetes_io_part-of", "app_kubernetes_io_managed-by", "app_kubernetes_io_created-by"]
    keep = {}
    for_each(object!(.kubernetes.labels))->|k,v|{
      if !includes(exclusions, k) {
        .kubernetes.labels, err = remove(object!(.kubernetes.labels),[k],true)
        if err != null {
          log(err, level: "error")
        }
      }
    }
  }
'''
[sinks.es_1]
type = "elasticsearch"
inputs = ["es_1_dedot_and_flatten"]
endpoints = ["http://es.svc.infra.cluster:9200"]
bulk.index = "{{ write_index }}"
bulk.action = "create"
encoding.except_fields = ["write_index"]
id_key = "_id"
api_version = "v6"
[sinks.es_1.buffer]
when_full = "drop_newest"

[sinks.es_1.request]
retry_attempts = 17
timeout_secs = 2147483648
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
								ElasticsearchStructuredSpec: logging.ElasticsearchStructuredSpec{
									StructuredTypeKey:  "kubernetes.labels.mylabel",
									StructuredTypeName: "myindex",
								},
							},
						},
					},
				},
			},
			Secrets: common.NoSecrets,
			ExpectedConf: `
# Set Elasticsearch index
[transforms.es_1_add_es_index]
type = "remap"
inputs = ["application"]
source = '''
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
  if .structured != null && .write_index == "app-write" {
    .message = encode_json(.structured)
    del(.structured)
  }
'''

[transforms.es_1_dedot_and_flatten]
type = "remap"
inputs = ["es_1_add_es_index"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
    .kubernetes.flat_labels = []
    for_each(object!(.kubernetes.labels)) -> |k,v| {
    .kubernetes.flat_labels = push(.kubernetes.flat_labels, join!([string(k),"=",string!(v)]))
    }
  }
  if exists(.kubernetes.labels) {
    exclusions = ["app_kubernetes_io_name", "app_kubernetes_io_instance", "app_kubernetes_io_version", "app_kubernetes_io_component", "app_kubernetes_io_part-of", "app_kubernetes_io_managed-by", "app_kubernetes_io_created-by"]
    keep = {}
    for_each(object!(.kubernetes.labels))->|k,v|{
      if !includes(exclusions, k) {
        .kubernetes.labels, err = remove(object!(.kubernetes.labels),[k],true)
        if err != null {
          log(err, level: "error")
        }
      }
    }
  }
'''
[sinks.es_1]
type = "elasticsearch"
inputs = ["es_1_dedot_and_flatten"]
endpoints = ["http://es.svc.infra.cluster:9200"]
bulk.index = "{{ write_index }}"
bulk.action = "create"
encoding.except_fields = ["write_index"]
id_key = "_id"
api_version = "v6"
[sinks.es_1.buffer]
when_full = "drop_newest"

[sinks.es_1.request]
retry_attempts = 17
timeout_secs = 2147483648
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
								ElasticsearchStructuredSpec: logging.ElasticsearchStructuredSpec{
									StructuredTypeKey:             "kubernetes.labels.mylabel",
									StructuredTypeName:            "myindex",
									EnableStructuredContainerLogs: true,
								},
							},
						},
					},
				},
			},
			Secrets: common.NoSecrets,
			ExpectedConf: `
# Set Elasticsearch index
[transforms.es_1_add_es_index]
type = "remap"
inputs = ["application"]
source = '''
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

  if .structured != null && .write_index == "app-write" {
    .message = encode_json(.structured)
    del(.structured)
  }
'''

[transforms.es_1_dedot_and_flatten]
type = "remap"
inputs = ["es_1_add_es_index"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
    .kubernetes.flat_labels = []
    for_each(object!(.kubernetes.labels)) -> |k,v| {
    .kubernetes.flat_labels = push(.kubernetes.flat_labels, join!([string(k),"=",string!(v)]))
    }
  }
  if exists(.kubernetes.labels) {
    exclusions = ["app_kubernetes_io_name", "app_kubernetes_io_instance", "app_kubernetes_io_version", "app_kubernetes_io_component", "app_kubernetes_io_part-of", "app_kubernetes_io_managed-by", "app_kubernetes_io_created-by"]
    keep = {}
    for_each(object!(.kubernetes.labels))->|k,v|{
      if !includes(exclusions, k) {
        .kubernetes.labels, err = remove(object!(.kubernetes.labels),[k],true)
        if err != null {
          log(err, level: "error")
        }
      }
    }
  }
'''
[sinks.es_1]
type = "elasticsearch"
inputs = ["es_1_dedot_and_flatten"]
endpoints = ["http://es.svc.infra.cluster:9200"]
bulk.index = "{{ write_index }}"
bulk.action = "create"
encoding.except_fields = ["write_index"]
id_key = "_id"
api_version = "v6"
[sinks.es_1.buffer]
when_full = "drop_newest"

[sinks.es_1.request]
retry_attempts = 17
timeout_secs = 2147483648
`,
		}),
		Entry("without an Elasticsearch version", helpers.ConfGenerateTest{
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
			Secrets: common.NoSecrets,
			ExpectedConf: `
# Set Elasticsearch index
[transforms.es_1_add_es_index]
type = "remap"
inputs = ["application"]
source = '''
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
'''

[transforms.es_1_dedot_and_flatten]
type = "remap"
inputs = ["es_1_add_es_index"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
    .kubernetes.flat_labels = []
    for_each(object!(.kubernetes.labels)) -> |k,v| {
    .kubernetes.flat_labels = push(.kubernetes.flat_labels, join!([string(k),"=",string!(v)]))
    }
  }
  if exists(.kubernetes.labels) {
    exclusions = ["app_kubernetes_io_name", "app_kubernetes_io_instance", "app_kubernetes_io_version", "app_kubernetes_io_component", "app_kubernetes_io_part-of", "app_kubernetes_io_managed-by", "app_kubernetes_io_created-by"]
    keep = {}
    for_each(object!(.kubernetes.labels))->|k,v|{
      if !includes(exclusions, k) {
        .kubernetes.labels, err = remove(object!(.kubernetes.labels),[k],true)
        if err != null {
          log(err, level: "error")
        }
      }
    }
  }
'''

[sinks.es_1]
type = "elasticsearch"
inputs = ["es_1_dedot_and_flatten"]
endpoints = ["http://es.svc.infra.cluster:9200"]
bulk.index = "{{ write_index }}"
bulk.action = "create"
encoding.except_fields = ["write_index"]
id_key = "_id"
api_version = "v6"
[sinks.es_1.buffer]
when_full = "drop_newest"

[sinks.es_1.request]
retry_attempts = 17
timeout_secs = 2147483648
`,
		}),
		Entry("with an Elasticsearch version less than our default", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeElasticsearch,
						Name: "es-1",
						URL:  "http://es.svc.infra.cluster:9200",
						OutputTypeSpec: logging.OutputTypeSpec{
							Elasticsearch: &logging.Elasticsearch{
								Version: 5,
							},
						},
						Secret: nil,
					},
				},
			},
			Secrets: common.NoSecrets,
			ExpectedConf: `
# Set Elasticsearch index
[transforms.es_1_add_es_index]
type = "remap"
inputs = ["application"]
source = '''
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
  if .structured != null && .write_index == "app-write" {
    .message = encode_json(.structured)
    del(.structured)
  }
'''

[transforms.es_1_dedot_and_flatten]
type = "remap"
inputs = ["es_1_add_es_index"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
    .kubernetes.flat_labels = []
    for_each(object!(.kubernetes.labels)) -> |k,v| {
    .kubernetes.flat_labels = push(.kubernetes.flat_labels, join!([string(k),"=",string!(v)]))
    }
  }
  if exists(.kubernetes.labels) {
    exclusions = ["app_kubernetes_io_name", "app_kubernetes_io_instance", "app_kubernetes_io_version", "app_kubernetes_io_component", "app_kubernetes_io_part-of", "app_kubernetes_io_managed-by", "app_kubernetes_io_created-by"]
    keep = {}
    for_each(object!(.kubernetes.labels))->|k,v|{
      if !includes(exclusions, k) {
        .kubernetes.labels, err = remove(object!(.kubernetes.labels),[k],true)
        if err != null {
          log(err, level: "error")
        }
      }
    }
  }
'''
[sinks.es_1]
type = "elasticsearch"
inputs = ["es_1_dedot_and_flatten"]
endpoints = ["http://es.svc.infra.cluster:9200"]
bulk.index = "{{ write_index }}"
bulk.action = "create"
encoding.except_fields = ["write_index"]
id_key = "_id"
api_version = "v5"
[sinks.es_1.buffer]
when_full = "drop_newest"

[sinks.es_1.request]
retry_attempts = 17
timeout_secs = 2147483648
`,
		}),
		Entry("with our default Elasticsearch version", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeElasticsearch,
						Name: "es-1",
						URL:  "http://es.svc.infra.cluster:9200",
						OutputTypeSpec: logging.OutputTypeSpec{
							Elasticsearch: &logging.Elasticsearch{
								Version: logging.DefaultESVersion,
							},
						},
						Secret: nil,
					},
				},
			},
			Secrets: common.NoSecrets,
			ExpectedConf: `
# Set Elasticsearch index
[transforms.es_1_add_es_index]
type = "remap"
inputs = ["application"]
source = '''
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
  if .structured != null && .write_index == "app-write" {
    .message = encode_json(.structured)
    del(.structured)
  }
'''

[transforms.es_1_dedot_and_flatten]
type = "remap"
inputs = ["es_1_add_es_index"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
    .kubernetes.flat_labels = []
    for_each(object!(.kubernetes.labels)) -> |k,v| {
    .kubernetes.flat_labels = push(.kubernetes.flat_labels, join!([string(k),"=",string!(v)]))
    }
  }
  if exists(.kubernetes.labels) {
    exclusions = ["app_kubernetes_io_name", "app_kubernetes_io_instance", "app_kubernetes_io_version", "app_kubernetes_io_component", "app_kubernetes_io_part-of", "app_kubernetes_io_managed-by", "app_kubernetes_io_created-by"]
    keep = {}
    for_each(object!(.kubernetes.labels))->|k,v|{
      if !includes(exclusions, k) {
        .kubernetes.labels, err = remove(object!(.kubernetes.labels),[k],true)
        if err != null {
          log(err, level: "error")
        }
      }
    }
  }
'''
[sinks.es_1]
type = "elasticsearch"
inputs = ["es_1_dedot_and_flatten"]
endpoints = ["http://es.svc.infra.cluster:9200"]
bulk.index = "{{ write_index }}"
bulk.action = "create"
encoding.except_fields = ["write_index"]
id_key = "_id"
api_version = "v6"
[sinks.es_1.buffer]
when_full = "drop_newest"

[sinks.es_1.request]
retry_attempts = 17
timeout_secs = 2147483648
`,
		}),
		Entry("with Elasticsearch version 7", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeElasticsearch,
						Name: "es-1",
						URL:  "http://es.svc.infra.cluster:9200",
						OutputTypeSpec: logging.OutputTypeSpec{
							Elasticsearch: &logging.Elasticsearch{
								Version: 7,
							},
						},
						Secret: nil,
					},
				},
			},
			Secrets: common.NoSecrets,
			ExpectedConf: `
# Set Elasticsearch index
[transforms.es_1_add_es_index]
type = "remap"
inputs = ["application"]
source = '''
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
  if .structured != null && .write_index == "app-write" {
    .message = encode_json(.structured)
    del(.structured)
  }
'''

[transforms.es_1_dedot_and_flatten]
type = "remap"
inputs = ["es_1_add_es_index"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
    .kubernetes.flat_labels = []
    for_each(object!(.kubernetes.labels)) -> |k,v| {
    .kubernetes.flat_labels = push(.kubernetes.flat_labels, join!([string(k),"=",string!(v)]))
    }
  }
  if exists(.kubernetes.labels) {
    exclusions = ["app_kubernetes_io_name", "app_kubernetes_io_instance", "app_kubernetes_io_version", "app_kubernetes_io_component", "app_kubernetes_io_part-of", "app_kubernetes_io_managed-by", "app_kubernetes_io_created-by"]
    keep = {}
    for_each(object!(.kubernetes.labels))->|k,v|{
      if !includes(exclusions, k) {
        .kubernetes.labels, err = remove(object!(.kubernetes.labels),[k],true)
        if err != null {
          log(err, level: "error")
        }
      }
    }
  }
'''
[sinks.es_1]
type = "elasticsearch"
inputs = ["es_1_dedot_and_flatten"]
endpoints = ["http://es.svc.infra.cluster:9200"]
bulk.index = "{{ write_index }}"
bulk.action = "create"
encoding.except_fields = ["write_index"]
id_key = "_id"
api_version = "v7"
[sinks.es_1.buffer]
when_full = "drop_newest"

[sinks.es_1.request]
retry_attempts = 17
timeout_secs = 2147483648
`,
		}),
		Entry("with an Elasticsearch version greater than latest version", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeElasticsearch,
						Name: "es-1",
						URL:  "http://es.svc.infra.cluster:9200",
						OutputTypeSpec: logging.OutputTypeSpec{
							Elasticsearch: &logging.Elasticsearch{
								Version: logging.FirstESVersionWithoutType + 1,
							},
						},
						Secret: nil,
					},
				},
			},
			Secrets: common.NoSecrets,
			ExpectedConf: `
# Set Elasticsearch index
[transforms.es_1_add_es_index]
type = "remap"
inputs = ["application"]
source = '''
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
  if .structured != null && .write_index == "app-write" {
    .message = encode_json(.structured)
    del(.structured)
  }
'''

[transforms.es_1_dedot_and_flatten]
type = "remap"
inputs = ["es_1_add_es_index"]
source = '''
  .openshift.sequence = to_unix_timestamp(now(), unit: "nanoseconds")
  if exists(.kubernetes.namespace_labels) {
	  for_each(object!(.kubernetes.namespace_labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.namespace_labels = set!(.kubernetes.namespace_labels,[newkey],value)
		if newkey != key {
		  .kubernetes.namespace_labels = remove!(.kubernetes.namespace_labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
	  for_each(object!(.kubernetes.labels)) -> |key,value| { 
		newkey = replace(key, r'[\./]', "_") 
		.kubernetes.labels = set!(.kubernetes.labels,[newkey],value)
		if newkey != key {
		  .kubernetes.labels = remove!(.kubernetes.labels,[key],true)
		}
	  }
  }
  if exists(.kubernetes.labels) {
    .kubernetes.flat_labels = []
    for_each(object!(.kubernetes.labels)) -> |k,v| {
    .kubernetes.flat_labels = push(.kubernetes.flat_labels, join!([string(k),"=",string!(v)]))
    }
  }
  if exists(.kubernetes.labels) {
    exclusions = ["app_kubernetes_io_name", "app_kubernetes_io_instance", "app_kubernetes_io_version", "app_kubernetes_io_component", "app_kubernetes_io_part-of", "app_kubernetes_io_managed-by", "app_kubernetes_io_created-by"]
    keep = {}
    for_each(object!(.kubernetes.labels))->|k,v|{
      if !includes(exclusions, k) {
        .kubernetes.labels, err = remove(object!(.kubernetes.labels),[k],true)
        if err != null {
          log(err, level: "error")
        }
      }
    }
  }
'''

[sinks.es_1]
type = "elasticsearch"
inputs = ["es_1_dedot_and_flatten"]
endpoints = ["http://es.svc.infra.cluster:9200"]
bulk.index = "{{ write_index }}"
bulk.action = "create"
encoding.except_fields = ["write_index"]
id_key = "_id"
api_version = "v9"
[sinks.es_1.buffer]
when_full = "drop_newest"

[sinks.es_1.request]
retry_attempts = 17
timeout_secs = 2147483648
`,
		}),
	)
})

func TestVectorConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vector Conf Generation")
}
