package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	"github.com/openshift/cluster-logging-operator/internal/generator/fluentd/output/security"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Generating fluentd config", func() {
	var (
		forwarder  *logging.ClusterLogForwarderSpec
		g          generator.Generator
		op         generator.Options
		secret     corev1.Secret
		secretData = map[string][]byte{
			"tls.key":       []byte("test-key"),
			"tls.crt":       []byte("test-crt"),
			"ca-bundle.crt": []byte("test-bundle"),
		}
		secrets = map[string]*corev1.Secret{
			"infra-es": {
				ObjectMeta: v1.ObjectMeta{
					Name: "my-infra-secret",
				},
				Data: secretData,
			},
			"apps-es-1": {
				ObjectMeta: v1.ObjectMeta{
					Name: "my-es-secret",
				},
				Data: secretData,
			},
			"apps-es-2": {
				ObjectMeta: v1.ObjectMeta{
					Name: "my-other-secret",
				},
				Data: secretData,
			},
			"audit-es": {
				ObjectMeta: v1.ObjectMeta{
					Name: "my-audit-secret",
				},
				Data: secretData,
			},
		}
	)
	BeforeEach(func() {
		g = generator.MakeGenerator()
		op = generator.NoOptions
		forwarder = &logging.ClusterLogForwarderSpec{
			Outputs: []logging.OutputSpec{
				{
					Type:   logging.OutputTypeElasticsearch,
					Name:   "infra-es",
					URL:    "https://es.svc.infra.cluster:9999",
					Secret: &logging.OutputSecretSpec{Name: "my-infra-secret"},
				},
				{
					Type:   logging.OutputTypeElasticsearch,
					Name:   "apps-es-1",
					URL:    "https://es.svc.messaging.cluster.local:9654",
					Secret: &logging.OutputSecretSpec{Name: "my-es-secret"},
				},
				{
					Type: logging.OutputTypeElasticsearch,
					Name: "apps-es-2",
					URL:  "https://es.svc.messaging.cluster.local2:9654",
					Secret: &logging.OutputSecretSpec{
						Name: "my-other-secret",
					},
				},
				{
					Type: logging.OutputTypeElasticsearch,
					Name: "audit-es",
					URL:  "https://es.svc.audit.cluster:9654",
					Secret: &logging.OutputSecretSpec{
						Name: "my-audit-secret",
					},
				},
			},
			Pipelines: []logging.PipelineSpec{
				{
					Name:       "infra-pipeline",
					InputRefs:  []string{logging.InputNameInfrastructure},
					OutputRefs: []string{"infra-es"},
				},
				{
					Name:       "apps-pipeline",
					InputRefs:  []string{logging.InputNameApplication},
					OutputRefs: []string{"apps-es-1", "apps-es-2"},
				},
				{
					Name:       "audit-pipeline",
					InputRefs:  []string{logging.InputNameAudit},
					OutputRefs: []string{"audit-es"},
				},
			},
		}
		secret = corev1.Secret{
			Data: map[string][]byte{
				"tls.key":       []byte(""),
				"tls.crt":       []byte(""),
				"ca-bundle.crt": []byte(""),
			},
		}
		secrets = map[string]*corev1.Secret{
			"infra-es":  &secret,
			"apps-es-1": &secret,
			"apps-es-2": &secret,
			"audit-es":  &secret,
		}
	})

	It("should generate fluent config for sending specific namespace logs to separate pipelines", func() {
		forwarder = &logging.ClusterLogForwarderSpec{
			Outputs: []logging.OutputSpec{
				{
					Type: logging.OutputTypeElasticsearch,
					Name: "apps-es-1",
					URL:  "https://es.svc.messaging.cluster.local:9654",
					Secret: &logging.OutputSecretSpec{
						Name: "my-es-secret",
					},
				},
				{
					Type: logging.OutputTypeElasticsearch,
					Name: "apps-es-2",
					URL:  "https://es2.svc.messaging.cluster.local:9654",
					Secret: &logging.OutputSecretSpec{
						Name: "my-es-secret",
					},
				},
			},
			Inputs: []logging.InputSpec{
				{
					Name: "myInput",
					Application: &logging.Application{
						Namespaces: []string{"project1-namespace", "project2-namespace"},
					},
				},
				{
					Name: "myInput2",
					Application: &logging.Application{
						Namespaces: []string{"dev-apple", "project2-namespace"},
					},
				},
			},
			Pipelines: []logging.PipelineSpec{
				{
					Name:       "my-default-pipeline",
					InputRefs:  []string{"application"},
					OutputRefs: []string{"apps-es-2"},
				},
				{
					Name:       "apps-pipeline",
					InputRefs:  []string{"myInput"},
					OutputRefs: []string{"apps-es-1", "apps-es-2"},
				},
				{
					Name:       "apps-pipeline2",
					InputRefs:  []string{"myInput2", "application"},
					OutputRefs: []string{"apps-es-2"},
				},
			},
		}
		secrets = map[string]*corev1.Secret{
			"apps-es-1": {
				ObjectMeta: v1.ObjectMeta{
					Name: "my-es-secret",
				},
				Data: secretData,
			},
			"apps-es-2": {
				ObjectMeta: v1.ObjectMeta{
					Name: "my-es-secret",
				},
				Data: secretData,
			},
		}
		c := generator.MergeSections(Conf(nil, secrets, forwarder, op))
		results, err := g.GenerateConf(c...)
		Expect(err).To(BeNil())

		Expect(results).To(EqualTrimLines(`
## CLO GENERATED CONFIGURATION ###
# This file is a copy of the fluentd configuration entrypoint
# which should normally be supplied in a configmap.

<system>
  log_level "#{ENV['LOG_LEVEL'] || 'warn'}"
</system>

# Prometheus Monitoring
<source>
  @type prometheus
  bind "[::]"
  <transport tls>
    cert_path /etc/collector/metrics/tls.crt
    private_key_path /etc/collector/metrics/tls.key
  </transport>
</source>

<source>
  @type prometheus_monitor
  <labels>
    hostname ${hostname}
  </labels>
</source>

# excluding prometheus_tail_monitor
# since it leaks namespace/pod info
# via file paths

# tail_monitor plugin which publishes log_collected_bytes_total
<source>
  @type collected_tail_monitor
  <labels>
    hostname ${hostname}
  </labels>
</source>

# This is considered experimental by the repo
<source>
  @type prometheus_output_monitor
  <labels>
    hostname ${hostname}
  </labels>
</source>

# Logs from containers (including openshift containers)
<source>
  @type tail
  @id container-input
  path "/var/log/pods/*/*/*.log"
  exclude_path ["/var/log/pods/openshift-logging_collector-*/*/*.log", "/var/log/pods/openshift-logging_elasticsearch-*/*/*.log", "/var/log/pods/openshift-logging_kibana-*/*/*.log", "/var/log/pods/*/*/*.gz", "/var/log/pods/*/*/*.tmp"]
  pos_file "/var/lib/fluentd/pos/es-containers.log.pos"
  follow_inodes true
  refresh_interval 5
  rotate_wait 5
  tag kubernetes.*
  read_from_head "true"
  skip_refresh_on_startup true
  @label @CONCAT
  <parse>
    @type regexp
    expression /^(?<@timestamp>[^\s]+) (?<stream>stdout|stderr) (?<logtag>[F|P]) (?<message>.*)$/
    time_key '@timestamp'
    keep_time_key true
  </parse>
</source>

# Concat log lines of container logs, and send to INGRESS pipeline
<label @CONCAT>
  <filter kubernetes.**>
    @type concat
    key message
    partial_key logtag
    partial_value P
    separator ''
  </filter>
  
  <match kubernetes.**>
    @type relabel
    @label @INGRESS
  </match>
</label>

# Ingress pipeline
<label @INGRESS>
  # Filter out PRIORITY from journal logs
  <filter journal>
    @type grep
    <exclude>
      key PRIORITY
      pattern ^7$
    </exclude>
  </filter>
  
  # Process OVN logs
  <filter ovn-audit.log**>
    @type record_modifier
    <record>
      @timestamp ${DateTime.parse(record['message'].split('|')[0]).rfc3339(6)}
      level ${record['message'].split('|')[3].downcase}
    </record>
  </filter>
  
  # Retag Journal logs to specific tags
  <match journal>
    @type rewrite_tag_filter
    # skip to @INGRESS label section
    @label @INGRESS
  
    # see if this is a kibana container for special log handling
    # looks like this:
    # k8s_kibana.a67f366_logging-kibana-1-d90e3_logging_26c51a61-2835-11e6-ad29-fa163e4944d5_f0db49a2
    # we filter these logs through the kibana_transform.conf filter
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_kibana\.
      tag kubernetes.journal.container.kibana
    </rule>
  
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_logging-eventrouter-[^_]+_
      tag kubernetes.journal.container._default_.kubernetes-event
    </rule>
  
    # mark logs from default namespace for processing as k8s logs but stored as system logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_[^_]+_default_
      tag kubernetes.journal.container._default_
    </rule>
  
    # mark logs from kube-* namespaces for processing as k8s logs but stored as system logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_[^_]+_kube-(.+)_
      tag kubernetes.journal.container._kube-$1_
    </rule>
  
    # mark logs from openshift-* namespaces for processing as k8s logs but stored as system logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_[^_]+_openshift-(.+)_
      tag kubernetes.journal.container._openshift-$1_
    </rule>
  
    # mark logs from openshift namespace for processing as k8s logs but stored as system logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_[^_]+_openshift_
      tag kubernetes.journal.container._openshift_
    </rule>
  
    # mark fluentd container logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_.*fluentd
      tag kubernetes.journal.container.fluentd
    </rule>
  
    # this is a kubernetes container
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_
      tag kubernetes.journal.container
    </rule>
  
    # not kubernetes - assume a system log or system container log
    <rule>
      key _TRANSPORT
      pattern .+
      tag journal.system
    </rule>
  </match>
  
  # Invoke kubernetes apiserver to get kubernetes metadata
  <filter kubernetes.**>
	@id kubernetes-metadata
    @type kubernetes_metadata
    kubernetes_url 'https://kubernetes.default.svc'
    annotation_match ["^containerType\.logging\.openshift\.io\/.*$"]
    allow_orphans false
    cache_size '1000'
    ssl_partial_chain 'true'
  </filter>
  
  # Parse Json fields for container, journal and eventrouter logs
  <filter kubernetes.var.log.pods.**_eventrouter-**>
    @type parse_json_field
    merge_json_log true
    preserve_json_log true
    json_fields 'message'
  </filter>
  
  # Fix level field in audit logs
  <filter k8s-audit.log**>
    @type record_modifier
    <record>
      k8s_audit_level ${record['level']}
    </record>
  </filter>
  
  <filter openshift-audit.log**>
    @type record_modifier
    <record>
      openshift_audit_level ${record['level']}
    </record>
  </filter>
  
  # Viaq Data Model
  <filter **>
    @type viaq_data_model
    enable_flatten_labels true
    enable_prune_empty_fields false
    default_keep_fields CEE,time,@timestamp,aushape,ci_job,collectd,docker,fedora-ci,file,foreman,geoip,hostname,ipaddr4,ipaddr6,kubernetes,level,message,namespace_name,namespace_uuid,offset,openstack,ovirt,pid,pipeline_metadata,rsyslog,service,systemd,tags,testcase,tlog,viaq_msg_id
    keep_empty_fields 'message'
    rename_time true
    pipeline_type 'collector'
    process_kubernetes_events false
    <level>
      name warn
      match 'Warning|WARN|^W[0-9]+|level=warn|Value:warn|"level":"warn"'
    </level>
    <level>
      name info
      match 'Info|INFO|^I[0-9]+|level=info|Value:info|"level":"info"'
    </level>
    <level>
      name error
      match 'Error|ERROR|^E[0-9]+|level=error|Value:error|"level":"error"'
    </level>
    <level>
      name critical
      match 'Critical|CRITICAL|^C[0-9]+|level=critical|Value:critical|"level":"critical"'
    </level>
    <level>
      name debug
      match 'Debug|DEBUG|^D[0-9]+|level=debug|Value:debug|"level":"debug"'
    </level>
    <formatter>
      tag "journal.system**"
      type sys_journal
      remove_keys log,stream,MESSAGE,_SOURCE_REALTIME_TIMESTAMP,__REALTIME_TIMESTAMP,CONTAINER_ID,CONTAINER_ID_FULL,CONTAINER_NAME,PRIORITY,_BOOT_ID,_CAP_EFFECTIVE,_CMDLINE,_COMM,_EXE,_GID,_HOSTNAME,_MACHINE_ID,_PID,_SELINUX_CONTEXT,_SYSTEMD_CGROUP,_SYSTEMD_SLICE,_SYSTEMD_UNIT,_TRANSPORT,_UID,_AUDIT_LOGINUID,_AUDIT_SESSION,_SYSTEMD_OWNER_UID,_SYSTEMD_SESSION,_SYSTEMD_USER_UNIT,CODE_FILE,CODE_FUNCTION,CODE_LINE,ERRNO,MESSAGE_ID,RESULT,UNIT,_KERNEL_DEVICE,_KERNEL_SUBSYSTEM,_UDEV_SYSNAME,_UDEV_DEVNODE,_UDEV_DEVLINK,SYSLOG_FACILITY,SYSLOG_IDENTIFIER,SYSLOG_PID
    </formatter>
    <formatter>
      tag "kubernetes.var.log.pods.**_eventrouter-** k8s-audit.log** openshift-audit.log** ovn-audit.log**"
      type k8s_json_file
      remove_keys stream
      process_kubernetes_events 'true'
    </formatter>
    <formatter>
      tag "kubernetes.var.log.pods**"
      type k8s_json_file
      remove_keys stream
    </formatter>
  </filter>
  
  # Generate elasticsearch id
  <filter **>
    @type elasticsearch_genid_ext
    hash_id_key viaq_msg_id
    alt_key kubernetes.event.metadata.uid
    alt_tags 'kubernetes.var.log.pods.**_eventrouter-*.** kubernetes.journal.container._default_.kubernetes-event'
  </filter>
  
  # Discard Infrastructure logs
  <match kubernetes.var.log.pods.openshift_** kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_** journal.** system.var.log**>
    @type null
  </match>
  
  # Include Application logs
  <match kubernetes.**>
    @type relabel
    @label @_APPLICATION
  </match>
  
  # Discard Audit logs
  <match linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**>
    @type null
  </match>
  
  # Send any remaining unmatched tags to stdout
  <match **>
   @type stdout
  </match>
</label>

# Routing Application to pipelines
<label @_APPLICATION>
  <filter **>
    @type record_modifier
    <record>
      log_type application
    </record>
  </filter>
  
  <match **>
    @type label_router
    <route>
      @label @APPS_PIPELINE
      <match>
        namespaces project1-namespace, project2-namespace
      </match>
    </route>
    
    <route>
      @label @APPS_PIPELINE2
      <match>
        namespaces dev-apple, project2-namespace
      </match>
    </route>
    
    <route>
      @label @_APPLICATION_ALL
      <match>
      
      </match>
    </route>
  </match>
</label>

# Copying unrouted application to pipelines
<label @_APPLICATION_ALL>
  <match **>
    @type copy
    <store>
      @type relabel
      @label @MY_DEFAULT_PIPELINE
    </store>
    
    <store>
      @type relabel
      @label @APPS_PIPELINE2
    </store>
  </match>
</label>

# Copying pipeline apps-pipeline to outputs
<label @APPS_PIPELINE>
  <match **>
    @type copy
    copy_mode deep
    <store>
      @type relabel
      @label @APPS_ES_1
    </store>
    
    <store>
      @type relabel
      @label @APPS_ES_2
    </store>
  </match>
</label>

# Copying pipeline apps-pipeline2 to outputs
<label @APPS_PIPELINE2>
  <match **>
    @type relabel
    @label @APPS_ES_2
  </match>
</label>

# Copying pipeline my-default-pipeline to outputs
<label @MY_DEFAULT_PIPELINE>
  <match **>
    @type relabel
    @label @APPS_ES_2
  </match>
</label>

# Ship logs to specific outputs
<label @APPS_ES_1>
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
  
  <match retry_apps_es_1>
    @type elasticsearch
    @id retry_apps_es_1
    host es.svc.messaging.cluster.local
    port 9654
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    client_key '/var/run/ocp-collector/secrets/my-es-secret/tls.key'
    client_cert '/var/run/ocp-collector/secrets/my-es-secret/tls.crt'
    ca_file '/var/run/ocp-collector/secrets/my-es-secret/ca-bundle.crt'
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
      path '/var/lib/fluentd/retry_apps_es_1'
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
    @id apps_es_1
    host es.svc.messaging.cluster.local
    port 9654
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    client_key '/var/run/ocp-collector/secrets/my-es-secret/tls.key'
    client_cert '/var/run/ocp-collector/secrets/my-es-secret/tls.crt'
    ca_file '/var/run/ocp-collector/secrets/my-es-secret/ca-bundle.crt'
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    retry_tag retry_apps_es_1
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
      path '/var/lib/fluentd/apps_es_1'
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

<label @APPS_ES_2>
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
  
  <match retry_apps_es_2>
    @type elasticsearch
    @id retry_apps_es_2
    host es2.svc.messaging.cluster.local
    port 9654
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    client_key '/var/run/ocp-collector/secrets/my-es-secret/tls.key'
    client_cert '/var/run/ocp-collector/secrets/my-es-secret/tls.crt'
    ca_file '/var/run/ocp-collector/secrets/my-es-secret/ca-bundle.crt'
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
      path '/var/lib/fluentd/retry_apps_es_2'
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
    @id apps_es_2
    host es2.svc.messaging.cluster.local
    port 9654
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    client_key '/var/run/ocp-collector/secrets/my-es-secret/tls.key'
    client_cert '/var/run/ocp-collector/secrets/my-es-secret/tls.crt'
    ca_file '/var/run/ocp-collector/secrets/my-es-secret/ca-bundle.crt'
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    retry_tag retry_apps_es_2
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
      path '/var/lib/fluentd/apps_es_2'
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
`))
	})

	It("should generate fluent config for sending only logs from pods identified by labels to output", func() {
		forwarder = &logging.ClusterLogForwarderSpec{
			Outputs: []logging.OutputSpec{
				{
					Type:   logging.OutputTypeElasticsearch,
					Name:   "apps-es-1",
					URL:    "https://es.svc.messaging.cluster.local:9654",
					Secret: &logging.OutputSecretSpec{Name: "my-es-secret"},
				},
				{
					Type:   logging.OutputTypeElasticsearch,
					Name:   "apps-es-2",
					URL:    "https://es.svc.messaging.cluster.local2:9654",
					Secret: &logging.OutputSecretSpec{Name: "my-other-secret"},
				},
			},
			Inputs: []logging.InputSpec{
				{
					Name: "myInput-1",
					Application: &logging.Application{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"environment": "production",
								"app":         "nginx",
							},
						},
					},
				},
				{
					Name: "myInput-2",
					Application: &logging.Application{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"environment": "dev",
								"app":         "nginx",
							},
						},
					},
				},
			},
			Pipelines: []logging.PipelineSpec{
				{
					Name:       "apps-prod-pipeline",
					InputRefs:  []string{"myInput-1"},
					OutputRefs: []string{"apps-es-1"},
				},
				{
					Name:       "apps-dev-pipeline",
					InputRefs:  []string{"myInput-2"},
					OutputRefs: []string{"apps-es-2"},
				},
				{
					Name:       "apps-default",
					InputRefs:  []string{logging.InputNameApplication},
					OutputRefs: []string{"default"},
				},
			},
		}
		secrets = map[string]*corev1.Secret{
			"apps-es-1": {
				ObjectMeta: v1.ObjectMeta{
					Name: "my-es-secret",
				},
				Data: secretData,
			},
			"apps-es-2": {
				ObjectMeta: v1.ObjectMeta{
					Name: "my-other-secret",
				},
				Data: secretData,
			},
		}
		c := generator.MergeSections(Conf(nil, secrets, forwarder, op))
		results, err := g.GenerateConf(c...)
		Expect(err).To(BeNil())
		Expect(results).To(EqualTrimLines(`
## CLO GENERATED CONFIGURATION ###
# This file is a copy of the fluentd configuration entrypoint
# which should normally be supplied in a configmap.

<system>
  log_level "#{ENV['LOG_LEVEL'] || 'warn'}"
</system>

# Prometheus Monitoring
<source>
  @type prometheus
  bind "[::]"
  <transport tls>
    cert_path /etc/collector/metrics/tls.crt
    private_key_path /etc/collector/metrics/tls.key
  </transport>
</source>

<source>
  @type prometheus_monitor
  <labels>
    hostname ${hostname}
  </labels>
</source>

# excluding prometheus_tail_monitor
# since it leaks namespace/pod info
# via file paths

# tail_monitor plugin which publishes log_collected_bytes_total
<source>
  @type collected_tail_monitor
  <labels>
    hostname ${hostname}
  </labels>
</source>

# This is considered experimental by the repo
<source>
  @type prometheus_output_monitor
  <labels>
    hostname ${hostname}
  </labels>
</source>

# Logs from containers (including openshift containers)
<source>
  @type tail
  @id container-input
  path "/var/log/pods/*/*/*.log"
  exclude_path ["/var/log/pods/openshift-logging_collector-*/*/*.log", "/var/log/pods/openshift-logging_elasticsearch-*/*/*.log", "/var/log/pods/openshift-logging_kibana-*/*/*.log", "/var/log/pods/*/*/*.gz", "/var/log/pods/*/*/*.tmp"]
  pos_file "/var/lib/fluentd/pos/es-containers.log.pos"
  follow_inodes true
  refresh_interval 5
  rotate_wait 5
  tag kubernetes.*
  read_from_head "true"
  skip_refresh_on_startup true
  @label @CONCAT
  <parse>
    @type regexp
    expression /^(?<@timestamp>[^\s]+) (?<stream>stdout|stderr) (?<logtag>[F|P]) (?<message>.*)$/
    time_key '@timestamp'
    keep_time_key true
  </parse>
</source>

# Concat log lines of container logs, and send to INGRESS pipeline
<label @CONCAT>
  <filter kubernetes.**>
    @type concat
    key message
    partial_key logtag
    partial_value P
    separator ''
  </filter>
  
  <match kubernetes.**>
    @type relabel
    @label @INGRESS
  </match>
</label>

# Ingress pipeline
<label @INGRESS>
  # Filter out PRIORITY from journal logs
  <filter journal>
    @type grep
    <exclude>
      key PRIORITY
      pattern ^7$
    </exclude>
  </filter>
  
  # Process OVN logs
  <filter ovn-audit.log**>
    @type record_modifier
    <record>
      @timestamp ${DateTime.parse(record['message'].split('|')[0]).rfc3339(6)}
      level ${record['message'].split('|')[3].downcase}
    </record>
  </filter>
  
  # Retag Journal logs to specific tags
  <match journal>
    @type rewrite_tag_filter
    # skip to @INGRESS label section
    @label @INGRESS
  
    # see if this is a kibana container for special log handling
    # looks like this:
    # k8s_kibana.a67f366_logging-kibana-1-d90e3_logging_26c51a61-2835-11e6-ad29-fa163e4944d5_f0db49a2
    # we filter these logs through the kibana_transform.conf filter
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_kibana\.
      tag kubernetes.journal.container.kibana
    </rule>
  
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_logging-eventrouter-[^_]+_
      tag kubernetes.journal.container._default_.kubernetes-event
    </rule>
  
    # mark logs from default namespace for processing as k8s logs but stored as system logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_[^_]+_default_
      tag kubernetes.journal.container._default_
    </rule>
  
    # mark logs from kube-* namespaces for processing as k8s logs but stored as system logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_[^_]+_kube-(.+)_
      tag kubernetes.journal.container._kube-$1_
    </rule>
  
    # mark logs from openshift-* namespaces for processing as k8s logs but stored as system logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_[^_]+_openshift-(.+)_
      tag kubernetes.journal.container._openshift-$1_
    </rule>
  
    # mark logs from openshift namespace for processing as k8s logs but stored as system logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_[^_]+_openshift_
      tag kubernetes.journal.container._openshift_
    </rule>
  
    # mark fluentd container logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_.*fluentd
      tag kubernetes.journal.container.fluentd
    </rule>
  
    # this is a kubernetes container
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_
      tag kubernetes.journal.container
    </rule>
  
    # not kubernetes - assume a system log or system container log
    <rule>
      key _TRANSPORT
      pattern .+
      tag journal.system
    </rule>
  </match>
  
  # Invoke kubernetes apiserver to get kubernetes metadata
  <filter kubernetes.**>
	@id kubernetes-metadata
    @type kubernetes_metadata
    kubernetes_url 'https://kubernetes.default.svc'
	annotation_match ["^containerType\.logging\.openshift\.io\/.*$"]
    allow_orphans false
    cache_size '1000'
    ssl_partial_chain 'true'
  </filter>
  
  # Parse Json fields for container, journal and eventrouter logs
  <filter kubernetes.var.log.pods.**_eventrouter-**>
    @type parse_json_field
    merge_json_log true
    preserve_json_log true
    json_fields 'message'
  </filter>
  
  # Fix level field in audit logs
  <filter k8s-audit.log**>
    @type record_modifier
    <record>
      k8s_audit_level ${record['level']}
    </record>
  </filter>
  
  <filter openshift-audit.log**>
    @type record_modifier
    <record>
      openshift_audit_level ${record['level']}
    </record>
  </filter>
  
  # Viaq Data Model
  <filter **>
    @type viaq_data_model
    enable_flatten_labels true
    enable_prune_empty_fields false
    default_keep_fields CEE,time,@timestamp,aushape,ci_job,collectd,docker,fedora-ci,file,foreman,geoip,hostname,ipaddr4,ipaddr6,kubernetes,level,message,namespace_name,namespace_uuid,offset,openstack,ovirt,pid,pipeline_metadata,rsyslog,service,systemd,tags,testcase,tlog,viaq_msg_id
    keep_empty_fields 'message'
    rename_time true
    pipeline_type 'collector'
    process_kubernetes_events false
    <level>
      name warn
      match 'Warning|WARN|^W[0-9]+|level=warn|Value:warn|"level":"warn"'
    </level>
    <level>
      name info
      match 'Info|INFO|^I[0-9]+|level=info|Value:info|"level":"info"'
    </level>
    <level>
      name error
      match 'Error|ERROR|^E[0-9]+|level=error|Value:error|"level":"error"'
    </level>
    <level>
      name critical
      match 'Critical|CRITICAL|^C[0-9]+|level=critical|Value:critical|"level":"critical"'
    </level>
    <level>
      name debug
      match 'Debug|DEBUG|^D[0-9]+|level=debug|Value:debug|"level":"debug"'
    </level>
    <formatter>
      tag "journal.system**"
      type sys_journal
      remove_keys log,stream,MESSAGE,_SOURCE_REALTIME_TIMESTAMP,__REALTIME_TIMESTAMP,CONTAINER_ID,CONTAINER_ID_FULL,CONTAINER_NAME,PRIORITY,_BOOT_ID,_CAP_EFFECTIVE,_CMDLINE,_COMM,_EXE,_GID,_HOSTNAME,_MACHINE_ID,_PID,_SELINUX_CONTEXT,_SYSTEMD_CGROUP,_SYSTEMD_SLICE,_SYSTEMD_UNIT,_TRANSPORT,_UID,_AUDIT_LOGINUID,_AUDIT_SESSION,_SYSTEMD_OWNER_UID,_SYSTEMD_SESSION,_SYSTEMD_USER_UNIT,CODE_FILE,CODE_FUNCTION,CODE_LINE,ERRNO,MESSAGE_ID,RESULT,UNIT,_KERNEL_DEVICE,_KERNEL_SUBSYSTEM,_UDEV_SYSNAME,_UDEV_DEVNODE,_UDEV_DEVLINK,SYSLOG_FACILITY,SYSLOG_IDENTIFIER,SYSLOG_PID
    </formatter>
    <formatter>
      tag "kubernetes.var.log.pods.**_eventrouter-** k8s-audit.log** openshift-audit.log** ovn-audit.log**"
      type k8s_json_file
      remove_keys stream
      process_kubernetes_events 'true'
    </formatter>
    <formatter>
      tag "kubernetes.var.log.pods**"
      type k8s_json_file
      remove_keys stream
    </formatter>
  </filter>
  
  # Generate elasticsearch id
  <filter **>
    @type elasticsearch_genid_ext
    hash_id_key viaq_msg_id
    alt_key kubernetes.event.metadata.uid
    alt_tags 'kubernetes.var.log.pods.**_eventrouter-*.** kubernetes.journal.container._default_.kubernetes-event'
  </filter>
  
  # Discard Infrastructure logs
  <match kubernetes.var.log.pods.openshift_** kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_** journal.** system.var.log**>
    @type null
  </match>
  
  # Include Application logs
  <match kubernetes.**>
    @type relabel
    @label @_APPLICATION
  </match>
  
  # Discard Audit logs
  <match linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**>
    @type null
  </match>
  
  # Send any remaining unmatched tags to stdout
  <match **>
   @type stdout
  </match>
</label>

# Routing Application to pipelines
<label @_APPLICATION>
  <filter **>
    @type record_modifier
    <record>
      log_type application
    </record>
  </filter>
  
  <match **>
    @type label_router
    <route>
      @label @APPS_PROD_PIPELINE
      <match>
        labels app:nginx, environment:production
      </match>
    </route>
    
    <route>
      @label @APPS_DEV_PIPELINE
      <match>
        labels app:nginx, environment:dev
      </match>
    </route>
    
    <route>
      @label @_APPLICATION_ALL
      <match>
      
      </match>
    </route>
  </match>
</label>

# Sending unrouted application to pipelines
<label @_APPLICATION_ALL>
  <match **>
    @type relabel
    @label @APPS_DEFAULT
  </match>
</label>

# Copying pipeline apps-default to outputs
<label @APPS_DEFAULT>
  <match **>
    @type relabel
    @label @DEFAULT
  </match>
</label>

# Copying pipeline apps-dev-pipeline to outputs
<label @APPS_DEV_PIPELINE>
  <match **>
    @type relabel
    @label @APPS_ES_2
  </match>
</label>

# Copying pipeline apps-prod-pipeline to outputs
<label @APPS_PROD_PIPELINE>
  <match **>
    @type relabel
    @label @APPS_ES_1
  </match>
</label>

# Ship logs to specific outputs
<label @APPS_ES_1>
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
  
  <match retry_apps_es_1>
    @type elasticsearch
    @id retry_apps_es_1
    host es.svc.messaging.cluster.local
    port 9654
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    client_key '/var/run/ocp-collector/secrets/my-es-secret/tls.key'
    client_cert '/var/run/ocp-collector/secrets/my-es-secret/tls.crt'
    ca_file '/var/run/ocp-collector/secrets/my-es-secret/ca-bundle.crt'
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
      path '/var/lib/fluentd/retry_apps_es_1'
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
    @id apps_es_1
    host es.svc.messaging.cluster.local
    port 9654
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    client_key '/var/run/ocp-collector/secrets/my-es-secret/tls.key'
    client_cert '/var/run/ocp-collector/secrets/my-es-secret/tls.crt'
    ca_file '/var/run/ocp-collector/secrets/my-es-secret/ca-bundle.crt'
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    retry_tag retry_apps_es_1
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
      path '/var/lib/fluentd/apps_es_1'
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

<label @APPS_ES_2>
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
  
  <match retry_apps_es_2>
    @type elasticsearch
    @id retry_apps_es_2
    host es.svc.messaging.cluster.local2
    port 9654
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    client_key '/var/run/ocp-collector/secrets/my-other-secret/tls.key'
    client_cert '/var/run/ocp-collector/secrets/my-other-secret/tls.crt'
    ca_file '/var/run/ocp-collector/secrets/my-other-secret/ca-bundle.crt'
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
      path '/var/lib/fluentd/retry_apps_es_2'
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
    @id apps_es_2
    host es.svc.messaging.cluster.local2
    port 9654
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    client_key '/var/run/ocp-collector/secrets/my-other-secret/tls.key'
    client_cert '/var/run/ocp-collector/secrets/my-other-secret/tls.crt'
    ca_file '/var/run/ocp-collector/secrets/my-other-secret/ca-bundle.crt'
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    retry_tag retry_apps_es_2
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
      path '/var/lib/fluentd/apps_es_2'
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
`))
	})

	It("should generate fluent config for sending only logs from pods identified by labels and namespaces to output", func() {
		forwarder = &logging.ClusterLogForwarderSpec{
			Outputs: []logging.OutputSpec{
				{
					Type:   logging.OutputTypeElasticsearch,
					Name:   "apps-es-1",
					URL:    "https://es.svc.messaging.cluster.local:9654",
					Secret: &logging.OutputSecretSpec{Name: "my-es-secret"},
				},
				{
					Type:   logging.OutputTypeElasticsearch,
					Name:   "apps-es-2",
					URL:    "https://es.svc.messaging.cluster.local2:9654",
					Secret: &logging.OutputSecretSpec{Name: "my-other-secret"},
				},
			},
			Inputs: []logging.InputSpec{
				{
					Name: "myInput-1",
					Application: &logging.Application{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"environment": "production",
								"app":         "nginx",
							},
						},
						Namespaces: []string{"project1-namespace"},
					},
				},
				{
					Name: "myInput-2",
					Application: &logging.Application{
						Selector: &metav1.LabelSelector{
							MatchLabels: map[string]string{
								"environment": "dev",
								"app":         "nginx",
							},
						},
						Namespaces: []string{"project2-namespace"},
					},
				},
			},
			Pipelines: []logging.PipelineSpec{
				{
					Name:       "apps-prod-pipeline",
					InputRefs:  []string{"myInput-1"},
					OutputRefs: []string{"apps-es-1"},
				},
				{
					Name:       "apps-dev-pipeline",
					InputRefs:  []string{"myInput-2"},
					OutputRefs: []string{"apps-es-2"},
				},
				{
					Name:       "apps-default",
					InputRefs:  []string{logging.InputNameApplication},
					OutputRefs: []string{"default"},
				},
			},
		}
		secrets = map[string]*corev1.Secret{
			"apps-es-1": {
				ObjectMeta: v1.ObjectMeta{
					Name: "my-es-secret",
				},
				Data: secretData,
			},
			"apps-es-2": {
				ObjectMeta: v1.ObjectMeta{
					Name: "my-other-secret",
				},
				Data: secretData,
			},
		}
		c := generator.MergeSections(Conf(nil, secrets, forwarder, op))
		results, err := g.GenerateConf(c...)
		Expect(err).To(BeNil())
		Expect(results).To(EqualTrimLines(`
## CLO GENERATED CONFIGURATION ###
# This file is a copy of the fluentd configuration entrypoint
# which should normally be supplied in a configmap.

<system>
  log_level "#{ENV['LOG_LEVEL'] || 'warn'}"
</system>

# Prometheus Monitoring
<source>
  @type prometheus
  bind "[::]"
  <transport tls>
    cert_path /etc/collector/metrics/tls.crt
    private_key_path /etc/collector/metrics/tls.key
  </transport>
</source>

<source>
  @type prometheus_monitor
  <labels>
    hostname ${hostname}
  </labels>
</source>

# excluding prometheus_tail_monitor
# since it leaks namespace/pod info
# via file paths

# tail_monitor plugin which publishes log_collected_bytes_total
<source>
  @type collected_tail_monitor
  <labels>
    hostname ${hostname}
  </labels>
</source>

# This is considered experimental by the repo
<source>
  @type prometheus_output_monitor
  <labels>
    hostname ${hostname}
  </labels>
</source>

# Logs from containers (including openshift containers)
<source>
  @type tail
  @id container-input
  path "/var/log/pods/*/*/*.log"
  exclude_path ["/var/log/pods/openshift-logging_collector-*/*/*.log", "/var/log/pods/openshift-logging_elasticsearch-*/*/*.log", "/var/log/pods/openshift-logging_kibana-*/*/*.log", "/var/log/pods/*/*/*.gz", "/var/log/pods/*/*/*.tmp"]
  pos_file "/var/lib/fluentd/pos/es-containers.log.pos"
  follow_inodes true
  refresh_interval 5
  rotate_wait 5
  tag kubernetes.*
  read_from_head "true"
  skip_refresh_on_startup true
  @label @CONCAT
  <parse>
    @type regexp
    expression /^(?<@timestamp>[^\s]+) (?<stream>stdout|stderr) (?<logtag>[F|P]) (?<message>.*)$/
    time_key '@timestamp'
    keep_time_key true
  </parse>
</source>

# Concat log lines of container logs, and send to INGRESS pipeline
<label @CONCAT>
  <filter kubernetes.**>
    @type concat
    key message
    partial_key logtag
    partial_value P
    separator ''
  </filter>
  
  <match kubernetes.**>
    @type relabel
    @label @INGRESS
  </match>
</label>

# Ingress pipeline
<label @INGRESS>
  # Filter out PRIORITY from journal logs
  <filter journal>
    @type grep
    <exclude>
      key PRIORITY
      pattern ^7$
    </exclude>
  </filter>
  
  # Process OVN logs
  <filter ovn-audit.log**>
    @type record_modifier
    <record>
      @timestamp ${DateTime.parse(record['message'].split('|')[0]).rfc3339(6)}
      level ${record['message'].split('|')[3].downcase}
    </record>
  </filter>
  
  # Retag Journal logs to specific tags
  <match journal>
    @type rewrite_tag_filter
    # skip to @INGRESS label section
    @label @INGRESS
  
    # see if this is a kibana container for special log handling
    # looks like this:
    # k8s_kibana.a67f366_logging-kibana-1-d90e3_logging_26c51a61-2835-11e6-ad29-fa163e4944d5_f0db49a2
    # we filter these logs through the kibana_transform.conf filter
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_kibana\.
      tag kubernetes.journal.container.kibana
    </rule>
  
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_logging-eventrouter-[^_]+_
      tag kubernetes.journal.container._default_.kubernetes-event
    </rule>
  
    # mark logs from default namespace for processing as k8s logs but stored as system logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_[^_]+_default_
      tag kubernetes.journal.container._default_
    </rule>
  
    # mark logs from kube-* namespaces for processing as k8s logs but stored as system logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_[^_]+_kube-(.+)_
      tag kubernetes.journal.container._kube-$1_
    </rule>
  
    # mark logs from openshift-* namespaces for processing as k8s logs but stored as system logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_[^_]+_openshift-(.+)_
      tag kubernetes.journal.container._openshift-$1_
    </rule>
  
    # mark logs from openshift namespace for processing as k8s logs but stored as system logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_[^_]+_openshift_
      tag kubernetes.journal.container._openshift_
    </rule>
  
    # mark fluentd container logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_.*fluentd
      tag kubernetes.journal.container.fluentd
    </rule>
  
    # this is a kubernetes container
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_
      tag kubernetes.journal.container
    </rule>
  
    # not kubernetes - assume a system log or system container log
    <rule>
      key _TRANSPORT
      pattern .+
      tag journal.system
    </rule>
  </match>
  
  # Invoke kubernetes apiserver to get kubernetes metadata
  <filter kubernetes.**>
	@id kubernetes-metadata
    @type kubernetes_metadata
    kubernetes_url 'https://kubernetes.default.svc'
	annotation_match ["^containerType\.logging\.openshift\.io\/.*$"]
    allow_orphans false
    cache_size '1000'
    ssl_partial_chain 'true'
  </filter>
  
  # Parse Json fields for container, journal and eventrouter logs
  <filter kubernetes.var.log.pods.**_eventrouter-**>
    @type parse_json_field
    merge_json_log true
    preserve_json_log true
    json_fields 'message'
  </filter>
  
  # Fix level field in audit logs
  <filter k8s-audit.log**>
    @type record_modifier
    <record>
      k8s_audit_level ${record['level']}
    </record>
  </filter>
  
  <filter openshift-audit.log**>
    @type record_modifier
    <record>
      openshift_audit_level ${record['level']}
    </record>
  </filter>
  
  # Viaq Data Model
  <filter **>
    @type viaq_data_model
    enable_flatten_labels true
    enable_prune_empty_fields false
    default_keep_fields CEE,time,@timestamp,aushape,ci_job,collectd,docker,fedora-ci,file,foreman,geoip,hostname,ipaddr4,ipaddr6,kubernetes,level,message,namespace_name,namespace_uuid,offset,openstack,ovirt,pid,pipeline_metadata,rsyslog,service,systemd,tags,testcase,tlog,viaq_msg_id
    keep_empty_fields 'message'
    rename_time true
    pipeline_type 'collector'
    process_kubernetes_events false
    <level>
      name warn
      match 'Warning|WARN|^W[0-9]+|level=warn|Value:warn|"level":"warn"'
    </level>
    <level>
      name info
      match 'Info|INFO|^I[0-9]+|level=info|Value:info|"level":"info"'
    </level>
    <level>
      name error
      match 'Error|ERROR|^E[0-9]+|level=error|Value:error|"level":"error"'
    </level>
    <level>
      name critical
      match 'Critical|CRITICAL|^C[0-9]+|level=critical|Value:critical|"level":"critical"'
    </level>
    <level>
      name debug
      match 'Debug|DEBUG|^D[0-9]+|level=debug|Value:debug|"level":"debug"'
    </level>
    <formatter>
      tag "journal.system**"
      type sys_journal
      remove_keys log,stream,MESSAGE,_SOURCE_REALTIME_TIMESTAMP,__REALTIME_TIMESTAMP,CONTAINER_ID,CONTAINER_ID_FULL,CONTAINER_NAME,PRIORITY,_BOOT_ID,_CAP_EFFECTIVE,_CMDLINE,_COMM,_EXE,_GID,_HOSTNAME,_MACHINE_ID,_PID,_SELINUX_CONTEXT,_SYSTEMD_CGROUP,_SYSTEMD_SLICE,_SYSTEMD_UNIT,_TRANSPORT,_UID,_AUDIT_LOGINUID,_AUDIT_SESSION,_SYSTEMD_OWNER_UID,_SYSTEMD_SESSION,_SYSTEMD_USER_UNIT,CODE_FILE,CODE_FUNCTION,CODE_LINE,ERRNO,MESSAGE_ID,RESULT,UNIT,_KERNEL_DEVICE,_KERNEL_SUBSYSTEM,_UDEV_SYSNAME,_UDEV_DEVNODE,_UDEV_DEVLINK,SYSLOG_FACILITY,SYSLOG_IDENTIFIER,SYSLOG_PID
    </formatter>
    <formatter>
      tag "kubernetes.var.log.pods.**_eventrouter-** k8s-audit.log** openshift-audit.log** ovn-audit.log**"
      type k8s_json_file
      remove_keys stream
      process_kubernetes_events 'true'
    </formatter>
    <formatter>
      tag "kubernetes.var.log.pods**"
      type k8s_json_file
      remove_keys stream
    </formatter>
  </filter>
  
  # Generate elasticsearch id
  <filter **>
    @type elasticsearch_genid_ext
    hash_id_key viaq_msg_id
    alt_key kubernetes.event.metadata.uid
    alt_tags 'kubernetes.var.log.pods.**_eventrouter-*.** kubernetes.journal.container._default_.kubernetes-event'
  </filter>
  
  # Discard Infrastructure logs
  <match kubernetes.var.log.pods.openshift_** kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_** journal.** system.var.log**>
    @type null
  </match>
  
  # Include Application logs
  <match kubernetes.**>
    @type relabel
    @label @_APPLICATION
  </match>
  
  # Discard Audit logs
  <match linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**>
    @type null
  </match>
  
  # Send any remaining unmatched tags to stdout
  <match **>
   @type stdout
  </match>
</label>

# Routing Application to pipelines
<label @_APPLICATION>
  <filter **>
    @type record_modifier
    <record>
      log_type application
    </record>
  </filter>
  
  <match **>
    @type label_router
    <route>
      @label @APPS_PROD_PIPELINE
      <match>
        namespaces project1-namespace
        labels app:nginx, environment:production
      </match>
    </route>
    
    <route>
      @label @APPS_DEV_PIPELINE
      <match>
        namespaces project2-namespace
        labels app:nginx, environment:dev
      </match>
    </route>
    
    <route>
      @label @_APPLICATION_ALL
      <match>
      
      </match>
    </route>
  </match>
</label>

# Sending unrouted application to pipelines
<label @_APPLICATION_ALL>
  <match **>
    @type relabel
    @label @APPS_DEFAULT
  </match>
</label>

# Copying pipeline apps-default to outputs
<label @APPS_DEFAULT>
  <match **>
    @type relabel
    @label @DEFAULT
  </match>
</label>

# Copying pipeline apps-dev-pipeline to outputs
<label @APPS_DEV_PIPELINE>
  <match **>
    @type relabel
    @label @APPS_ES_2
  </match>
</label>

# Copying pipeline apps-prod-pipeline to outputs
<label @APPS_PROD_PIPELINE>
  <match **>
    @type relabel
    @label @APPS_ES_1
  </match>
</label>

# Ship logs to specific outputs
<label @APPS_ES_1>
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
  
  <match retry_apps_es_1>
    @type elasticsearch
    @id retry_apps_es_1
    host es.svc.messaging.cluster.local
    port 9654
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    client_key '/var/run/ocp-collector/secrets/my-es-secret/tls.key'
    client_cert '/var/run/ocp-collector/secrets/my-es-secret/tls.crt'
    ca_file '/var/run/ocp-collector/secrets/my-es-secret/ca-bundle.crt'
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
      path '/var/lib/fluentd/retry_apps_es_1'
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
    @id apps_es_1
    host es.svc.messaging.cluster.local
    port 9654
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    client_key '/var/run/ocp-collector/secrets/my-es-secret/tls.key'
    client_cert '/var/run/ocp-collector/secrets/my-es-secret/tls.crt'
    ca_file '/var/run/ocp-collector/secrets/my-es-secret/ca-bundle.crt'
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    retry_tag retry_apps_es_1
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
      path '/var/lib/fluentd/apps_es_1'
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

<label @APPS_ES_2>
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
  
  <match retry_apps_es_2>
    @type elasticsearch
    @id retry_apps_es_2
    host es.svc.messaging.cluster.local2
    port 9654
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    client_key '/var/run/ocp-collector/secrets/my-other-secret/tls.key'
    client_cert '/var/run/ocp-collector/secrets/my-other-secret/tls.crt'
    ca_file '/var/run/ocp-collector/secrets/my-other-secret/ca-bundle.crt'
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
      path '/var/lib/fluentd/retry_apps_es_2'
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
    @id apps_es_2
    host es.svc.messaging.cluster.local2
    port 9654
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    client_key '/var/run/ocp-collector/secrets/my-other-secret/tls.key'
    client_cert '/var/run/ocp-collector/secrets/my-other-secret/tls.crt'
    ca_file '/var/run/ocp-collector/secrets/my-other-secret/ca-bundle.crt'
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    retry_tag retry_apps_es_2
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
      path '/var/lib/fluentd/apps_es_2'
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
`))
	})

	It("should exclude source to pipeline labels when there are no pipelines for a given sourceType (e.g. only logs.app)", func() {
		forwarder = &logging.ClusterLogForwarderSpec{
			Outputs: []logging.OutputSpec{
				{
					Type: "fluentdForward",
					Name: "secureforward-receiver",
					URL:  "http://es.svc.messaging.cluster.local:9654",
				},
			},
			Pipelines: []logging.PipelineSpec{
				{
					Name:       "apps-pipeline",
					InputRefs:  []string{logging.InputNameApplication},
					OutputRefs: []string{"secureforward-receiver"},
				},
			},
		}
		c := generator.MergeSections(Conf(nil, nil, forwarder, op))
		results, err := g.GenerateConf(c...)
		Expect(err).To(BeNil())
		Expect(results).To(EqualTrimLines(`
## CLO GENERATED CONFIGURATION ###
# This file is a copy of the fluentd configuration entrypoint
# which should normally be supplied in a configmap.

<system>
  log_level "#{ENV['LOG_LEVEL'] || 'warn'}"
</system>

# Prometheus Monitoring
<source>
  @type prometheus
  bind "[::]"
  <transport tls>
    cert_path /etc/collector/metrics/tls.crt
    private_key_path /etc/collector/metrics/tls.key
  </transport>
</source>

<source>
  @type prometheus_monitor
  <labels>
    hostname ${hostname}
  </labels>
</source>

# excluding prometheus_tail_monitor
# since it leaks namespace/pod info
# via file paths

# tail_monitor plugin which publishes log_collected_bytes_total
<source>
  @type collected_tail_monitor
  <labels>
    hostname ${hostname}
  </labels>
</source>

# This is considered experimental by the repo
<source>
  @type prometheus_output_monitor
  <labels>
    hostname ${hostname}
  </labels>
</source>

# Logs from containers (including openshift containers)
<source>
  @type tail
  @id container-input
  path "/var/log/pods/*/*/*.log"
  exclude_path ["/var/log/pods/openshift-logging_collector-*/*/*.log", "/var/log/pods/openshift-logging_elasticsearch-*/*/*.log", "/var/log/pods/openshift-logging_kibana-*/*/*.log", "/var/log/pods/*/*/*.gz", "/var/log/pods/*/*/*.tmp"]
  pos_file "/var/lib/fluentd/pos/es-containers.log.pos"
  follow_inodes true
  refresh_interval 5
  rotate_wait 5
  tag kubernetes.*
  read_from_head "true"
  skip_refresh_on_startup true
  @label @CONCAT
  <parse>
    @type regexp
    expression /^(?<@timestamp>[^\s]+) (?<stream>stdout|stderr) (?<logtag>[F|P]) (?<message>.*)$/
    time_key '@timestamp'
    keep_time_key true
  </parse>
</source>

# Concat log lines of container logs, and send to INGRESS pipeline
<label @CONCAT>
  <filter kubernetes.**>
    @type concat
    key message
    partial_key logtag
    partial_value P
    separator ''
  </filter>
  
  <match kubernetes.**>
    @type relabel
    @label @INGRESS
  </match>
</label>

# Ingress pipeline
<label @INGRESS>
  # Filter out PRIORITY from journal logs
  <filter journal>
    @type grep
    <exclude>
      key PRIORITY
      pattern ^7$
    </exclude>
  </filter>
  
  # Process OVN logs
  <filter ovn-audit.log**>
    @type record_modifier
    <record>
      @timestamp ${DateTime.parse(record['message'].split('|')[0]).rfc3339(6)}
      level ${record['message'].split('|')[3].downcase}
    </record>
  </filter>
  
  # Retag Journal logs to specific tags
  <match journal>
    @type rewrite_tag_filter
    # skip to @INGRESS label section
    @label @INGRESS
  
    # see if this is a kibana container for special log handling
    # looks like this:
    # k8s_kibana.a67f366_logging-kibana-1-d90e3_logging_26c51a61-2835-11e6-ad29-fa163e4944d5_f0db49a2
    # we filter these logs through the kibana_transform.conf filter
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_kibana\.
      tag kubernetes.journal.container.kibana
    </rule>
  
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_logging-eventrouter-[^_]+_
      tag kubernetes.journal.container._default_.kubernetes-event
    </rule>
  
    # mark logs from default namespace for processing as k8s logs but stored as system logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_[^_]+_default_
      tag kubernetes.journal.container._default_
    </rule>
  
    # mark logs from kube-* namespaces for processing as k8s logs but stored as system logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_[^_]+_kube-(.+)_
      tag kubernetes.journal.container._kube-$1_
    </rule>
  
    # mark logs from openshift-* namespaces for processing as k8s logs but stored as system logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_[^_]+_openshift-(.+)_
      tag kubernetes.journal.container._openshift-$1_
    </rule>
  
    # mark logs from openshift namespace for processing as k8s logs but stored as system logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_[^_]+_openshift_
      tag kubernetes.journal.container._openshift_
    </rule>
  
    # mark fluentd container logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_.*fluentd
      tag kubernetes.journal.container.fluentd
    </rule>
  
    # this is a kubernetes container
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_
      tag kubernetes.journal.container
    </rule>
  
    # not kubernetes - assume a system log or system container log
    <rule>
      key _TRANSPORT
      pattern .+
      tag journal.system
    </rule>
  </match>
  
  # Invoke kubernetes apiserver to get kubernetes metadata
  <filter kubernetes.**>
	@id kubernetes-metadata
    @type kubernetes_metadata
    kubernetes_url 'https://kubernetes.default.svc'
	annotation_match ["^containerType\.logging\.openshift\.io\/.*$"]
    allow_orphans false
    cache_size '1000'
    ssl_partial_chain 'true'
  </filter>
  
  # Parse Json fields for container, journal and eventrouter logs
  <filter kubernetes.var.log.pods.**_eventrouter-**>
    @type parse_json_field
    merge_json_log true
    preserve_json_log true
    json_fields 'message'
  </filter>
  
  # Fix level field in audit logs
  <filter k8s-audit.log**>
    @type record_modifier
    <record>
      k8s_audit_level ${record['level']}
    </record>
  </filter>
  
  <filter openshift-audit.log**>
    @type record_modifier
    <record>
      openshift_audit_level ${record['level']}
    </record>
  </filter>
  
  # Viaq Data Model
  <filter **>
    @type viaq_data_model
    enable_flatten_labels true
    enable_prune_empty_fields false
    default_keep_fields CEE,time,@timestamp,aushape,ci_job,collectd,docker,fedora-ci,file,foreman,geoip,hostname,ipaddr4,ipaddr6,kubernetes,level,message,namespace_name,namespace_uuid,offset,openstack,ovirt,pid,pipeline_metadata,rsyslog,service,systemd,tags,testcase,tlog,viaq_msg_id
    keep_empty_fields 'message'
    rename_time true
    pipeline_type 'collector'
    process_kubernetes_events false
    <level>
      name warn
      match 'Warning|WARN|^W[0-9]+|level=warn|Value:warn|"level":"warn"'
    </level>
    <level>
      name info
      match 'Info|INFO|^I[0-9]+|level=info|Value:info|"level":"info"'
    </level>
    <level>
      name error
      match 'Error|ERROR|^E[0-9]+|level=error|Value:error|"level":"error"'
    </level>
    <level>
      name critical
      match 'Critical|CRITICAL|^C[0-9]+|level=critical|Value:critical|"level":"critical"'
    </level>
    <level>
      name debug
      match 'Debug|DEBUG|^D[0-9]+|level=debug|Value:debug|"level":"debug"'
    </level>
    <formatter>
      tag "journal.system**"
      type sys_journal
      remove_keys log,stream,MESSAGE,_SOURCE_REALTIME_TIMESTAMP,__REALTIME_TIMESTAMP,CONTAINER_ID,CONTAINER_ID_FULL,CONTAINER_NAME,PRIORITY,_BOOT_ID,_CAP_EFFECTIVE,_CMDLINE,_COMM,_EXE,_GID,_HOSTNAME,_MACHINE_ID,_PID,_SELINUX_CONTEXT,_SYSTEMD_CGROUP,_SYSTEMD_SLICE,_SYSTEMD_UNIT,_TRANSPORT,_UID,_AUDIT_LOGINUID,_AUDIT_SESSION,_SYSTEMD_OWNER_UID,_SYSTEMD_SESSION,_SYSTEMD_USER_UNIT,CODE_FILE,CODE_FUNCTION,CODE_LINE,ERRNO,MESSAGE_ID,RESULT,UNIT,_KERNEL_DEVICE,_KERNEL_SUBSYSTEM,_UDEV_SYSNAME,_UDEV_DEVNODE,_UDEV_DEVLINK,SYSLOG_FACILITY,SYSLOG_IDENTIFIER,SYSLOG_PID
    </formatter>
    <formatter>
      tag "kubernetes.var.log.pods.**_eventrouter-** k8s-audit.log** openshift-audit.log** ovn-audit.log**"
      type k8s_json_file
      remove_keys stream
      process_kubernetes_events 'true'
    </formatter>
    <formatter>
      tag "kubernetes.var.log.pods**"
      type k8s_json_file
      remove_keys stream
    </formatter>
  </filter>
  
  # Generate elasticsearch id
  <filter **>
    @type elasticsearch_genid_ext
    hash_id_key viaq_msg_id
    alt_key kubernetes.event.metadata.uid
    alt_tags 'kubernetes.var.log.pods.**_eventrouter-*.** kubernetes.journal.container._default_.kubernetes-event'
  </filter>
  
  # Discard Infrastructure logs
  <match kubernetes.var.log.pods.openshift_** kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_** journal.** system.var.log**>
    @type null
  </match>
  
  # Include Application logs
  <match kubernetes.**>
    @type relabel
    @label @_APPLICATION
  </match>
  
  # Discard Audit logs
  <match linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**>
    @type null
  </match>
  
  # Send any remaining unmatched tags to stdout
  <match **>
   @type stdout
  </match>
</label>

# Sending application source type to pipeline
<label @_APPLICATION>
  <filter **>
    @type record_modifier
    <record>
      log_type application
    </record>
  </filter>
  
  <match **>
    @type relabel
    @label @APPS_PIPELINE
  </match>
</label>

# Copying pipeline apps-pipeline to outputs
<label @APPS_PIPELINE>
  <match **>
    @type relabel
    @label @SECUREFORWARD_RECEIVER
  </match>
</label>

# Ship logs to specific outputs
<label @SECUREFORWARD_RECEIVER>
  <match **>
    @type forward
    @id secureforward_receiver
    <server>
      host es.svc.messaging.cluster.local
      port 9654
    </server>
    heartbeat_type none
    keepalive true
    keepalive_timeout 30s
    
    <buffer>
      @type file
      path '/var/lib/fluentd/secureforward_receiver'
      flush_mode interval
      flush_interval 5s
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
`))
	})

	It("should produce well formed fluent.conf", func() {
		c := generator.MergeSections(Conf(nil, secrets, forwarder, op))
		results, err := g.GenerateConf(c...)
		Expect(err).To(BeNil())
		Expect(results).To(EqualTrimLines(`
## CLO GENERATED CONFIGURATION ###
# This file is a copy of the fluentd configuration entrypoint
# which should normally be supplied in a configmap.

<system>
  log_level "#{ENV['LOG_LEVEL'] || 'warn'}"
</system>

# Prometheus Monitoring
<source>
  @type prometheus
  bind "[::]"
  <transport tls>
    cert_path /etc/collector/metrics/tls.crt
    private_key_path /etc/collector/metrics/tls.key
  </transport>
</source>

<source>
  @type prometheus_monitor
  <labels>
    hostname ${hostname}
  </labels>
</source>

# excluding prometheus_tail_monitor
# since it leaks namespace/pod info
# via file paths

# tail_monitor plugin which publishes log_collected_bytes_total
<source>
  @type collected_tail_monitor
  <labels>
    hostname ${hostname}
  </labels>
</source>

# This is considered experimental by the repo
<source>
  @type prometheus_output_monitor
  <labels>
    hostname ${hostname}
  </labels>
</source>

# Logs from linux journal
<source>
  @type systemd
  @id systemd-input
  @label @INGRESS
  path '/var/log/journal'
  <storage>
    @type local
    persistent true
    # NOTE: if this does not end in .json, fluentd will think it
    # is the name of a directory - see fluentd storage_local.rb
    path '/var/lib/fluentd/pos/journal_pos.json'
  </storage>
  matches "#{ENV['JOURNAL_FILTERS_JSON'] || '[]'}"
  tag journal
  read_from_head "#{if (val = ENV.fetch('JOURNAL_READ_FROM_HEAD','')) && (val.length > 0); val; else 'false'; end}"
</source>

# Logs from containers (including openshift containers)
<source>
  @type tail
  @id container-input
  path "/var/log/pods/*/*/*.log"
  exclude_path ["/var/log/pods/openshift-logging_collector-*/*/*.log", "/var/log/pods/openshift-logging_elasticsearch-*/*/*.log", "/var/log/pods/openshift-logging_kibana-*/*/*.log", "/var/log/pods/*/*/*.gz", "/var/log/pods/*/*/*.tmp"]
  pos_file "/var/lib/fluentd/pos/es-containers.log.pos"
  follow_inodes true
  refresh_interval 5
  rotate_wait 5
  tag kubernetes.*
  read_from_head "true"
  skip_refresh_on_startup true
  @label @CONCAT
  <parse>
    @type regexp
    expression /^(?<@timestamp>[^\s]+) (?<stream>stdout|stderr) (?<logtag>[F|P]) (?<message>.*)$/
    time_key '@timestamp'
    keep_time_key true
  </parse>
</source>

# linux audit logs
<source>
  @type tail
  @id audit-input
  @label @INGRESS
  path "/var/log/audit/audit.log"
  pos_file "/var/lib/fluentd/pos/audit.log.pos"
  follow_inodes true
  tag linux-audit.log
  <parse>
    @type viaq_host_audit
  </parse>
</source>

# k8s audit logs
<source>
  @type tail
  @id k8s-audit-input
  @label @INGRESS
  path "/var/log/kube-apiserver/audit.log"
  pos_file "/var/lib/fluentd/pos/kube-apiserver.audit.log.pos"
  follow_inodes true
  tag k8s-audit.log
  <parse>
    @type json
    time_key requestReceivedTimestamp
    # In case folks want to parse based on the requestReceivedTimestamp key
    keep_time_key true
    time_format %Y-%m-%dT%H:%M:%S.%N%z
  </parse>
</source>

# Openshift audit logs
<source>
  @type tail
  @id openshift-audit-input
  @label @INGRESS
  path /var/log/oauth-apiserver/audit.log,/var/log/openshift-apiserver/audit.log
  pos_file /var/lib/fluentd/pos/oauth-apiserver.audit.log
  follow_inodes true
  tag openshift-audit.log
  <parse>
    @type json
    time_key requestReceivedTimestamp
    # In case folks want to parse based on the requestReceivedTimestamp key
    keep_time_key true
    time_format %Y-%m-%dT%H:%M:%S.%N%z
  </parse>
</source>

# Openshift Virtual Network (OVN) audit logs
<source>
  @type tail
  @id ovn-audit-input
  @label @INGRESS
  path "/var/log/ovn/acl-audit-log.log"
  pos_file "/var/lib/fluentd/pos/acl-audit-log.pos"
  follow_inodes true
  tag ovn-audit.log
  refresh_interval 5
  rotate_wait 5
  read_from_head true
  <parse>
    @type none
  </parse>
</source>

# Concat log lines of container logs, and send to INGRESS pipeline
<label @CONCAT>
  <filter kubernetes.**>
    @type concat
    key message
    partial_key logtag
    partial_value P
    separator ''
  </filter>
  
  <match kubernetes.**>
    @type relabel
    @label @INGRESS
  </match>
</label>

# Ingress pipeline
<label @INGRESS>
  # Filter out PRIORITY from journal logs
  <filter journal>
    @type grep
    <exclude>
      key PRIORITY
      pattern ^7$
    </exclude>
  </filter>
  
  # Process OVN logs
  <filter ovn-audit.log**>
    @type record_modifier
    <record>
      @timestamp ${DateTime.parse(record['message'].split('|')[0]).rfc3339(6)}
      level ${record['message'].split('|')[3].downcase}
    </record>
  </filter>
  
  # Retag Journal logs to specific tags
  <match journal>
    @type rewrite_tag_filter
    # skip to @INGRESS label section
    @label @INGRESS
  
    # see if this is a kibana container for special log handling
    # looks like this:
    # k8s_kibana.a67f366_logging-kibana-1-d90e3_logging_26c51a61-2835-11e6-ad29-fa163e4944d5_f0db49a2
    # we filter these logs through the kibana_transform.conf filter
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_kibana\.
      tag kubernetes.journal.container.kibana
    </rule>
  
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_logging-eventrouter-[^_]+_
      tag kubernetes.journal.container._default_.kubernetes-event
    </rule>
  
    # mark logs from default namespace for processing as k8s logs but stored as system logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_[^_]+_default_
      tag kubernetes.journal.container._default_
    </rule>
  
    # mark logs from kube-* namespaces for processing as k8s logs but stored as system logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_[^_]+_kube-(.+)_
      tag kubernetes.journal.container._kube-$1_
    </rule>
  
    # mark logs from openshift-* namespaces for processing as k8s logs but stored as system logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_[^_]+_openshift-(.+)_
      tag kubernetes.journal.container._openshift-$1_
    </rule>
  
    # mark logs from openshift namespace for processing as k8s logs but stored as system logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_[^_]+_openshift_
      tag kubernetes.journal.container._openshift_
    </rule>
  
    # mark fluentd container logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_.*fluentd
      tag kubernetes.journal.container.fluentd
    </rule>
  
    # this is a kubernetes container
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_
      tag kubernetes.journal.container
    </rule>
  
    # not kubernetes - assume a system log or system container log
    <rule>
      key _TRANSPORT
      pattern .+
      tag journal.system
    </rule>
  </match>
  
  # Invoke kubernetes apiserver to get kubernetes metadata
  <filter kubernetes.**>
	@id kubernetes-metadata
    @type kubernetes_metadata
    kubernetes_url 'https://kubernetes.default.svc'
	annotation_match ["^containerType\.logging\.openshift\.io\/.*$"]
    allow_orphans false
    cache_size '1000'
    ssl_partial_chain 'true'
  </filter>
  
  # Parse Json fields for container, journal and eventrouter logs
  <filter kubernetes.var.log.pods.**_eventrouter-**>
    @type parse_json_field
    merge_json_log true
    preserve_json_log true
    json_fields 'message'
  </filter>
  
  # Fix level field in audit logs
  <filter k8s-audit.log**>
    @type record_modifier
    <record>
      k8s_audit_level ${record['level']}
    </record>
  </filter>
  
  <filter openshift-audit.log**>
    @type record_modifier
    <record>
      openshift_audit_level ${record['level']}
    </record>
  </filter>
  
  # Viaq Data Model
  <filter **>
    @type viaq_data_model
    enable_flatten_labels true
    enable_prune_empty_fields false
    default_keep_fields CEE,time,@timestamp,aushape,ci_job,collectd,docker,fedora-ci,file,foreman,geoip,hostname,ipaddr4,ipaddr6,kubernetes,level,message,namespace_name,namespace_uuid,offset,openstack,ovirt,pid,pipeline_metadata,rsyslog,service,systemd,tags,testcase,tlog,viaq_msg_id
    keep_empty_fields 'message'
    rename_time true
    pipeline_type 'collector'
    process_kubernetes_events false
    <level>
      name warn
      match 'Warning|WARN|^W[0-9]+|level=warn|Value:warn|"level":"warn"'
    </level>
    <level>
      name info
      match 'Info|INFO|^I[0-9]+|level=info|Value:info|"level":"info"'
    </level>
    <level>
      name error
      match 'Error|ERROR|^E[0-9]+|level=error|Value:error|"level":"error"'
    </level>
    <level>
      name critical
      match 'Critical|CRITICAL|^C[0-9]+|level=critical|Value:critical|"level":"critical"'
    </level>
    <level>
      name debug
      match 'Debug|DEBUG|^D[0-9]+|level=debug|Value:debug|"level":"debug"'
    </level>
    <formatter>
      tag "journal.system**"
      type sys_journal
      remove_keys log,stream,MESSAGE,_SOURCE_REALTIME_TIMESTAMP,__REALTIME_TIMESTAMP,CONTAINER_ID,CONTAINER_ID_FULL,CONTAINER_NAME,PRIORITY,_BOOT_ID,_CAP_EFFECTIVE,_CMDLINE,_COMM,_EXE,_GID,_HOSTNAME,_MACHINE_ID,_PID,_SELINUX_CONTEXT,_SYSTEMD_CGROUP,_SYSTEMD_SLICE,_SYSTEMD_UNIT,_TRANSPORT,_UID,_AUDIT_LOGINUID,_AUDIT_SESSION,_SYSTEMD_OWNER_UID,_SYSTEMD_SESSION,_SYSTEMD_USER_UNIT,CODE_FILE,CODE_FUNCTION,CODE_LINE,ERRNO,MESSAGE_ID,RESULT,UNIT,_KERNEL_DEVICE,_KERNEL_SUBSYSTEM,_UDEV_SYSNAME,_UDEV_DEVNODE,_UDEV_DEVLINK,SYSLOG_FACILITY,SYSLOG_IDENTIFIER,SYSLOG_PID
    </formatter>
    <formatter>
      tag "kubernetes.var.log.pods.**_eventrouter-** k8s-audit.log** openshift-audit.log** ovn-audit.log**"
      type k8s_json_file
      remove_keys stream
      process_kubernetes_events 'true'
    </formatter>
    <formatter>
      tag "kubernetes.var.log.pods**"
      type k8s_json_file
      remove_keys stream
    </formatter>
  </filter>
  
  # Generate elasticsearch id
  <filter **>
    @type elasticsearch_genid_ext
    hash_id_key viaq_msg_id
    alt_key kubernetes.event.metadata.uid
    alt_tags 'kubernetes.var.log.pods.**_eventrouter-*.** kubernetes.journal.container._default_.kubernetes-event'
  </filter>
  
  # Include Infrastructure logs
  <match kubernetes.var.log.pods.openshift_** kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_** journal.** system.var.log**>
    @type relabel
    @label @_INFRASTRUCTURE
  </match>
  
  # Include Application logs
  <match kubernetes.**>
    @type relabel
    @label @_APPLICATION
  </match>
  
  # Include Audit logs
  <match linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**>
    @type relabel
    @label @_AUDIT
  </match>
  
  # Send any remaining unmatched tags to stdout
  <match **>
   @type stdout
  </match>
</label>

# Sending application source type to pipeline
<label @_APPLICATION>
  <filter **>
    @type record_modifier
    <record>
      log_type application
    </record>
  </filter>
  
  <match **>
    @type relabel
    @label @APPS_PIPELINE
  </match>
</label>

# Sending infrastructure source type to pipeline
<label @_INFRASTRUCTURE>
  <filter **>
    @type record_modifier
    <record>
      log_type infrastructure
    </record>
  </filter>
  
  <match **>
    @type relabel
    @label @INFRA_PIPELINE
  </match>
</label>

# Sending audit source type to pipeline
<label @_AUDIT>
  <filter **>
    @type record_modifier
    <record>
      log_type audit
    </record>
  </filter>
  
  <match **>
    @type relabel
    @label @AUDIT_PIPELINE
  </match>
</label>

# Copying pipeline apps-pipeline to outputs
<label @APPS_PIPELINE>
  <match **>
    @type copy
	copy_mode deep
    <store>
      @type relabel
      @label @APPS_ES_1
    </store>
    
    <store>
      @type relabel
      @label @APPS_ES_2
    </store>
  </match>
</label>

# Copying pipeline audit-pipeline to outputs
<label @AUDIT_PIPELINE>
  <match **>
    @type relabel
    @label @AUDIT_ES
  </match>
</label>

# Copying pipeline infra-pipeline to outputs
<label @INFRA_PIPELINE>
  <match **>
    @type relabel
    @label @INFRA_ES
  </match>
</label>

# Ship logs to specific outputs
<label @INFRA_ES>
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
  
  <match retry_infra_es>
    @type elasticsearch
    @id retry_infra_es
    host es.svc.infra.cluster
    port 9999
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    client_key '/var/run/ocp-collector/secrets/my-infra-secret/tls.key'
    client_cert '/var/run/ocp-collector/secrets/my-infra-secret/tls.crt'
    ca_file '/var/run/ocp-collector/secrets/my-infra-secret/ca-bundle.crt'
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
      path '/var/lib/fluentd/retry_infra_es'
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
    @id infra_es
    host es.svc.infra.cluster
    port 9999
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    client_key '/var/run/ocp-collector/secrets/my-infra-secret/tls.key'
    client_cert '/var/run/ocp-collector/secrets/my-infra-secret/tls.crt'
    ca_file '/var/run/ocp-collector/secrets/my-infra-secret/ca-bundle.crt'
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    retry_tag retry_infra_es
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
      path '/var/lib/fluentd/infra_es'
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

<label @APPS_ES_1>
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
  
  <match retry_apps_es_1>
    @type elasticsearch
    @id retry_apps_es_1
    host es.svc.messaging.cluster.local
    port 9654
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    client_key '/var/run/ocp-collector/secrets/my-es-secret/tls.key'
    client_cert '/var/run/ocp-collector/secrets/my-es-secret/tls.crt'
    ca_file '/var/run/ocp-collector/secrets/my-es-secret/ca-bundle.crt'
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
      path '/var/lib/fluentd/retry_apps_es_1'
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
    @id apps_es_1
    host es.svc.messaging.cluster.local
    port 9654
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    client_key '/var/run/ocp-collector/secrets/my-es-secret/tls.key'
    client_cert '/var/run/ocp-collector/secrets/my-es-secret/tls.crt'
    ca_file '/var/run/ocp-collector/secrets/my-es-secret/ca-bundle.crt'
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    retry_tag retry_apps_es_1
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
      path '/var/lib/fluentd/apps_es_1'
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

<label @APPS_ES_2>
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
  
  <match retry_apps_es_2>
    @type elasticsearch
    @id retry_apps_es_2
    host es.svc.messaging.cluster.local2
    port 9654
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    client_key '/var/run/ocp-collector/secrets/my-other-secret/tls.key'
    client_cert '/var/run/ocp-collector/secrets/my-other-secret/tls.crt'
    ca_file '/var/run/ocp-collector/secrets/my-other-secret/ca-bundle.crt'
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
      path '/var/lib/fluentd/retry_apps_es_2'
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
    @id apps_es_2
    host es.svc.messaging.cluster.local2
    port 9654
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    client_key '/var/run/ocp-collector/secrets/my-other-secret/tls.key'
    client_cert '/var/run/ocp-collector/secrets/my-other-secret/tls.crt'
    ca_file '/var/run/ocp-collector/secrets/my-other-secret/ca-bundle.crt'
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    retry_tag retry_apps_es_2
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
      path '/var/lib/fluentd/apps_es_2'
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

<label @AUDIT_ES>
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
  
  <match retry_audit_es>
    @type elasticsearch
    @id retry_audit_es
    host es.svc.audit.cluster
    port 9654
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    client_key '/var/run/ocp-collector/secrets/my-audit-secret/tls.key'
    client_cert '/var/run/ocp-collector/secrets/my-audit-secret/tls.crt'
    ca_file '/var/run/ocp-collector/secrets/my-audit-secret/ca-bundle.crt'
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
      path '/var/lib/fluentd/retry_audit_es'
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
    @id audit_es
    host es.svc.audit.cluster
    port 9654
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    client_key '/var/run/ocp-collector/secrets/my-audit-secret/tls.key'
    client_cert '/var/run/ocp-collector/secrets/my-audit-secret/tls.crt'
    ca_file '/var/run/ocp-collector/secrets/my-audit-secret/ca-bundle.crt'
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    retry_tag retry_audit_es
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
      path '/var/lib/fluentd/audit_es'
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
`))
	})

	It("should generate sources for reserved inputs used as names or types", func() {
		sources := generator.GatherSources(&logging.ClusterLogForwarderSpec{
			Inputs: []logging.InputSpec{{Name: "in", Application: &logging.Application{}}},
			Pipelines: []logging.PipelineSpec{
				{
					InputRefs:  []string{"in"},
					OutputRefs: []string{"default"},
				},
				{
					InputRefs:  []string{"audit"},
					OutputRefs: []string{"default"},
				},
			},
		}, generator.NoOptions)
		Expect(sources.List()).To(ContainElements("application", "audit"))
	})

	Describe("Json Parsing", func() {

		fw := &logging.ClusterLogForwarderSpec{
			Outputs: []logging.OutputSpec{
				{
					Type: logging.OutputTypeElasticsearch,
					Name: "apps-es-1",
					URL:  "https://es.svc.messaging.cluster.local:9654",
				},
			},
			Inputs: []logging.InputSpec{
				{
					Name:        "myInput",
					Application: &logging.Application{},
				},
			},
			Pipelines: []logging.PipelineSpec{
				{
					Name:       "apps-pipeline",
					InputRefs:  []string{"myInput"},
					OutputRefs: []string{"apps-es-1"},
					Parse:      "json",
				},
			},
		}
		It("should generate config for Json parsing", func() {
			c := PipelineToOutputs(fw, nil)
			spec, err := g.GenerateConf(c...)
			Expect(err).To(BeNil())
			Expect(spec).To(EqualTrimLines(`
# Copying pipeline apps-pipeline to outputs
<label @APPS_PIPELINE>
  # Parse the logs into json
  <filter **>
    @type parser
    key_name message
    reserve_data yes
    hash_value_field structured
    emit_invalid_record_to_error false
    <parse>
      @type json
      json_parser oj
    </parse>
  </filter>
  
  <match **>
    @type relabel
    @label @APPS_ES_1
  </match>
</label>
`))
		})
	})

	var _ = DescribeTable("Verify generated fluentd.conf for valid forwarder specs",
		func(yamlSpec, wantFluentdConf string) {
			var spec logging.ClusterLogForwarderSpec
			Expect(yaml.Unmarshal([]byte(yamlSpec), &spec)).To(Succeed())
			g := generator.MakeGenerator()
			s := Conf(nil, security.NoSecrets, &spec, generator.NoOptions)
			gotFluentdConf, err := g.GenerateConf(generator.MergeSections(s)...)
			Expect(err).To(Succeed())
			Expect(gotFluentdConf).To(EqualTrimLines(wantFluentdConf))
		},
		Entry("namespaces", `
pipelines:
  - name: test-app
    inputrefs:
      - inputest
    outputrefs:
      - default
inputs:
  - name: inputest
    application:
      namespaces:
       - project1
       - project2
        `, `
## CLO GENERATED CONFIGURATION ###
# This file is a copy of the fluentd configuration entrypoint
# which should normally be supplied in a configmap.

<system>
  log_level "#{ENV['LOG_LEVEL'] || 'warn'}"
</system>

# Prometheus Monitoring
<source>
  @type prometheus
  bind "[::]"
  <transport tls>
    cert_path /etc/collector/metrics/tls.crt
    private_key_path /etc/collector/metrics/tls.key
  </transport>
</source>

<source>
  @type prometheus_monitor
  <labels>
    hostname ${hostname}
  </labels>
</source>

# excluding prometheus_tail_monitor
# since it leaks namespace/pod info
# via file paths

# tail_monitor plugin which publishes log_collected_bytes_total
<source>
  @type collected_tail_monitor
  <labels>
    hostname ${hostname}
  </labels>
</source>

# This is considered experimental by the repo
<source>
  @type prometheus_output_monitor
  <labels>
    hostname ${hostname}
  </labels>
</source>

# Logs from containers (including openshift containers)
<source>
  @type tail
  @id container-input
  path "/var/log/pods/*/*/*.log"
  exclude_path ["/var/log/pods/openshift-logging_collector-*/*/*.log", "/var/log/pods/openshift-logging_elasticsearch-*/*/*.log", "/var/log/pods/openshift-logging_kibana-*/*/*.log", "/var/log/pods/*/*/*.gz", "/var/log/pods/*/*/*.tmp"]
  pos_file "/var/lib/fluentd/pos/es-containers.log.pos"
  follow_inodes true
  refresh_interval 5
  rotate_wait 5
  tag kubernetes.*
  read_from_head "true"
  skip_refresh_on_startup true
  @label @CONCAT
  <parse>
    @type regexp
    expression /^(?<@timestamp>[^\s]+) (?<stream>stdout|stderr) (?<logtag>[F|P]) (?<message>.*)$/
    time_key '@timestamp'
    keep_time_key true
  </parse>
</source>

# Concat log lines of container logs, and send to INGRESS pipeline
<label @CONCAT>
  <filter kubernetes.**>
    @type concat
    key message
    partial_key logtag
    partial_value P
    separator ''
  </filter>
  
  <match kubernetes.**>
    @type relabel
    @label @INGRESS
  </match>
</label>

# Ingress pipeline
<label @INGRESS>
  # Filter out PRIORITY from journal logs
  <filter journal>
    @type grep
    <exclude>
      key PRIORITY
      pattern ^7$
    </exclude>
  </filter>
  
  # Process OVN logs
  <filter ovn-audit.log**>
    @type record_modifier
    <record>
      @timestamp ${DateTime.parse(record['message'].split('|')[0]).rfc3339(6)}
      level ${record['message'].split('|')[3].downcase}
    </record>
  </filter>
  
  # Retag Journal logs to specific tags
  <match journal>
    @type rewrite_tag_filter
    # skip to @INGRESS label section
    @label @INGRESS
  
    # see if this is a kibana container for special log handling
    # looks like this:
    # k8s_kibana.a67f366_logging-kibana-1-d90e3_logging_26c51a61-2835-11e6-ad29-fa163e4944d5_f0db49a2
    # we filter these logs through the kibana_transform.conf filter
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_kibana\.
      tag kubernetes.journal.container.kibana
    </rule>
  
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_logging-eventrouter-[^_]+_
      tag kubernetes.journal.container._default_.kubernetes-event
    </rule>
  
    # mark logs from default namespace for processing as k8s logs but stored as system logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_[^_]+_default_
      tag kubernetes.journal.container._default_
    </rule>
  
    # mark logs from kube-* namespaces for processing as k8s logs but stored as system logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_[^_]+_kube-(.+)_
      tag kubernetes.journal.container._kube-$1_
    </rule>
  
    # mark logs from openshift-* namespaces for processing as k8s logs but stored as system logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_[^_]+_openshift-(.+)_
      tag kubernetes.journal.container._openshift-$1_
    </rule>
  
    # mark logs from openshift namespace for processing as k8s logs but stored as system logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_[^_]+_[^_]+_openshift_
      tag kubernetes.journal.container._openshift_
    </rule>
  
    # mark fluentd container logs
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_.*fluentd
      tag kubernetes.journal.container.fluentd
    </rule>
  
    # this is a kubernetes container
    <rule>
      key CONTAINER_NAME
      pattern ^k8s_
      tag kubernetes.journal.container
    </rule>
  
    # not kubernetes - assume a system log or system container log
    <rule>
      key _TRANSPORT
      pattern .+
      tag journal.system
    </rule>
  </match>
  
  # Invoke kubernetes apiserver to get kubernetes metadata
  <filter kubernetes.**>
	@id kubernetes-metadata
    @type kubernetes_metadata
    kubernetes_url 'https://kubernetes.default.svc'
	annotation_match ["^containerType\.logging\.openshift\.io\/.*$"]
	allow_orphans false
    cache_size '1000'
    ssl_partial_chain 'true'
  </filter>
  
  # Parse Json fields for container, journal and eventrouter logs
  <filter kubernetes.var.log.pods.**_eventrouter-**>
    @type parse_json_field
    merge_json_log true
    preserve_json_log true
    json_fields 'message'
  </filter>
  
  # Fix level field in audit logs
  <filter k8s-audit.log**>
    @type record_modifier
    <record>
      k8s_audit_level ${record['level']}
    </record>
  </filter>
  
  <filter openshift-audit.log**>
    @type record_modifier
    <record>
      openshift_audit_level ${record['level']}
    </record>
  </filter>
  
  # Viaq Data Model
  <filter **>
    @type viaq_data_model
    enable_flatten_labels true
    enable_prune_empty_fields false
    default_keep_fields CEE,time,@timestamp,aushape,ci_job,collectd,docker,fedora-ci,file,foreman,geoip,hostname,ipaddr4,ipaddr6,kubernetes,level,message,namespace_name,namespace_uuid,offset,openstack,ovirt,pid,pipeline_metadata,rsyslog,service,systemd,tags,testcase,tlog,viaq_msg_id
    keep_empty_fields 'message'
    rename_time true
    pipeline_type 'collector'
    process_kubernetes_events false
    <level>
      name warn
      match 'Warning|WARN|^W[0-9]+|level=warn|Value:warn|"level":"warn"'
    </level>
    <level>
      name info
      match 'Info|INFO|^I[0-9]+|level=info|Value:info|"level":"info"'
    </level>
    <level>
      name error
      match 'Error|ERROR|^E[0-9]+|level=error|Value:error|"level":"error"'
    </level>
    <level>
      name critical
      match 'Critical|CRITICAL|^C[0-9]+|level=critical|Value:critical|"level":"critical"'
    </level>
    <level>
      name debug
      match 'Debug|DEBUG|^D[0-9]+|level=debug|Value:debug|"level":"debug"'
    </level>
    <formatter>
      tag "journal.system**"
      type sys_journal
      remove_keys log,stream,MESSAGE,_SOURCE_REALTIME_TIMESTAMP,__REALTIME_TIMESTAMP,CONTAINER_ID,CONTAINER_ID_FULL,CONTAINER_NAME,PRIORITY,_BOOT_ID,_CAP_EFFECTIVE,_CMDLINE,_COMM,_EXE,_GID,_HOSTNAME,_MACHINE_ID,_PID,_SELINUX_CONTEXT,_SYSTEMD_CGROUP,_SYSTEMD_SLICE,_SYSTEMD_UNIT,_TRANSPORT,_UID,_AUDIT_LOGINUID,_AUDIT_SESSION,_SYSTEMD_OWNER_UID,_SYSTEMD_SESSION,_SYSTEMD_USER_UNIT,CODE_FILE,CODE_FUNCTION,CODE_LINE,ERRNO,MESSAGE_ID,RESULT,UNIT,_KERNEL_DEVICE,_KERNEL_SUBSYSTEM,_UDEV_SYSNAME,_UDEV_DEVNODE,_UDEV_DEVLINK,SYSLOG_FACILITY,SYSLOG_IDENTIFIER,SYSLOG_PID
    </formatter>
    <formatter>
      tag "kubernetes.var.log.pods.**_eventrouter-** k8s-audit.log** openshift-audit.log** ovn-audit.log**"
      type k8s_json_file
      remove_keys stream
      process_kubernetes_events 'true'
    </formatter>
    <formatter>
      tag "kubernetes.var.log.pods**"
      type k8s_json_file
      remove_keys stream
    </formatter>
  </filter>
  
  # Generate elasticsearch id
  <filter **>
    @type elasticsearch_genid_ext
    hash_id_key viaq_msg_id
    alt_key kubernetes.event.metadata.uid
    alt_tags 'kubernetes.var.log.pods.**_eventrouter-*.** kubernetes.journal.container._default_.kubernetes-event'
  </filter>
  
  # Discard Infrastructure logs
  <match kubernetes.var.log.pods.openshift_** kubernetes.var.log.pods.openshift-*_** kubernetes.var.log.pods.default_** kubernetes.var.log.pods.kube-*_** journal.** system.var.log**>
    @type null
  </match>
  
  # Include Application logs
  <match kubernetes.**>
    @type relabel
    @label @_APPLICATION
  </match>
  
  # Discard Audit logs
  <match linux-audit.log** k8s-audit.log** openshift-audit.log** ovn-audit.log**>
    @type null
  </match>
  
  # Send any remaining unmatched tags to stdout
  <match **>
   @type stdout
  </match>
</label>

# Routing Application to pipelines
<label @_APPLICATION>
  <filter **>
    @type record_modifier
    <record>
      log_type application
    </record>
  </filter>
  
  <match **>
    @type label_router
    <route>
      @label @TEST_APP
      <match>
        namespaces project1, project2
      </match>
    </route>
  </match>
</label>

# Copying pipeline test-app to outputs
<label @TEST_APP>
  <match **>
    @type relabel
    @label @DEFAULT
  </match>
</label>

# Ship logs to specific outputs
`))

})
