package elasticsearch

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/elements"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	testhelpers "github.com/openshift/cluster-logging-operator/test/helpers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Generate fluentd config", func() {
	var f = func(clspec logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		var bufspec *logging.FluentdBufferSpec = nil
		if clspec.Fluentd != nil &&
			clspec.Fluentd.Buffer != nil {
			bufspec = clspec.Fluentd.Buffer
		}
		return Conf(bufspec, secrets[clfspec.Outputs[0].Name], clfspec.Outputs[0], op)
	}
	DescribeTable("for Elasticsearch output", testhelpers.TestGenerateConfAndFormatWith(f, helpers.FormatFluentConf),
		Entry("with username,password", testhelpers.ConfGenerateTest{
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
<label @ES_1>
  # Viaq Data Model
  <filter **>
    @type viaq_data_model
    enable_openshift_model false
    enable_prune_empty_fields false
    rename_time false
    undefined_dot_replace_char UNUSED
    elasticsearch_index_prefix_field 'viaq_index_name'
    <elasticsearch_index_name>
      enabled 'true'
      tag "kubernetes.var.log.pods.openshift_** kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_** var.log.pods.openshift_** var.log.pods.openshift-*_** var.log.pods.default_** var.log.pods.kube-*_** journal.system** system.var.log**"
      name_type static
      static_index_name infra-write
    </elasticsearch_index_name>
    <elasticsearch_index_name>
      enabled 'true'
      tag "linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**"
      name_type static
      static_index_name audit-write
    </elasticsearch_index_name>
    <elasticsearch_index_name>
      enabled 'true'
      tag "**"
      name_type structured
      static_index_name app-write
    </elasticsearch_index_name>
  </filter>
  <filter **>
    @type viaq_data_model
    enable_prune_labels true
    enable_openshift_model false
    rename_time false
    undefined_dot_replace_char UNUSED
    prune_labels_exclusions app.kubernetes.io/name,app.kubernetes.io/instance,app.kubernetes.io/version,app.kubernetes.io/component,app.kubernetes.io/part-of,app.kubernetes.io/managed-by,app.kubernetes.io/created-by
  </filter>
  
  #remove structured field if present
  <filter **>
    @type record_modifier
    remove_keys structured
  </filter>
  
  <match retry_es_1>
    @type elasticsearch
    @id retry_es_1
    host es.svc.infra.cluster
    port 9999
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    user "#{File.exists?('/var/run/ocp-collector/secrets/es-1/username') ? open('/var/run/ocp-collector/secrets/es-1/username','r') do |f|f.read end : ''}"
    password "#{File.exists?('/var/run/ocp-collector/secrets/es-1/password') ? open('/var/run/ocp-collector/secrets/es-1/password','r') do |f|f.read end : ''}"
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    http_backend typhoeus
    write_operation create
    # https://github.com/uken/fluent-plugin-elasticsearch#suppress_type_name
    suppress_type_name 'true'
    reload_connections 'true'
    # https://github.com/uken/fluent-plugin-elasticsearch#reload-after
    reload_after '200'
    # https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
    sniffer_class_name 'Fluent::Plugin::ElasticsearchSimpleSniffer'
    reload_on_failure false
    # 2 ^ 31
    request_timeout 2147483648
    <buffer>
      @type file
      path '/var/lib/fluentd/retry_es_1'
      flush_mode interval
      flush_interval 1s
      flush_thread_count 2
      retry_type exponential_backoff
      retry_wait 1s
      retry_max_interval 60s
      retry_timeout 60m
      queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
      total_limit_size "#{ENV['TOTAL_LIMIT_SIZE_PER_BUFFER'] || '8589934592'}"
      chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
      overflow_action block
      disable_chunk_backup true
    </buffer>
  </match>
  
  <match **>
    @type elasticsearch
    @id es_1
    host es.svc.infra.cluster
    port 9999
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    user "#{File.exists?('/var/run/ocp-collector/secrets/es-1/username') ? open('/var/run/ocp-collector/secrets/es-1/username','r') do |f|f.read end : ''}"
    password "#{File.exists?('/var/run/ocp-collector/secrets/es-1/password') ? open('/var/run/ocp-collector/secrets/es-1/password','r') do |f|f.read end : ''}"
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    retry_tag retry_es_1
    http_backend typhoeus
    write_operation create
    # https://github.com/uken/fluent-plugin-elasticsearch#suppress_type_name
    suppress_type_name 'true'
    reload_connections 'true'
    # https://github.com/uken/fluent-plugin-elasticsearch#reload-after
    reload_after '200'
    # https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
    sniffer_class_name 'Fluent::Plugin::ElasticsearchSimpleSniffer'
    reload_on_failure false
    # 2 ^ 31
    request_timeout 2147483648
    <buffer>
      @type file
      path '/var/lib/fluentd/es_1'
      flush_mode interval
      flush_interval 1s
      flush_thread_count 2
      retry_type exponential_backoff
      retry_wait 1s
      retry_max_interval 60s
      retry_timeout 60m
      queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
      total_limit_size "#{ENV['TOTAL_LIMIT_SIZE_PER_BUFFER'] || '8589934592'}"
      chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
      overflow_action block
      disable_chunk_backup true
    </buffer>
  </match>
</label>
`,
		}),
		Entry("with tls key,cert,ca-bundle", testhelpers.ConfGenerateTest{
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
						"tls.key":       []byte("junk"),
						"tls.crt":       []byte("junk"),
						"ca-bundle.crt": []byte("junk"),
					},
				},
			},
			ExpectedConf: `
<label @ES_1>
  # Viaq Data Model
  <filter **>
    @type viaq_data_model
    enable_openshift_model false
    enable_prune_empty_fields false
    rename_time false
    undefined_dot_replace_char UNUSED
    elasticsearch_index_prefix_field 'viaq_index_name'
    <elasticsearch_index_name>
      enabled 'true'
      tag "kubernetes.var.log.pods.openshift_** kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_** var.log.pods.openshift_** var.log.pods.openshift-*_** var.log.pods.default_** var.log.pods.kube-*_** journal.system** system.var.log**"
      name_type static
      static_index_name infra-write
    </elasticsearch_index_name>
    <elasticsearch_index_name>
      enabled 'true'
      tag "linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**"
      name_type static
      static_index_name audit-write
    </elasticsearch_index_name>
    <elasticsearch_index_name>
      enabled 'true'
      tag "**"
      name_type structured
      static_index_name app-write
    </elasticsearch_index_name>
  </filter>
  <filter **>
    @type viaq_data_model
    enable_prune_labels true
    enable_openshift_model false
    rename_time false
    undefined_dot_replace_char UNUSED
    prune_labels_exclusions app.kubernetes.io/name,app.kubernetes.io/instance,app.kubernetes.io/version,app.kubernetes.io/component,app.kubernetes.io/part-of,app.kubernetes.io/managed-by,app.kubernetes.io/created-by
  </filter>
  
  #remove structured field if present
  <filter **>
    @type record_modifier
    remove_keys structured
  </filter>
  
  <match retry_es_1>
    @type elasticsearch
    @id retry_es_1
    host es.svc.infra.cluster
    port 9999
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    client_key '/var/run/ocp-collector/secrets/es-1/tls.key'
    client_cert '/var/run/ocp-collector/secrets/es-1/tls.crt'
    ca_file '/var/run/ocp-collector/secrets/es-1/ca-bundle.crt'
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    http_backend typhoeus
    write_operation create
    # https://github.com/uken/fluent-plugin-elasticsearch#suppress_type_name
    suppress_type_name 'true'
    reload_connections 'true'
    # https://github.com/uken/fluent-plugin-elasticsearch#reload-after
    reload_after '200'
    # https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
    sniffer_class_name 'Fluent::Plugin::ElasticsearchSimpleSniffer'
    reload_on_failure false
    # 2 ^ 31
    request_timeout 2147483648
    <buffer>
      @type file
      path '/var/lib/fluentd/retry_es_1'
      flush_mode interval
      flush_interval 1s
      flush_thread_count 2
      retry_type exponential_backoff
      retry_wait 1s
      retry_max_interval 60s
      retry_timeout 60m
      queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
      total_limit_size "#{ENV['TOTAL_LIMIT_SIZE_PER_BUFFER'] || '8589934592'}"
      chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
      overflow_action block
      disable_chunk_backup true
    </buffer>
  </match>
  
  <match **>
    @type elasticsearch
    @id es_1
    host es.svc.infra.cluster
    port 9999
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    client_key '/var/run/ocp-collector/secrets/es-1/tls.key'
    client_cert '/var/run/ocp-collector/secrets/es-1/tls.crt'
    ca_file '/var/run/ocp-collector/secrets/es-1/ca-bundle.crt'
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    retry_tag retry_es_1
    http_backend typhoeus
    write_operation create
    # https://github.com/uken/fluent-plugin-elasticsearch#suppress_type_name
    suppress_type_name 'true'
    reload_connections 'true'
    # https://github.com/uken/fluent-plugin-elasticsearch#reload-after
    reload_after '200'
    # https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
    sniffer_class_name 'Fluent::Plugin::ElasticsearchSimpleSniffer'
    reload_on_failure false
    # 2 ^ 31
    request_timeout 2147483648
    <buffer>
      @type file
      path '/var/lib/fluentd/es_1'
      flush_mode interval
      flush_interval 1s
      flush_thread_count 2
      retry_type exponential_backoff
      retry_wait 1s
      retry_max_interval 60s
      retry_timeout 60m
      queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
      total_limit_size "#{ENV['TOTAL_LIMIT_SIZE_PER_BUFFER'] || '8589934592'}"
      chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
      overflow_action block
      disable_chunk_backup true
    </buffer>
  </match>
</label>
`,
		}),
		Entry("without security", testhelpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeElasticsearch,
						Name: "es-1",
						URL:  "http://es.svc.infra.cluster:9999",
						Secret: &logging.OutputSecretSpec{
							Name: "es-1",
						},
					},
				},
			},
			Secrets: security.NoSecrets,
			ExpectedConf: `
<label @ES_1>
  # Viaq Data Model
  <filter **>
    @type viaq_data_model
    enable_openshift_model false
    enable_prune_empty_fields false
    rename_time false
    undefined_dot_replace_char UNUSED
    elasticsearch_index_prefix_field 'viaq_index_name'
    <elasticsearch_index_name>
      enabled 'true'
      tag "kubernetes.var.log.pods.openshift_** kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_** var.log.pods.openshift_** var.log.pods.openshift-*_** var.log.pods.default_** var.log.pods.kube-*_** journal.system** system.var.log**"
      name_type static
      static_index_name infra-write
    </elasticsearch_index_name>
    <elasticsearch_index_name>
      enabled 'true'
      tag "linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**"
      name_type static
      static_index_name audit-write
    </elasticsearch_index_name>
    <elasticsearch_index_name>
      enabled 'true'
      tag "**"
      name_type structured
      static_index_name app-write
    </elasticsearch_index_name>
  </filter>
  <filter **>
    @type viaq_data_model
    enable_prune_labels true
    enable_openshift_model false
    rename_time false
    undefined_dot_replace_char UNUSED
    prune_labels_exclusions app.kubernetes.io/name,app.kubernetes.io/instance,app.kubernetes.io/version,app.kubernetes.io/component,app.kubernetes.io/part-of,app.kubernetes.io/managed-by,app.kubernetes.io/created-by
  </filter>
  
  #remove structured field if present
  <filter **>
    @type record_modifier
    remove_keys structured
  </filter>
  
  <match retry_es_1>
    @type elasticsearch
    @id retry_es_1
    host es.svc.infra.cluster
    port 9999
    verify_es_version_at_startup false
    scheme http
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    http_backend typhoeus
    write_operation create
    # https://github.com/uken/fluent-plugin-elasticsearch#suppress_type_name
    suppress_type_name 'true'
    reload_connections 'true'
    # https://github.com/uken/fluent-plugin-elasticsearch#reload-after
    reload_after '200'
    # https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
    sniffer_class_name 'Fluent::Plugin::ElasticsearchSimpleSniffer'
    reload_on_failure false
    # 2 ^ 31
    request_timeout 2147483648
    <buffer>
      @type file
      path '/var/lib/fluentd/retry_es_1'
      flush_mode interval
      flush_interval 1s
      flush_thread_count 2
      retry_type exponential_backoff
      retry_wait 1s
      retry_max_interval 60s
      retry_timeout 60m
      queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
      total_limit_size "#{ENV['TOTAL_LIMIT_SIZE_PER_BUFFER'] || '8589934592'}"
      chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
      overflow_action block
      disable_chunk_backup true
    </buffer>
  </match>
  
  <match **>
    @type elasticsearch
    @id es_1
    host es.svc.infra.cluster
    port 9999
    verify_es_version_at_startup false
    scheme http
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    retry_tag retry_es_1
    http_backend typhoeus
    write_operation create
    # https://github.com/uken/fluent-plugin-elasticsearch#suppress_type_name
    suppress_type_name 'true'
    reload_connections 'true'
    # https://github.com/uken/fluent-plugin-elasticsearch#reload-after
    reload_after '200'
    # https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
    sniffer_class_name 'Fluent::Plugin::ElasticsearchSimpleSniffer'
    reload_on_failure false
    # 2 ^ 31
    request_timeout 2147483648
    <buffer>
      @type file
      path '/var/lib/fluentd/es_1'
      flush_mode interval
      flush_interval 1s
      flush_thread_count 2
      retry_type exponential_backoff
      retry_wait 1s
      retry_max_interval 60s
      retry_timeout 60m
      queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
      total_limit_size "#{ENV['TOTAL_LIMIT_SIZE_PER_BUFFER'] || '8589934592'}"
      chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
      overflow_action block
      disable_chunk_backup true
    </buffer>
  </match>
</label>
`,
		}),
		Entry("with https but without a secret", testhelpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeElasticsearch,
						Name: "es-1",
						URL:  "https://es.svc.infra.cluster:9999",
					},
				},
			},
			Secrets: security.NoSecrets,
			ExpectedConf: `
<label @ES_1>
  # Viaq Data Model
  <filter **>
    @type viaq_data_model
    enable_openshift_model false
    enable_prune_empty_fields false
    rename_time false
    undefined_dot_replace_char UNUSED
    elasticsearch_index_prefix_field 'viaq_index_name'
    <elasticsearch_index_name>
      enabled 'true'
      tag "kubernetes.var.log.pods.openshift_** kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_** var.log.pods.openshift_** var.log.pods.openshift-*_** var.log.pods.default_** var.log.pods.kube-*_** journal.system** system.var.log**"
      name_type static
      static_index_name infra-write
    </elasticsearch_index_name>
    <elasticsearch_index_name>
      enabled 'true'
      tag "linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**"
      name_type static
      static_index_name audit-write
    </elasticsearch_index_name>
    <elasticsearch_index_name>
      enabled 'true'
      tag "**"
      name_type structured
      static_index_name app-write
    </elasticsearch_index_name>
  </filter>
  <filter **>
    @type viaq_data_model
    enable_prune_labels true
    enable_openshift_model false
    rename_time false
    undefined_dot_replace_char UNUSED
    prune_labels_exclusions app.kubernetes.io/name,app.kubernetes.io/instance,app.kubernetes.io/version,app.kubernetes.io/component,app.kubernetes.io/part-of,app.kubernetes.io/managed-by,app.kubernetes.io/created-by
  </filter>
  
  #remove structured field if present
  <filter **>
    @type record_modifier
    remove_keys structured
  </filter>
  
  <match retry_es_1>
    @type elasticsearch
    @id retry_es_1
    host es.svc.infra.cluster
    port 9999
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    http_backend typhoeus
    write_operation create
    # https://github.com/uken/fluent-plugin-elasticsearch#suppress_type_name
    suppress_type_name 'true'
    reload_connections 'true'
    # https://github.com/uken/fluent-plugin-elasticsearch#reload-after
    reload_after '200'
    # https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
    sniffer_class_name 'Fluent::Plugin::ElasticsearchSimpleSniffer'
    reload_on_failure false
    # 2 ^ 31
    request_timeout 2147483648
    <buffer>
      @type file
      path '/var/lib/fluentd/retry_es_1'
      flush_mode interval
      flush_interval 1s
      flush_thread_count 2
      retry_type exponential_backoff
      retry_wait 1s
      retry_max_interval 60s
      retry_timeout 60m
      queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
      total_limit_size "#{ENV['TOTAL_LIMIT_SIZE_PER_BUFFER'] || '8589934592'}"
      chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
      overflow_action block
      disable_chunk_backup true
    </buffer>
  </match>
  
  <match **>
    @type elasticsearch
    @id es_1
    host es.svc.infra.cluster
    port 9999
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    retry_tag retry_es_1
    http_backend typhoeus
    write_operation create
    # https://github.com/uken/fluent-plugin-elasticsearch#suppress_type_name
    suppress_type_name 'true'
    reload_connections 'true'
    # https://github.com/uken/fluent-plugin-elasticsearch#reload-after
    reload_after '200'
    # https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
    sniffer_class_name 'Fluent::Plugin::ElasticsearchSimpleSniffer'
    reload_on_failure false
    # 2 ^ 31
    request_timeout 2147483648
    <buffer>
      @type file
      path '/var/lib/fluentd/es_1'
      flush_mode interval
      flush_interval 1s
      flush_thread_count 2
      retry_type exponential_backoff
      retry_wait 1s
      retry_max_interval 60s
      retry_timeout 60m
      queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
      total_limit_size "#{ENV['TOTAL_LIMIT_SIZE_PER_BUFFER'] || '8589934592'}"
      chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
      overflow_action block
      disable_chunk_backup true
    </buffer>
  </match>
</label>
`,
		}),
		Entry("with http and username/password", testhelpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeElasticsearch,
						Name: "es-1",
						URL:  "http://es.svc.infra.cluster:9999",
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
<label @ES_1>
  # Viaq Data Model
  <filter **>
    @type viaq_data_model
    enable_openshift_model false
    enable_prune_empty_fields false
    rename_time false
    undefined_dot_replace_char UNUSED
    elasticsearch_index_prefix_field 'viaq_index_name'
    <elasticsearch_index_name>
      enabled 'true'
      tag "kubernetes.var.log.pods.openshift_** kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_** var.log.pods.openshift_** var.log.pods.openshift-*_** var.log.pods.default_** var.log.pods.kube-*_** journal.system** system.var.log**"
      name_type static
      static_index_name infra-write
    </elasticsearch_index_name>
    <elasticsearch_index_name>
      enabled 'true'
      tag "linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**"
      name_type static
      static_index_name audit-write
    </elasticsearch_index_name>
    <elasticsearch_index_name>
      enabled 'true'
      tag "**"
      name_type structured
      static_index_name app-write
    </elasticsearch_index_name>
  </filter>
  <filter **>
    @type viaq_data_model
    enable_prune_labels true
    enable_openshift_model false
    rename_time false
    undefined_dot_replace_char UNUSED
    prune_labels_exclusions app.kubernetes.io/name,app.kubernetes.io/instance,app.kubernetes.io/version,app.kubernetes.io/component,app.kubernetes.io/part-of,app.kubernetes.io/managed-by,app.kubernetes.io/created-by
  </filter>
  
  #remove structured field if present
  <filter **>
    @type record_modifier
    remove_keys structured
  </filter>
  
  <match retry_es_1>
    @type elasticsearch
    @id retry_es_1
    host es.svc.infra.cluster
    port 9999
    verify_es_version_at_startup false
    scheme http
    user "#{File.exists?('/var/run/ocp-collector/secrets/es-1/username') ? open('/var/run/ocp-collector/secrets/es-1/username','r') do |f|f.read end : ''}"
    password "#{File.exists?('/var/run/ocp-collector/secrets/es-1/password') ? open('/var/run/ocp-collector/secrets/es-1/password','r') do |f|f.read end : ''}"
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    http_backend typhoeus
    write_operation create
    # https://github.com/uken/fluent-plugin-elasticsearch#suppress_type_name
    suppress_type_name 'true'
    reload_connections 'true'
    # https://github.com/uken/fluent-plugin-elasticsearch#reload-after
    reload_after '200'
    # https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
    sniffer_class_name 'Fluent::Plugin::ElasticsearchSimpleSniffer'
    reload_on_failure false
    # 2 ^ 31
    request_timeout 2147483648
    <buffer>
      @type file
      path '/var/lib/fluentd/retry_es_1'
      flush_mode interval
      flush_interval 1s
      flush_thread_count 2
      retry_type exponential_backoff
      retry_wait 1s
      retry_max_interval 60s
      retry_timeout 60m
      queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
      total_limit_size "#{ENV['TOTAL_LIMIT_SIZE_PER_BUFFER'] || '8589934592'}"
      chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
      overflow_action block
      disable_chunk_backup true
    </buffer>
  </match>
  
  <match **>
    @type elasticsearch
    @id es_1
    host es.svc.infra.cluster
    port 9999
    verify_es_version_at_startup false
    scheme http
    user "#{File.exists?('/var/run/ocp-collector/secrets/es-1/username') ? open('/var/run/ocp-collector/secrets/es-1/username','r') do |f|f.read end : ''}"
    password "#{File.exists?('/var/run/ocp-collector/secrets/es-1/password') ? open('/var/run/ocp-collector/secrets/es-1/password','r') do |f|f.read end : ''}"
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    retry_tag retry_es_1
    http_backend typhoeus
    write_operation create
    # https://github.com/uken/fluent-plugin-elasticsearch#suppress_type_name
    suppress_type_name 'true'
    reload_connections 'true'
    # https://github.com/uken/fluent-plugin-elasticsearch#reload-after
    reload_after '200'
    # https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
    sniffer_class_name 'Fluent::Plugin::ElasticsearchSimpleSniffer'
    reload_on_failure false
    # 2 ^ 31
    request_timeout 2147483648
    <buffer>
      @type file
      path '/var/lib/fluentd/es_1'
      flush_mode interval
      flush_interval 1s
      flush_thread_count 2
      retry_type exponential_backoff
      retry_wait 1s
      retry_max_interval 60s
      retry_timeout 60m
      queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
      total_limit_size "#{ENV['TOTAL_LIMIT_SIZE_PER_BUFFER'] || '8589934592'}"
      chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
      overflow_action block
      disable_chunk_backup true
    </buffer>
  </match>
</label>
`,
		}),
		Entry("with structured container logs enabled", testhelpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeElasticsearch,
						Name: "es-1",
						URL:  "https://es.svc.infra.cluster:9999",
						OutputTypeSpec: logging.OutputTypeSpec{
							Elasticsearch: &logging.Elasticsearch{
								EnableStructuredContainerLogs: true,
							},
						},
					},
				},
			},
			Secrets: security.NoSecrets,
			ExpectedConf: `
<label @ES_1>
  # Viaq Data Model
  <filter **>
    @type viaq_data_model
    enable_openshift_model false
    enable_prune_empty_fields false
    rename_time false
    undefined_dot_replace_char UNUSED
    elasticsearch_index_prefix_field 'viaq_index_name'
    <elasticsearch_index_name>
      enabled 'true'
      tag "kubernetes.var.log.pods.openshift_** kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_** var.log.pods.openshift_** var.log.pods.openshift-*_** var.log.pods.default_** var.log.pods.kube-*_** journal.system** system.var.log**"
      name_type static
      static_index_name infra-write
      structured_type_annotation_prefix containerType.logging.openshift.io
    </elasticsearch_index_name>
    <elasticsearch_index_name>
      enabled 'true'
      tag "linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**"
      name_type static
      static_index_name audit-write
    </elasticsearch_index_name>
    <elasticsearch_index_name>
      enabled 'true'
      tag "**"
      name_type structured
      static_index_name app-write
      structured_type_annotation_prefix containerType.logging.openshift.io
    </elasticsearch_index_name>
  </filter>
  <filter **>
    @type viaq_data_model
    enable_prune_labels true
    enable_openshift_model false
    rename_time false
    undefined_dot_replace_char UNUSED
    prune_labels_exclusions app.kubernetes.io/name,app.kubernetes.io/instance,app.kubernetes.io/version,app.kubernetes.io/component,app.kubernetes.io/part-of,app.kubernetes.io/managed-by,app.kubernetes.io/created-by
  </filter>
  
  <match retry_es_1>
    @type elasticsearch
    @id retry_es_1
    host es.svc.infra.cluster
    port 9999
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    http_backend typhoeus
    write_operation create
    # https://github.com/uken/fluent-plugin-elasticsearch#suppress_type_name
    suppress_type_name 'true'
    reload_connections 'true'
    # https://github.com/uken/fluent-plugin-elasticsearch#reload-after
    reload_after '200'
    # https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
    sniffer_class_name 'Fluent::Plugin::ElasticsearchSimpleSniffer'
    reload_on_failure false
    # 2 ^ 31
    request_timeout 2147483648
    <buffer>
      @type file
      path '/var/lib/fluentd/retry_es_1'
      flush_mode interval
      flush_interval 1s
      flush_thread_count 2
      retry_type exponential_backoff
      retry_wait 1s
      retry_max_interval 60s
      retry_timeout 60m
      queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
      total_limit_size "#{ENV['TOTAL_LIMIT_SIZE_PER_BUFFER'] || '8589934592'}"
      chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
      overflow_action block
      disable_chunk_backup true
    </buffer>
  </match>
  
  <match **>
    @type elasticsearch
    @id es_1
    host es.svc.infra.cluster
    port 9999
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    retry_tag retry_es_1
    http_backend typhoeus
    write_operation create
    # https://github.com/uken/fluent-plugin-elasticsearch#suppress_type_name
    suppress_type_name 'true'
    reload_connections 'true'
    # https://github.com/uken/fluent-plugin-elasticsearch#reload-after
    reload_after '200'
    # https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
    sniffer_class_name 'Fluent::Plugin::ElasticsearchSimpleSniffer'
    reload_on_failure false
    # 2 ^ 31
    request_timeout 2147483648
    <buffer>
      @type file
      path '/var/lib/fluentd/es_1'
      flush_mode interval
      flush_interval 1s
      flush_thread_count 2
      retry_type exponential_backoff
      retry_wait 1s
      retry_max_interval 60s
      retry_timeout 60m
      queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
      total_limit_size "#{ENV['TOTAL_LIMIT_SIZE_PER_BUFFER'] || '8589934592'}"
      chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
      overflow_action block
      disable_chunk_backup true
    </buffer>
  </match>
</label>
`,
		}),
		Entry("with structured types enabled", testhelpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeElasticsearch,
						Name: "es-1",
						URL:  "https://es.svc.infra.cluster:9999",
						OutputTypeSpec: logging.OutputTypeSpec{
							Elasticsearch: &logging.Elasticsearch{
								StructuredTypeKey:  "kubernetes.labels.foo-bar",
								StructuredTypeName: "my-name",
							},
						},
					},
				},
			},
			Secrets: security.NoSecrets,
			ExpectedConf: `
<label @ES_1>
  # Viaq Data Model
  <filter **>
    @type viaq_data_model
    enable_openshift_model false
    enable_prune_empty_fields false
    rename_time false
    undefined_dot_replace_char UNUSED
    elasticsearch_index_prefix_field 'viaq_index_name'
    <elasticsearch_index_name>
      enabled 'true'
      tag "kubernetes.var.log.pods.openshift_** kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_** var.log.pods.openshift_** var.log.pods.openshift-*_** var.log.pods.default_** var.log.pods.kube-*_** journal.system** system.var.log**"
      name_type static
      static_index_name infra-write
      structured_type_key kubernetes.labels.foo-bar
      structured_type_name my-name
    </elasticsearch_index_name>
    <elasticsearch_index_name>
      enabled 'true'
      tag "linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**"
      name_type static
      static_index_name audit-write
    </elasticsearch_index_name>
    <elasticsearch_index_name>
      enabled 'true'
      tag "**"
      name_type structured
      static_index_name app-write
      structured_type_key kubernetes.labels.foo-bar
      structured_type_name my-name
    </elasticsearch_index_name>
  </filter>
  <filter **>
    @type viaq_data_model
    enable_prune_labels true
    enable_openshift_model false
    rename_time false
    undefined_dot_replace_char UNUSED
    prune_labels_exclusions app.kubernetes.io/name,app.kubernetes.io/instance,app.kubernetes.io/version,app.kubernetes.io/component,app.kubernetes.io/part-of,app.kubernetes.io/managed-by,app.kubernetes.io/created-by
  </filter>
  
  <match retry_es_1>
    @type elasticsearch
    @id retry_es_1
    host es.svc.infra.cluster
    port 9999
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    http_backend typhoeus
    write_operation create
    # https://github.com/uken/fluent-plugin-elasticsearch#suppress_type_name
    suppress_type_name 'true'
    reload_connections 'true'
    # https://github.com/uken/fluent-plugin-elasticsearch#reload-after
    reload_after '200'
    # https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
    sniffer_class_name 'Fluent::Plugin::ElasticsearchSimpleSniffer'
    reload_on_failure false
    # 2 ^ 31
    request_timeout 2147483648
    <buffer>
      @type file
      path '/var/lib/fluentd/retry_es_1'
      flush_mode interval
      flush_interval 1s
      flush_thread_count 2
      retry_type exponential_backoff
      retry_wait 1s
      retry_max_interval 60s
      retry_timeout 60m
      queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
      total_limit_size "#{ENV['TOTAL_LIMIT_SIZE_PER_BUFFER'] || '8589934592'}"
      chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
      overflow_action block
      disable_chunk_backup true
    </buffer>
  </match>
  
  <match **>
    @type elasticsearch
    @id es_1
    host es.svc.infra.cluster
    port 9999
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    retry_tag retry_es_1
    http_backend typhoeus
    write_operation create
    # https://github.com/uken/fluent-plugin-elasticsearch#suppress_type_name
    suppress_type_name 'true'
    reload_connections 'true'
    # https://github.com/uken/fluent-plugin-elasticsearch#reload-after
    reload_after '200'
    # https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
    sniffer_class_name 'Fluent::Plugin::ElasticsearchSimpleSniffer'
    reload_on_failure false
    # 2 ^ 31
    request_timeout 2147483648
    <buffer>
      @type file
      path '/var/lib/fluentd/es_1'
      flush_mode interval
      flush_interval 1s
      flush_thread_count 2
      retry_type exponential_backoff
      retry_wait 1s
      retry_max_interval 60s
      retry_timeout 60m
      queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
      total_limit_size "#{ENV['TOTAL_LIMIT_SIZE_PER_BUFFER'] || '8589934592'}"
      chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
      overflow_action block
      disable_chunk_backup true
    </buffer>
  </match>
</label>
`,
		}),
		Entry("with char encoding", testhelpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Outputs: []logging.OutputSpec{
					{
						Type: logging.OutputTypeElasticsearch,
						Name: "es-1",
						URL:  "http://es.svc.infra.cluster:9999",
						Secret: &logging.OutputSecretSpec{
							Name: "es-1",
						},
					},
				},
			},
			Secrets: security.NoSecrets,
			Options: map[string]interface{}{elements.CharEncoding: elements.DefaultCharEncoding},
			ExpectedConf: `
<label @ES_1>
  # Viaq Data Model
  <filter **>
    @type viaq_data_model
    enable_openshift_model false
    enable_prune_empty_fields false
    rename_time false
    undefined_dot_replace_char UNUSED
    elasticsearch_index_prefix_field 'viaq_index_name'
    <elasticsearch_index_name>
      enabled 'true'
      tag "kubernetes.var.log.pods.openshift_** kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_** var.log.pods.openshift_** var.log.pods.openshift-*_** var.log.pods.default_** var.log.pods.kube-*_** journal.system** system.var.log**"
      name_type static
      static_index_name infra-write
    </elasticsearch_index_name>
    <elasticsearch_index_name>
      enabled 'true'
      tag "linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**"
      name_type static
      static_index_name audit-write
    </elasticsearch_index_name>
    <elasticsearch_index_name>
      enabled 'true'
      tag "**"
      name_type structured
      static_index_name app-write
    </elasticsearch_index_name>
  </filter>
  <filter **>
    @type viaq_data_model
    enable_prune_labels true
    enable_openshift_model false
    rename_time false
    undefined_dot_replace_char UNUSED
    prune_labels_exclusions app.kubernetes.io/name,app.kubernetes.io/instance,app.kubernetes.io/version,app.kubernetes.io/component,app.kubernetes.io/part-of,app.kubernetes.io/managed-by,app.kubernetes.io/created-by
  </filter>
  
  #remove structured field if present
  <filter **>
    @type record_modifier
    char_encoding utf-8:utf-8
    remove_keys structured
  </filter>
  
  <match retry_es_1>
    @type elasticsearch
    @id retry_es_1
    host es.svc.infra.cluster
    port 9999
    verify_es_version_at_startup false
    scheme http
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    http_backend typhoeus
    write_operation create
    # https://github.com/uken/fluent-plugin-elasticsearch#suppress_type_name
    suppress_type_name 'true'
    reload_connections 'true'
    # https://github.com/uken/fluent-plugin-elasticsearch#reload-after
    reload_after '200'
    # https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
    sniffer_class_name 'Fluent::Plugin::ElasticsearchSimpleSniffer'
    reload_on_failure false
    # 2 ^ 31
    request_timeout 2147483648
    <buffer>
      @type file
      path '/var/lib/fluentd/retry_es_1'
      flush_mode interval
      flush_interval 1s
      flush_thread_count 2
      retry_type exponential_backoff
      retry_wait 1s
      retry_max_interval 60s
      retry_timeout 60m
      queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
      total_limit_size "#{ENV['TOTAL_LIMIT_SIZE_PER_BUFFER'] || '8589934592'}"
      chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
      overflow_action block
      disable_chunk_backup true
    </buffer>
  </match>
  
  <match **>
    @type elasticsearch
    @id es_1
    host es.svc.infra.cluster
    port 9999
    verify_es_version_at_startup false
    scheme http
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    retry_tag retry_es_1
    http_backend typhoeus
    write_operation create
    # https://github.com/uken/fluent-plugin-elasticsearch#suppress_type_name
    suppress_type_name 'true'
    reload_connections 'true'
    # https://github.com/uken/fluent-plugin-elasticsearch#reload-after
    reload_after '200'
    # https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
    sniffer_class_name 'Fluent::Plugin::ElasticsearchSimpleSniffer'
    reload_on_failure false
    # 2 ^ 31
    request_timeout 2147483648
    <buffer>
      @type file
      path '/var/lib/fluentd/es_1'
      flush_mode interval
      flush_interval 1s
      flush_thread_count 2
      retry_type exponential_backoff
      retry_wait 1s
      retry_max_interval 60s
      retry_timeout 60m
      queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32'}"
      total_limit_size "#{ENV['TOTAL_LIMIT_SIZE_PER_BUFFER'] || '8589934592'}"
      chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
      overflow_action block
      disable_chunk_backup true
    </buffer>
  </match>
</label>
`,
		}),
	)
})
