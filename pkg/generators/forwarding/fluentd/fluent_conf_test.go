package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	test "github.com/openshift/cluster-logging-operator/test"
)

var _ = Describe("Generating fluentd config", func() {
	var (
		forwarding *logging.ForwardingSpec
		generator  *ConfigGenerator
	)
	BeforeEach(func() {
		var err error
		generator, err = NewConfigGenerator()
		Expect(err).To(BeNil())
		Expect(generator).ToNot(BeNil())
		forwarding = &logging.ForwardingSpec{
			Outputs: []logging.OutputSpec{
				{
					Type:     logging.OutputTypeElasticsearch,
					Name:     "infra-es",
					Endpoint: "es.svc.infra.cluster:9999",
					Secret: &logging.OutputSecretSpec{
						Name: "my-infra-secret",
					},
				},
				{
					Type:     logging.OutputTypeElasticsearch,
					Name:     "apps-es-1",
					Endpoint: "es.svc.messaging.cluster.local:9654",
					Secret: &logging.OutputSecretSpec{
						Name: "my-es-secret",
					},
				},
				{
					Type:     logging.OutputTypeElasticsearch,
					Name:     "apps-es-2",
					Endpoint: "es.svc.messaging.cluster.local2:9654",
					Secret: &logging.OutputSecretSpec{
						Name: "my-other-secret",
					},
				},
			},
			Pipelines: []logging.PipelineSpec{
				logging.PipelineSpec{
					Name:       "infra-pipeline",
					SourceType: logging.LogSourceTypeInfra,
					OutputRefs: []string{"infra-es"},
				},
				logging.PipelineSpec{
					Name:       "apps-pipeline",
					SourceType: logging.LogSourceTypeApp,
					OutputRefs: []string{"apps-es-1", "apps-es-2"},
				},
			},
		}
	})

	It("should produce well formed fluent.conf", func() {
		results, err := generator.Generate(forwarding)
		Expect(err).To(BeNil())
		test.Expect(results).ToEqual(`
			## CLO GENERATED CONFIGURATION ### 
			# This file is a copy of the fluentd configuration entrypoint
			# which should normally be supplied in a configmap.
			
			<system>
  				@log_level warn
			</system>
			
			# In each section below, pre- and post- includes don't include anything initially;
			# they exist to enable future additions to openshift conf as needed.
			
			## sources
			## ordered so that syslog always runs last...
			<source>
				@type prometheus
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

			@include configs.d/dynamic/input-docker-*.conf
			
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
					<formatter>
						enabled false
						tag "audit.log**"
						type sys_var_log
					</formatter>
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
						tag "kubernetes.var.log.containers**"
						type k8s_json_file
						remove_keys log,stream,CONTAINER_ID_FULL,CONTAINER_NAME
					</formatter>
					<elasticsearch_index_name>
						enabled "#{ENV['ENABLE_ES_INDEX_NAME'] || 'true'}"
						tag "journal.system** system.var.log** audit.log** %OCP_FLUENTD_TAGS%"
						name_type operations_full
					</elasticsearch_index_name>
					<elasticsearch_index_name>
						enabled "#{ENV['ENABLE_ES_INDEX_NAME'] || 'true'}"
						tag "**"
						name_type project_full
					</elasticsearch_index_name>
				</filter>
				<filter **>
					@type elasticsearch_genid_ext
					hash_id_key viaq_msg_id
					alt_key kubernetes.event.metadata.uid
					alt_tags "#{ENV['GENID_ALT_TAG'] || 'kubernetes.var.log.containers.logging-eventrouter-*.** kubernetes.journal.container._default_.kubernetes-event'}"
				</filter>

				# Relabel specific source tags to specific intermediary labels for copy processing
				<match **_default_** **_kube-*_** **_openshift-*_** **_openshift_** journal.** system.var.log**>
					@type relabel
					@label @_LOGS_INFRA
				</match>
				<match kubernetes.**>
					@type relabel
					@label @_LOGS_APP
				</match>
				<match **>
					@type stdout
				</match>

			</label>
			
			# Relabel specific sources (e.g. logs.apps) to multiple pipelines
			<label @_LOGS_APP>
				<match **>
					@type copy
					<store>
						@type relabel
						@label @APPS_PIPELINE
					</store>
				</match>
			</label>
			<label @_LOGS_INFRA>
				<match **>
					@type copy
					<store>
						@type relabel
						@label @INFRA_PIPELINE
					</store>
				</match>
			</label>

			# Relabel specific pipelines to multiple, outputs (e.g. ES, kafka stores)
			<label @INFRA_PIPELINE>
				<match **>
					@type copy
					<store>
						@type relabel
						@label @INFRA_ES
					</store>
				</match>
			</label>
			<label @APPS_PIPELINE>
				<match **>
					@type copy
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

			# Ship logs to specific outputs
			<label @INFRA_ES>
				<match retry_infra_es>
					@type copy
					<store>
						@type elasticsearch
						@id retry_infra_es
						host es.svc.infra.cluster
						port 9999
						scheme https
						ssl_version TLSv1_2
						target_index_key viaq_index_name
						id_key viaq_msg_id
						remove_keys viaq_index_name
						user fluentd
						password changeme

						client_key '/var/run/ocp-collector/secrets/my-infra-secret/tls.key'
						client_cert '/var/run/ocp-collector/secrets/my-infra-secret/tls.crt'
						ca_file '/var/run/ocp-collector/secrets/my-infra-secret/ca-bundle.crt'
						type_name com.redhat.viaq.common
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
							path '/var/lib/fluentd/retry_infra_es'
							flush_interval "#{ENV['ES_FLUSH_INTERVAL'] || '1s'}"
							flush_thread_count "#{ENV['ES_FLUSH_THREAD_COUNT'] || 2}"
							flush_at_shutdown "#{ENV['FLUSH_AT_SHUTDOWN'] || 'false'}"
							retry_max_interval "#{ENV['ES_RETRY_WAIT'] || '300'}"
							retry_forever true
							queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
							chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m' }"
							overflow_action "#{ENV['BUFFER_QUEUE_FULL_ACTION'] || 'block'}"
						</buffer>
					</store>
				</match>
				<match **>
					@type copy
					<store>
						@type elasticsearch
						@id infra_es
						host es.svc.infra.cluster
						port 9999
						scheme https
						ssl_version TLSv1_2
						target_index_key viaq_index_name
						id_key viaq_msg_id
						remove_keys viaq_index_name
						user fluentd
						password changeme

						client_key '/var/run/ocp-collector/secrets/my-infra-secret/tls.key'
						client_cert '/var/run/ocp-collector/secrets/my-infra-secret/tls.crt'
						ca_file '/var/run/ocp-collector/secrets/my-infra-secret/ca-bundle.crt'
						type_name com.redhat.viaq.common
						retry_tag retry_infra_es
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
							path '/var/lib/fluentd/infra_es'
							flush_interval "#{ENV['ES_FLUSH_INTERVAL'] || '1s'}"
							flush_thread_count "#{ENV['ES_FLUSH_THREAD_COUNT'] || 2}"
							flush_at_shutdown "#{ENV['FLUSH_AT_SHUTDOWN'] || 'false'}"
							retry_max_interval "#{ENV['ES_RETRY_WAIT'] || '300'}"
							retry_forever true
							queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
							chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m' }"
							overflow_action "#{ENV['BUFFER_QUEUE_FULL_ACTION'] || 'block'}"
						</buffer>
					</store>
				</match>
			</label>
			<label @APPS_ES_1>
				<match retry_apps_es_1>
					@type copy
					<store>
						@type elasticsearch
						@id retry_apps_es_1
						host es.svc.messaging.cluster.local
						port 9654
						scheme https
						ssl_version TLSv1_2
						target_index_key viaq_index_name
						id_key viaq_msg_id
						remove_keys viaq_index_name
						user fluentd
						password changeme

						client_key '/var/run/ocp-collector/secrets/my-es-secret/tls.key'
						client_cert '/var/run/ocp-collector/secrets/my-es-secret/tls.crt'
						ca_file '/var/run/ocp-collector/secrets/my-es-secret/ca-bundle.crt'
						type_name com.redhat.viaq.common
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
							path '/var/lib/fluentd/retry_apps_es_1'
							flush_interval "#{ENV['ES_FLUSH_INTERVAL'] || '1s'}"
							flush_thread_count "#{ENV['ES_FLUSH_THREAD_COUNT'] || 2}"
							flush_at_shutdown "#{ENV['FLUSH_AT_SHUTDOWN'] || 'false'}"
							retry_max_interval "#{ENV['ES_RETRY_WAIT'] || '300'}"
							retry_forever true
							queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
							chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m' }"
							overflow_action "#{ENV['BUFFER_QUEUE_FULL_ACTION'] || 'block'}"
						</buffer>
					</store>
				</match>
				<match **>
					@type copy
					<store>
						@type elasticsearch
						@id apps_es_1
						host es.svc.messaging.cluster.local
						port 9654
						scheme https
						ssl_version TLSv1_2
						target_index_key viaq_index_name
						id_key viaq_msg_id
						remove_keys viaq_index_name
						user fluentd
						password changeme

						client_key '/var/run/ocp-collector/secrets/my-es-secret/tls.key'
						client_cert '/var/run/ocp-collector/secrets/my-es-secret/tls.crt'
						ca_file '/var/run/ocp-collector/secrets/my-es-secret/ca-bundle.crt'
						type_name com.redhat.viaq.common
						retry_tag retry_apps_es_1
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
							path '/var/lib/fluentd/apps_es_1'
							flush_interval "#{ENV['ES_FLUSH_INTERVAL'] || '1s'}"
							flush_thread_count "#{ENV['ES_FLUSH_THREAD_COUNT'] || 2}"
							flush_at_shutdown "#{ENV['FLUSH_AT_SHUTDOWN'] || 'false'}"
							retry_max_interval "#{ENV['ES_RETRY_WAIT'] || '300'}"
							retry_forever true
							queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
							chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m' }"
							overflow_action "#{ENV['BUFFER_QUEUE_FULL_ACTION'] || 'block'}"
						</buffer>
					</store>
				</match>
			</label>
			<label @APPS_ES_2>
				<match retry_apps_es_2>
					@type copy
					<store>
						@type elasticsearch
						@id retry_apps_es_2
						host es.svc.messaging.cluster.local2
						port 9654
						scheme https
						ssl_version TLSv1_2
						target_index_key viaq_index_name
						id_key viaq_msg_id
						remove_keys viaq_index_name
						user fluentd
						password changeme

						client_key '/var/run/ocp-collector/secrets/my-other-secret/tls.key'
						client_cert '/var/run/ocp-collector/secrets/my-other-secret/tls.crt'
						ca_file '/var/run/ocp-collector/secrets/my-other-secret/ca-bundle.crt'
						type_name com.redhat.viaq.common
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
							path '/var/lib/fluentd/retry_apps_es_2'
							flush_interval "#{ENV['ES_FLUSH_INTERVAL'] || '1s'}"
							flush_thread_count "#{ENV['ES_FLUSH_THREAD_COUNT'] || 2}"
							flush_at_shutdown "#{ENV['FLUSH_AT_SHUTDOWN'] || 'false'}"
							retry_max_interval "#{ENV['ES_RETRY_WAIT'] || '300'}"
							retry_forever true
							queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
							chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m' }"
							overflow_action "#{ENV['BUFFER_QUEUE_FULL_ACTION'] || 'block'}"
						</buffer>
					</store>
				</match>
				<match **>
					@type copy
					<store>
						@type elasticsearch
						@id apps_es_2
						host es.svc.messaging.cluster.local2
						port 9654
						scheme https
						ssl_version TLSv1_2
						target_index_key viaq_index_name
						id_key viaq_msg_id
						remove_keys viaq_index_name
						user fluentd
						password changeme

						client_key '/var/run/ocp-collector/secrets/my-other-secret/tls.key'
						client_cert '/var/run/ocp-collector/secrets/my-other-secret/tls.crt'
						ca_file '/var/run/ocp-collector/secrets/my-other-secret/ca-bundle.crt'
						type_name com.redhat.viaq.common
						retry_tag retry_apps_es_2
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
							path '/var/lib/fluentd/apps_es_2'
							flush_interval "#{ENV['ES_FLUSH_INTERVAL'] || '1s'}"
							flush_thread_count "#{ENV['ES_FLUSH_THREAD_COUNT'] || 2}"
							flush_at_shutdown "#{ENV['FLUSH_AT_SHUTDOWN'] || 'false'}"
							retry_max_interval "#{ENV['ES_RETRY_WAIT'] || '300'}"
							retry_forever true
							queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
							chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m' }"
							overflow_action "#{ENV['BUFFER_QUEUE_FULL_ACTION'] || 'block'}"
						</buffer>
					</store>
				</match>
			</label>
			`)
	})

})
