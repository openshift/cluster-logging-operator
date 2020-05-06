package fluentd

// fluentConf: source -> fan to pipelines -> pipeline -> output [store]
var templateRegistry = []string{
	inputSourceContainerTemplate,
	inputSourceJournalTemplate,
	inputSourceHostAuditTemplate,
	inputSourceK8sAuditTemplate,
	inputSourceOpenShiftAuditTemplate,
	fluentConfTemplate,
	pipelineToOutputCopyTemplate,
	sourceToPipelineCopyTemplate,
	outputLabelConfTemplate,
	outputLabelConfNocopyTemplate,
	outputLabelConfNoretryTemplate,
	storeElasticsearchTemplate,
	forwardTemplate,
	storeSyslogTemplateOld,
	storeSyslogTemplate,
}

const fluentConfTemplate = `{{- define "fluentConf" -}}
## CLO GENERATED CONFIGURATION ###
# This file is a copy of the fluentd configuration entrypoint
# which should normally be supplied in a configmap.

<system>
  @log_level "#{ENV['LOG_LEVEL'] || 'warn'}"
</system>

# In each section below, pre- and post- includes don't include anything initially;
# they exist to enable future additions to openshift conf as needed.

## sources
## ordered so that syslog always runs last...
<source>
  @type prometheus
  bind ''
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

# This is considered experimental by the repo
<source>
  @type prometheus_output_monitor
  <labels>
    hostname ${hostname}
  </labels>
</source>

{{- range .SourceInputLabels }}
{{ . }}
{{- end}}

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

#syslog input config here

<label @INGRESS>

  ## filters
  <filter **>
    @type record_modifier
    char_encoding utf-8
  </filter>

  <filter journal>
    @type grep
    <exclude>
      key PRIORITY
      pattern ^7$
    </exclude>
  </filter>

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

  <filter kubernetes.**>
    @type kubernetes_metadata
    kubernetes_url "#{ENV['K8S_HOST_URL']}"
    cache_size "#{ENV['K8S_METADATA_CACHE_SIZE'] || '1000'}"
    watch "#{ENV['K8S_METADATA_WATCH'] || 'false'}"
    use_journal "#{ENV['USE_JOURNAL'] || 'nil'}"
    ssl_partial_chain "#{ENV['SSL_PARTIAL_CHAIN'] || 'true'}"
  </filter>

  <filter kubernetes.journal.**>
    @type parse_json_field
    merge_json_log "#{ENV['MERGE_JSON_LOG'] || 'false'}"
    preserve_json_log "#{ENV['PRESERVE_JSON_LOG'] || 'true'}"
    json_fields "#{ENV['JSON_FIELDS'] || 'MESSAGE,log'}"
  </filter>

  <filter kubernetes.var.log.containers.**>
    @type parse_json_field
    merge_json_log "#{ENV['MERGE_JSON_LOG'] || 'false'}"
    preserve_json_log "#{ENV['PRESERVE_JSON_LOG'] || 'true'}"
    json_fields "#{ENV['JSON_FIELDS'] || 'log,MESSAGE'}"
  </filter>

  <filter kubernetes.var.log.containers.eventrouter-** kubernetes.var.log.containers.cluster-logging-eventrouter-**>
    @type parse_json_field
    merge_json_log true
    preserve_json_log true
    json_fields "#{ENV['JSON_FIELDS'] || 'log,MESSAGE'}"
  </filter>

  <filter **kibana**>
    @type record_transformer
    enable_ruby
    <record>
      log ${record['err'] || record['msg'] || record['MESSAGE'] || record['log']}
    </record>
    remove_keys req,res,msg,name,level,v,pid,err
  </filter>

  <filter **>
    @type viaq_data_model
    elasticsearch_index_prefix_field 'viaq_index_name'
    default_keep_fields CEE,time,@timestamp,aushape,ci_job,collectd,docker,fedora-ci,file,foreman,geoip,hostname,ipaddr4,ipaddr6,kubernetes,level,message,namespace_name,namespace_uuid,offset,openstack,ovirt,pid,pipeline_metadata,rsyslog,service,systemd,tags,testcase,tlog,viaq_msg_id
    extra_keep_fields "#{ENV['CDM_EXTRA_KEEP_FIELDS'] || ''}"
    keep_empty_fields "#{ENV['CDM_KEEP_EMPTY_FIELDS'] || 'message'}"
    use_undefined "#{ENV['CDM_USE_UNDEFINED'] || false}"
    undefined_name "#{ENV['CDM_UNDEFINED_NAME'] || 'undefined'}"
    rename_time "#{ENV['CDM_RENAME_TIME'] || true}"
    rename_time_if_missing "#{ENV['CDM_RENAME_TIME_IF_MISSING'] || false}"
    src_time_name "#{ENV['CDM_SRC_TIME_NAME'] || 'time'}"
    dest_time_name "#{ENV['CDM_DEST_TIME_NAME'] || '@timestamp'}"
    pipeline_type "#{ENV['PIPELINE_TYPE'] || 'collector'}"
    undefined_to_string "#{ENV['CDM_UNDEFINED_TO_STRING'] || 'false'}"
    undefined_dot_replace_char "#{ENV['CDM_UNDEFINED_DOT_REPLACE_CHAR'] || 'UNUSED'}"
    undefined_max_num_fields "#{ENV['CDM_UNDEFINED_MAX_NUM_FIELDS'] || '-1'}"
    process_kubernetes_events "#{ENV['TRANSFORM_EVENTS'] || 'false'}"
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
      remove_keys "#{ENV['K8S_FILTER_REMOVE_KEYS'] || 'log,stream,MESSAGE,_SOURCE_REALTIME_TIMESTAMP,__REALTIME_TIMESTAMP,CONTAINER_ID,CONTAINER_ID_FULL,CONTAINER_NAME,PRIORITY,_BOOT_ID,_CAP_EFFECTIVE,_CMDLINE,_COMM,_EXE,_GID,_HOSTNAME,_MACHINE_ID,_PID,_SELINUX_CONTEXT,_SYSTEMD_CGROUP,_SYSTEMD_SLICE,_SYSTEMD_UNIT,_TRANSPORT,_UID,_AUDIT_LOGINUID,_AUDIT_SESSION,_SYSTEMD_OWNER_UID,_SYSTEMD_SESSION,_SYSTEMD_USER_UNIT,CODE_FILE,CODE_FUNCTION,CODE_LINE,ERRNO,MESSAGE_ID,RESULT,UNIT,_KERNEL_DEVICE,_KERNEL_SUBSYSTEM,_UDEV_SYSNAME,_UDEV_DEVNODE,_UDEV_DEVLINK,SYSLOG_FACILITY,SYSLOG_IDENTIFIER,SYSLOG_PID'}"
    </formatter>
    <formatter>
      tag "kubernetes.var.log.containers.eventrouter-** kubernetes.var.log.containers.cluster-logging-eventrouter-** k8s-audit.log** openshift-audit.log**"
      type k8s_json_file
      remove_keys log,stream,CONTAINER_ID_FULL,CONTAINER_NAME
      process_kubernetes_events "#{ENV['TRANSFORM_EVENTS'] || 'true'}"
    </formatter>
    <formatter>
      tag "kubernetes.var.log.containers**"
      type k8s_json_file
      remove_keys log,stream,CONTAINER_ID_FULL,CONTAINER_NAME
    </formatter>
    <elasticsearch_index_name>
      enabled "#{ENV['ENABLE_ES_INDEX_NAME'] || 'true'}"
      tag "journal.system** system.var.log** **_default_** **_kube-*_** **_openshift-*_** **_openshift_**"
      name_type static
      static_index_name infra-write
    </elasticsearch_index_name>
    <elasticsearch_index_name>
      enabled "#{ENV['ENABLE_ES_INDEX_NAME'] || 'true'}"
      tag "linux-audit.log** k8s-audit.log** openshift-audit.log**"
      name_type static
      static_index_name audit-write
    </elasticsearch_index_name>
    <elasticsearch_index_name>
      enabled "#{ENV['ENABLE_ES_INDEX_NAME'] || 'true'}"
      tag "**"
      name_type static
      static_index_name app-write
    </elasticsearch_index_name>
  </filter>

  <filter **>
    @type elasticsearch_genid_ext
    hash_id_key viaq_msg_id
    alt_key kubernetes.event.metadata.uid
    alt_tags "#{ENV['GENID_ALT_TAG'] || 'kubernetes.var.log.containers.logging-eventrouter-*.** kubernetes.var.log.containers.eventrouter-*.** kubernetes.var.log.containers.cluster-logging-eventrouter-*.** kubernetes.journal.container._default_.kubernetes-event'}"
  </filter>

  #flatten labels to prevent field explosion in ES
  <filter ** >
    @type record_transformer
    enable_ruby true
    <record>
      kubernetes ${!record['kubernetes'].nil? ? record['kubernetes'].merge({"flat_labels": (record['kubernetes']['labels']||{}).map{|k,v| "#{k}=#{v}"}}) : {} }
    </record>
    remove_keys $.kubernetes.labels
  </filter>

  # Relabel specific source tags to specific intermediary labels for copy processing
{{ if .CollectInfraLogs }}
  <match **_default_** **_kube-*_** **_openshift-*_** **_openshift_** journal.** system.var.log**>
    @type relabel
    @label @_INFRASTRUCTURE
  </match>
{{- end}}
{{ if .CollectAppLogs}}
  <match kubernetes.**>
    @type relabel
    @label @_APPLICATION
  </match>
{{- end}}
{{ if .CollectAuditLogs}}
  <match linux-audit.log** k8s-audit.log** openshift-audit.log**>
    @type relabel
    @label @_AUDIT
  </match>
{{- end}}
  <match **>
    @type stdout
  </match>

</label>

# Relabel specific sources (e.g. logs-apps) to multiple pipelines
{{- range .SourceToPipelineLabels }}
{{ . }}
{{- end}}

# Relabel specific pipelines to multiple, outputs (e.g. ES, kafka stores)
{{- range .PipelinesToOutputLabels }}
{{ . }}
{{- end}}

# Ship logs to specific outputs
{{- range .OutputLabels }}
{{ . }}
{{- end}}
{{ if .IncludeLegacySecureForward }}
<label @_LEGACY_SECUREFORWARD>
  <match **>
    @type copy
    #include legacy secure-forward.conf
    @include /etc/fluent/configs.d/secure-forward/secure-forward.conf
  </match>
</label>
{{- end}}
{{ if .IncludeLegacySyslog }}
<label @_LEGACY_SYSLOG>
  <match **>
    @type copy
    #include legacy Syslog
    @include /etc/fluent/configs.d/syslog/syslog.conf
  </match>
</label>
{{- end}}

{{- end}}`

const inputSourceJournalTemplate = `{{- define "inputSourceJournalTemplate" -}}
#journal logs to gather node
<source>
  @type systemd
  @id systemd-input
  @label @INGRESS
  path "#{if (val = ENV.fetch('JOURNAL_SOURCE','')) && (val.length > 0); val; else '/run/log/journal'; end}"
  <storage>
    @type local
    persistent true
    # NOTE: if this does not end in .json, fluentd will think it
    # is the name of a directory - see fluentd storage_local.rb
    path "#{ENV['JOURNAL_POS_FILE'] || '/var/log/journal_pos.json'}"
  </storage>
  matches "#{ENV['JOURNAL_FILTERS_JSON'] || '[]'}"
  tag journal
  read_from_head "#{if (val = ENV.fetch('JOURNAL_READ_FROM_HEAD','')) && (val.length > 0); val; else 'false'; end}"
</source>
{{- end}}`

const inputSourceContainerTemplate = `{{- define "inputSourceContainerTemplate" -}}
# container logs
<source>
  @type tail
  @id container-input
  {{- if .AppNsPaths}}
  path {{.AppNsPaths}}
  {{else}}
  path "/var/log/containers/*.log"
  {{end -}}
  exclude_path ["/var/log/containers/{{.CollectorPodNamePrefix}}-*_{{.LoggingNamespace}}_*.log", "/var/log/containers/{{.LogStorePodNamePrefix}}-*_{{.LoggingNamespace}}_*.log", "/var/log/containers/{{.VisualizationPodNamePrefix}}-*_{{.LoggingNamespace}}_*.log"]
  pos_file "/var/log/es-containers.log.pos"
  refresh_interval 5
  rotate_wait 5
  tag kubernetes.*
  read_from_head "true"
  @label @CONCAT
  <parse>
    @type multi_format
    <pattern>
      format json
      time_format '%Y-%m-%dT%H:%M:%S.%N%Z'
      keep_time_key true
    </pattern>
    <pattern>
      format regexp
      expression /^(?<time>.+) (?<stream>stdout|stderr)( (?<logtag>.))? (?<log>.*)$/
      time_format '%Y-%m-%dT%H:%M:%S.%N%:z'
      keep_time_key true
    </pattern>
  </parse>
</source>
{{- end}}`

const inputSourceHostAuditTemplate = `{{- define "inputSourceHostAuditTemplate" -}}
# linux audit logs
<source>
  @type tail
  @id audit-input
  @label @INGRESS
  path "#{ENV['AUDIT_FILE'] || '/var/log/audit/audit.log'}"
  pos_file "#{ENV['AUDIT_POS_FILE'] || '/var/log/audit/audit.log.pos'}"
  tag linux-audit.log
  <parse>
    @type viaq_host_audit
  </parse>
</source>
{{- end}}`

const inputSourceK8sAuditTemplate = `{{- define "inputSourceK8sAuditTemplate" -}}
# k8s audit logs
<source>
  @type tail
  @id k8s-audit-input
  @label @INGRESS
  path "#{ENV['K8S_AUDIT_FILE'] || '/var/log/kube-apiserver/audit.log'}"
  pos_file "#{ENV['K8S_AUDIT_POS_FILE'] || '/var/log/kube-apiserver/audit.log.pos'}"
  tag k8s-audit.log
  <parse>
    @type json
    time_key requestReceivedTimestamp
    # In case folks want to parse based on the requestReceivedTimestamp key
    keep_time_key true
    time_format %Y-%m-%dT%H:%M:%S.%N%z
  </parse>
</source>
{{- end}}`

const inputSourceOpenShiftAuditTemplate = `{{- define "inputSourceOpenShiftAuditTemplate" }}
# Openshift audit logs
<source>
  @type tail
  @id openshift-audit-input
  @label @INGRESS
  path "#{ENV['OPENSHIFT_AUDIT_FILE'] || '/var/log/openshift-apiserver/audit.log'}"
  pos_file "#{ENV['OPENSHIFT_AUDIT_FILE'] || '/var/log/openshift-apiserver/audit.log.pos'}"
  tag openshift-audit.log
  <parse>
    @type json
    time_key requestReceivedTimestamp
    # In case folks want to parse based on the requestReceivedTimestamp key
    keep_time_key true
    time_format %Y-%m-%dT%H:%M:%S.%N%z
  </parse>
</source>
{{- end}}`

const sourceToPipelineCopyTemplate = `{{- define "sourceToPipelineCopyTemplate" -}}
<label {{sourceTypelabelName .Source}}>
  <match **>
    @type copy
{{ range $index, $pipelineLabel := .PipelineNames }}
    <store>
      @type relabel
      @label {{labelName $pipelineLabel}}
    </store>
{{- end }}
{{ if .IncludeLegacySecureForward }}
    <store>
      @type relabel
      @label @_LEGACY_SECUREFORWARD
    </store>
{{- end }}
{{ if .IncludeLegacySyslog }}
    <store>
      @type relabel
      @label @_LEGACY_SYSLOG
    </store>
{{- end }}
  </match>
</label>
{{- end}}`

const pipelineToOutputCopyTemplate = `{{- define "pipelineToOutputCopyTemplate" -}}
<label {{labelName .Name}}>
  <match **>
    @type copy
{{ range $index, $target := .Outputs }}
    <store>
      @type relabel
      @label {{labelName $target}}
    </store>
{{- end }}
  </match>
</label>
{{- end}}`

const outputLabelConfTemplate = `{{- define "outputLabelConf" -}}
<label {{.LabelName}}>
  <match {{.RetryTag}}>
    @type copy
{{ include .StoreTemplate . "prefix_as_retry" | indent 4}}
  </match>
  <match **>
    @type copy
{{ include .StoreTemplate . "include_retry_tag"| indent 4}}
  </match>
</label>
{{- end}}`

const outputLabelConfNocopyTemplate = `{{- define "outputLabelConfNoCopy" -}}
<label {{.LabelName}}>
  <match **>
{{include .StoreTemplate . "" | indent 4}}
  </match>
</label>
{{- end}}`

const outputLabelConfNoretryTemplate = `{{- define "outputLabelConfNoRetry" -}}
<label {{.LabelName}}>
  <match **>
    @type copy
{{include .StoreTemplate . "" | indent 4}}
  </match>
</label>
{{- end}}`

const forwardTemplate = `{{- define "forward" -}}
# https://docs.fluentd.org/v1.0/articles/in_forward
@type forward
{{ if .Target.Secret }}
<security>
  self_hostname "#{ENV['NODE_NAME']}"
  shared_key "#{File.open('{{ .SecretPath "shared_key" }}') do |f| f.readline end.rstrip}"
</security>

transport tls
tls_verify_hostname false
tls_version 'TLSv1_2'

#tls_client_private_key_path {{ .SecretPath "tls.key"}}
tls_client_cert_path {{ .SecretPath "tls.crt"}}
tls_cert_path {{ .SecretPath "ca-bundle.crt"}}
{{ end -}}

<buffer>
  @type file
  path '{{.BufferPath}}'
  queue_limit_length "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
  chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '1m' }"
  flush_interval "#{ENV['FORWARD_FLUSH_INTERVAL'] || '5s'}"
  flush_at_shutdown "#{ENV['FLUSH_AT_SHUTDOWN'] || 'false'}"
  flush_thread_count "#{ENV['FLUSH_THREAD_COUNT'] || 2}"
  retry_max_interval "#{ENV['FORWARD_RETRY_WAIT'] || '300'}"
  retry_forever true
  # the systemd journald 0.0.8 input plugin will just throw away records if the buffer
  # queue limit is hit - 'block' will halt further reads and keep retrying to flush the
  # buffer to the remote - default is 'exception' because in_tail handles that case
  overflow_action "#{ENV['BUFFER_QUEUE_FULL_ACTION'] || 'exception'}"
</buffer>

<server>
  host {{.Host}}
  port {{.Port}}
</server>
{{- end}}`

const storeElasticsearchTemplate = `{{ define "storeElasticsearch" -}}
<store>
  @type elasticsearch
  @id {{.StoreID }}
  host {{.Host}}
  port {{.Port}}
  verify_es_version_at_startup false
{{- if .Target.Secret }}
  scheme https
  ssl_version TLSv1_2
{{- else }}
  scheme http
{{- end }}
  target_index_key viaq_index_name
  id_key viaq_msg_id
  remove_keys viaq_index_name
  user fluentd
  password changeme
{{- if .Target.Secret }}
  client_key '{{ .SecretPath "tls.key"}}'
  client_cert '{{ .SecretPath "tls.crt"}}'
  ca_file '{{ .SecretPath "ca-bundle.crt"}}'
{{- end }}
  type_name _doc
{{- if .Hints.Has "include_retry_tag" }}
  retry_tag {{.RetryTag}}
{{- end }}
  write_operation create
  reload_connections "#{ENV['ES_RELOAD_CONNECTIONS'] || 'true'}"
  # https://github.com/uken/fluent-plugin-elasticsearch#reload-after
  reload_after "#{ENV['ES_RELOAD_AFTER'] || '200'}"
  # https://github.com/uken/fluent-plugin-elasticsearch#sniffer-class-name
  sniffer_class_name "#{ENV['ES_SNIFFER_CLASS_NAME'] || 'Fluent::Plugin::ElasticsearchSimpleSniffer'}"
  reload_on_failure false
  # 2 ^ 31
  request_timeout 2147483648
  <buffer>
    @type file
    path '{{.BufferPath}}'
    flush_interval "#{ENV['ES_FLUSH_INTERVAL'] || '1s'}"
    flush_thread_count "#{ENV['ES_FLUSH_THREAD_COUNT'] || 2}"
    flush_at_shutdown "#{ENV['FLUSH_AT_SHUTDOWN'] || 'false'}"
    retry_max_interval "#{ENV['ES_RETRY_WAIT'] || '300'}"
    retry_forever true
    queue_limit_length "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
    chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m' }"
    overflow_action "#{ENV['BUFFER_QUEUE_FULL_ACTION'] || 'block'}"
  </buffer>
</store>
{{- end}}`

const storeSyslogTemplateOld = `{{- define "storeSyslogOld" -}}
<store>
  @type {{.SyslogPlugin}}
  @id {{.StoreID}}
  remote_syslog {{.Host}}
  port {{.Port}}
  hostname ${hostname}
  facility user
  severity debug
</store>
{{- end}}`

//      hostname ${hostname}
const storeSyslogTemplate = `{{- define "storeSyslog" -}}
<store>
  @type remote_syslog
  @id {{.StoreID}}
  host {{.Host}}
  port {{.Port}}
  rfc {{.Rfc}}
  facility user
  severity debug
  program fluentd
  protocol {{.Protocol}}
  packet_size 4096
{{ if .Target.Secret -}}
  tls true
  ca_file '{{ .SecretPath "ca-bundle.crt"}}'
  verify_mode true
{{ end -}}
{{ if (eq .Protocol "tcp") -}}
  timeout 60
  timeout_exception true
  keep_alive true
  keep_alive_idle 75
  keep_alive_cnt 9
  keep_alive_intvl 7200
{{ end -}}
  <buffer>
    @type file
    path '{{.BufferPath}}'
    flush_interval "#{ENV['ES_FLUSH_INTERVAL'] || '1s'}"
    flush_thread_count "#{ENV['ES_FLUSH_THREAD_COUNT'] || 2}"
    flush_at_shutdown "#{ENV['FLUSH_AT_SHUTDOWN'] || 'false'}"
    retry_max_interval "#{ENV['ES_RETRY_WAIT'] || '300'}"
    retry_forever true
    queue_limit_length "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
    chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m' }"
    overflow_action "#{ENV['BUFFER_QUEUE_FULL_ACTION'] || 'block'}"
  </buffer>
</store>
{{- end}}`
