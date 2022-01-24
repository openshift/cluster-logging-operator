package fluentd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/generator"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	corev1 "k8s.io/api/core/v1"
)

//TODO: Use a detailed CLF spec
var _ = Describe("Testing Complete Config Generation", func() {
	var f = func(testcase generator.ConfGenerateTest) {
		g := generator.MakeGenerator()
		e := generator.MergeSections(Conf(&testcase.CLSpec, testcase.Secrets, &testcase.CLFSpec, generator.NoOptions))
		conf, err := g.GenerateConf(e...)
		Expect(err).To(BeNil())
		diff := cmp.Diff(
			strings.Split(strings.TrimSpace(testcase.ExpectedConf), "\n"),
			strings.Split(strings.TrimSpace(conf), "\n"))
		if diff != "" {
			b, _ := json.MarshalIndent(e, "", " ")
			fmt.Printf("elements:\n%s\n", string(b))
			fmt.Println(conf)
			fmt.Printf("diff: %s", diff)
		}
		Expect(diff).To(Equal(""))
	}
	DescribeTable("Generate full fluent.conf", f,
		Entry("with complex spec", generator.ConfGenerateTest{
			CLSpec: logging.ClusterLoggingSpec{
				Forwarder: &logging.ForwarderSpec{
					Fluentd: &logging.FluentdForwarderSpec{
						Buffer: &logging.FluentdBufferSpec{
							ChunkLimitSize: "8m",
							TotalLimitSize: "800000000",
							OverflowAction: "throw_exception",
						},
					},
				},
			},
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameApplication,
							logging.InputNameInfrastructure,
							logging.InputNameAudit},
						OutputRefs:            []string{"es-1"},
						Name:                  "pipeline",
						DetectMultilineErrors: true,
					},
				},
				Outputs: []logging.OutputSpec{
					{
						Name: "es-1",
						Type: logging.OutputTypeElasticsearch,
						URL:  "https://es.svc.infra.cluster:9999",
						Secret: &logging.OutputSecretSpec{
							Name: "es-1-secret",
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
## CLO GENERATED CONFIGURATION ###
# This file is a copy of the fluentd configuration entrypoint
# which should normally be supplied in a configmap.

<system>
  log_level "#{ENV['LOG_LEVEL'] || 'warn'}"
</system>

# Prometheus Monitoring
<source>
  @type prometheus
  bind "#{ENV['POD_IP']}"
  <ssl>
    enable true
    certificate_path "#{ENV['METRICS_CERT'] || '/etc/fluent/metrics/tls.crt'}"
    private_key_path "#{ENV['METRICS_KEY'] || '/etc/fluent/metrics/tls.key'}"
  </ssl>
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
  @label @MEASURE
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
  path "/var/log/pods/**/*.log"
  exclude_path ["/var/log/pods/openshift-logging_collector-*/*/*.log", "/var/log/pods/openshift-logging_elasticsearch-*/*/*.log", "/var/log/pods/openshift-logging_kibana-*/*/*.log"]
  pos_file "/var/lib/fluentd/pos/es-containers.log.pos"
  refresh_interval 5
  rotate_wait 5
  tag kubernetes.*
  read_from_head "true"
  skip_refresh_on_startup true
  @label @MEASURE
  <parse>
    @type multi_format
    <pattern>
      format json
      time_format '%Y-%m-%dT%H:%M:%S.%N%Z'
      keep_time_key true
    </pattern>
    <pattern>
      format regexp
      expression /^(?<time>[^\s]+) (?<stream>stdout|stderr)( (?<logtag>.))? (?<log>.*)$/
      time_format '%Y-%m-%dT%H:%M:%S.%N%:z'
      keep_time_key true
    </pattern>
  </parse>
</source>

# linux audit logs
<source>
  @type tail
  @id audit-input
  @label @MEASURE
  path "/var/log/audit/audit.log"
  pos_file "/var/lib/fluentd/pos/audit.log.pos"
  tag linux-audit.log
  <parse>
    @type viaq_host_audit
  </parse>
</source>

# k8s audit logs
<source>
  @type tail
  @id k8s-audit-input
  @label @MEASURE
  path "/var/log/kube-apiserver/audit.log"
  pos_file "/var/lib/fluentd/pos/kube-apiserver.audit.log.pos"
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
  @label @MEASURE
  path /var/log/oauth-apiserver/audit.log,/var/log/openshift-apiserver/audit.log
  pos_file /var/lib/fluentd/pos/oauth-apiserver.audit.log
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
  @label @MEASURE
  path "/var/log/ovn/acl-audit-log.log"
  pos_file "/var/lib/fluentd/pos/acl-audit-log.pos"
  tag ovn-audit.log
  refresh_interval 5
  rotate_wait 5
  read_from_head true
  <parse>
    @type none
  </parse>
</source>

# Increment Prometheus metrics
<label @MEASURE>
  <filter **>
    @type record_transformer
    enable_ruby
    <record>
      msg_size ${record.to_s.length}
    </record>
  </filter>
  
  <filter **>
    @type prometheus
    <metric>
      name cluster_logging_collector_input_record_total
      type counter
      desc The total number of incoming records
      <labels>
        tag ${tag}
        hostname ${hostname}
      </labels>
    </metric>
  </filter>
  
  <filter **>
    @type prometheus
    <metric>
      name cluster_logging_collector_input_record_bytes
      type counter
      desc The total bytes of incoming records
      key msg_size
      <labels>
        tag ${tag}
        hostname ${hostname}
      </labels>
    </metric>
  </filter>
  
  <filter **>
    @type record_transformer
    remove_keys msg_size
  </filter>
  
  # Journal Logs go to INGRESS pipeline
  <match journal>
    @type relabel
    @label @INGRESS
  </match>
  
  # Audit Logs go to INGRESS pipeline
  <match *audit.log>
    @type relabel
    @label @INGRESS
  </match>
  
  # Kubernetes Logs go to CONCAT pipeline
  <match kubernetes.**>
    @type relabel
    @label @CONCAT
  </match>
</label>

# Concat log lines of container logs, and send to INGRESS pipeline
<label @CONCAT>
  <filter kubernetes.**>
    @type concat
    key log
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
  
  # Invoke kubernetes apiserver to get kunbernetes metadata
  <filter kubernetes.**>
    @id kubernetes-metadata
    @type kubernetes_metadata
    kubernetes_url 'https://kubernetes.default.svc'
    cache_size '1000'
    watch 'false'
    use_journal 'nil'
    ssl_partial_chain 'true'
  </filter>
  
  # Parse Json fields for container, journal and eventrouter logs
  <filter kubernetes.journal.**>
    @type parse_json_field
    merge_json_log 'false'
    preserve_json_log 'true'
    json_fields 'log,MESSAGE'
  </filter>
  
  <filter kubernetes.var.log.pods.**>
    @type parse_json_field
    merge_json_log 'false'
    preserve_json_log 'true'
    json_fields 'log,MESSAGE'
  </filter>
  
  <filter kubernetes.var.log.pods.**_eventrouter-**>
    @type parse_json_field
    merge_json_log true
    preserve_json_log true
    json_fields 'log,MESSAGE'
  </filter>
  
  # Clean kibana log fields
  <filter **kibana**>
    @type record_transformer
    enable_ruby
    <record>
      log ${record['err'] || record['msg'] || record['MESSAGE'] || record['log']}
    </record>
    remove_keys req,res,msg,name,level,v,pid,err
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
    elasticsearch_index_prefix_field 'viaq_index_name'
    default_keep_fields CEE,time,@timestamp,aushape,ci_job,collectd,docker,fedora-ci,file,foreman,geoip,hostname,ipaddr4,ipaddr6,kubernetes,level,message,namespace_name,namespace_uuid,offset,openstack,ovirt,pid,pipeline_metadata,rsyslog,service,systemd,tags,testcase,tlog,viaq_msg_id
    extra_keep_fields ''
    keep_empty_fields 'message'
    use_undefined false
    undefined_name 'undefined'
    rename_time true
    rename_time_if_missing false
    src_time_name 'time'
    dest_time_name '@timestamp'
    pipeline_type 'collector'
    undefined_to_string 'false'
    undefined_dot_replace_char 'UNUSED'
    undefined_max_num_fields '-1'
    process_kubernetes_events 'false'
    <level>
      name warn
      match 'Warning|WARN|W[0-9]+|level=warn|Value:warn|"level":"warn"'
    </level>
    <level>
      name info
      match 'Info|INFO|I[0-9]+|level=info|Value:info|"level":"info"'
    </level>
    <level>
      name error
      match 'Error|ERROR|E[0-9]+|level=error|Value:error|"level":"error"'
    </level>
    <level>
      name debug
      match 'Debug|DEBUG|D[0-9]+|level=debug|Value:debug|"level":"debug"'
    </level>
    <formatter>
      tag "system.var.log**"
      type sys_var_log
      remove_keys host,pid,ident
    </formatter>
    <formatter>
      tag "journal.system**"
      type sys_journal
      remove_keys log,stream,MESSAGE,_SOURCE_REALTIME_TIMESTAMP,__REALTIME_TIMESTAMP,CONTAINER_ID,CONTAINER_ID_FULL,CONTAINER_NAME,PRIORITY,_BOOT_ID,_CAP_EFFECTIVE,_CMDLINE,_COMM,_EXE,_GID,_HOSTNAME,_MACHINE_ID,_PID,_SELINUX_CONTEXT,_SYSTEMD_CGROUP,_SYSTEMD_SLICE,_SYSTEMD_UNIT,_TRANSPORT,_UID,_AUDIT_LOGINUID,_AUDIT_SESSION,_SYSTEMD_OWNER_UID,_SYSTEMD_SESSION,_SYSTEMD_USER_UNIT,CODE_FILE,CODE_FUNCTION,CODE_LINE,ERRNO,MESSAGE_ID,RESULT,UNIT,_KERNEL_DEVICE,_KERNEL_SUBSYSTEM,_UDEV_SYSNAME,_UDEV_DEVNODE,_UDEV_DEVLINK,SYSLOG_FACILITY,SYSLOG_IDENTIFIER,SYSLOG_PID
    </formatter>
    <formatter>
      tag "kubernetes.journal.container**"
      type k8s_journal
      remove_keys 'log,stream,MESSAGE,_SOURCE_REALTIME_TIMESTAMP,__REALTIME_TIMESTAMP,CONTAINER_ID,CONTAINER_ID_FULL,CONTAINER_NAME,PRIORITY,_BOOT_ID,_CAP_EFFECTIVE,_CMDLINE,_COMM,_EXE,_GID,_HOSTNAME,_MACHINE_ID,_PID,_SELINUX_CONTEXT,_SYSTEMD_CGROUP,_SYSTEMD_SLICE,_SYSTEMD_UNIT,_TRANSPORT,_UID,_AUDIT_LOGINUID,_AUDIT_SESSION,_SYSTEMD_OWNER_UID,_SYSTEMD_SESSION,_SYSTEMD_USER_UNIT,CODE_FILE,CODE_FUNCTION,CODE_LINE,ERRNO,MESSAGE_ID,RESULT,UNIT,_KERNEL_DEVICE,_KERNEL_SUBSYSTEM,_UDEV_SYSNAME,_UDEV_DEVNODE,_UDEV_DEVLINK,SYSLOG_FACILITY,SYSLOG_IDENTIFIER,SYSLOG_PID'
    </formatter>
    <formatter>
      tag "kubernetes.var.log.pods.**_eventrouter-** k8s-audit.log** openshift-audit.log** ovn-audit.log**"
      type k8s_json_file
      remove_keys log,stream,CONTAINER_ID_FULL,CONTAINER_NAME
      process_kubernetes_events 'true'
    </formatter>
    <formatter>
      tag "kubernetes.var.log.pods**"
      type k8s_json_file
      remove_keys log,stream,CONTAINER_ID_FULL,CONTAINER_NAME
    </formatter>
    <elasticsearch_index_name>
      enabled 'true'
      tag "journal.system** system.var.log** **_default_** **_kube-*_** **_openshift-*_** **_openshift_**"
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
      name_type static
      static_index_name app-write
    </elasticsearch_index_name>
  </filter>
  
  # Generate elasticsearch id
  <filter **>
    @type elasticsearch_genid_ext
    hash_id_key viaq_msg_id
    alt_key kubernetes.event.metadata.uid
    alt_tags 'kubernetes.var.log.pods.**_eventrouter-*.** kubernetes.journal.container._default_.kubernetes-event'
  </filter>
  
  # Include Infrastructure logs
  <match **_default_** **_kube-*_** **_openshift-*_** **_openshift_** journal.** system.var.log**>
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
    @label @PIPELINE
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
    @label @PIPELINE
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
    @label @PIPELINE
  </match>
</label>

# Copying pipeline pipeline to outputs
<label @PIPELINE>
  <match kubernetes.**>
    @id multiline-detect-except
    @type detect_exceptions
    remove_tag_prefix 'kubernetes'
    message message
    force_line_breaks true
    multiline_flush_interval .2
  </match>
  <match **>
    @type relabel
    @label @ES_1
  </match>
</label>

# Ship logs to specific outputs
<label @ES_1>
  #remove structured field if present
  <filter **>
    @type record_modifier
    remove_keys structured
  </filter>
  
  #flatten labels to prevent field explosion in ES
  <filter **>
    @type record_transformer
    enable_ruby true
    <record>
      kubernetes ${!record['kubernetes'].nil? ? record['kubernetes'].merge({"flat_labels": (record['kubernetes']['labels']||{}).map{|k,v| "#{k}=#{v}"}}) : {} }
    </record>
    remove_keys $.kubernetes.labels
  </filter>
  
  <match retry_es_1>
    @type elasticsearch
    @id retry_es_1
    host es.svc.infra.cluster
    port 9999
    verify_es_version_at_startup false
    scheme https
    ssl_version TLSv1_2
    client_key '/var/run/ocp-collector/secrets/es-1-secret/tls.key'
    client_cert '/var/run/ocp-collector/secrets/es-1-secret/tls.crt'
    ca_file '/var/run/ocp-collector/secrets/es-1-secret/ca-bundle.crt'
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    http_backend typhoeus
    write_operation create
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
      total_limit_size 800000000
      chunk_limit_size 8m
      overflow_action throw_exception
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
    client_key '/var/run/ocp-collector/secrets/es-1-secret/tls.key'
    client_cert '/var/run/ocp-collector/secrets/es-1-secret/tls.crt'
    ca_file '/var/run/ocp-collector/secrets/es-1-secret/ca-bundle.crt'
    target_index_key viaq_index_name
    id_key viaq_msg_id
    remove_keys viaq_index_name
    type_name _doc
    retry_tag retry_es_1
    http_backend typhoeus
    write_operation create
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
      total_limit_size 800000000
      chunk_limit_size 8m
      overflow_action throw_exception
    </buffer>
  </match>
</label>
`,
		}),
	)
})
