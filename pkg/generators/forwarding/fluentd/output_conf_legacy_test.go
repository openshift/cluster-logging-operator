package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	v1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("Generating fluentd legacy output store config blocks", func() {
	var (
		err       error
		clfSpec   *loggingv1.ClusterLogForwarderSpec
		fwSpec    *v1.ForwarderSpec
		generator *ConfigGenerator
	)

	Context("based on legacy fluentd forward method", func() {
		BeforeEach(func() {
			clfSpec = &loggingv1.ClusterLogForwarderSpec{}
			generator, err = NewConfigGenerator(true, false, false)
			Expect(err).To(BeNil())
		})

		It("should produce well formed legacy output label config", func() {
			legacyConf := `
        ## CLO GENERATED CONFIGURATION ###
        # This file is a copy of the fluentd configuration entrypoint
        # which should normally be supplied in a configmap.

        <system>
          log_level "#{ENV['LOG_LEVEL'] || 'warn'}"
        </system>

        # In each section below, pre- and post- includes don't include anything initially;
        # they exist to enable future additions to openshift conf as needed.

        ## sources
        ## ordered so that syslog always runs last...
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

        # This is considered experimental by the repo
        <source>
          @type prometheus_output_monitor
          <labels>
            hostname ${hostname}
          </labels>
        </source>
        #journal logs to gather node
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
            path '/var/log/journal_pos.json'
          </storage>
          matches "#{ENV['JOURNAL_FILTERS_JSON'] || '[]'}"
          tag journal
          read_from_head "#{if (val = ENV.fetch('JOURNAL_READ_FROM_HEAD','')) && (val.length > 0); val; else 'false'; end}"
        </source>
        # container logs
        <source>
          @type tail
          @id container-input
          path "/var/log/containers/*.log"
          exclude_path ["/var/log/containers/fluentd-*_openshift-logging_*.log", "/var/log/containers/elasticsearch-*_openshift-logging_*.log", "/var/log/containers/kibana-*_openshift-logging_*.log"]
          pos_file "/var/log/es-containers.log.pos"
          refresh_interval 5
          rotate_wait 5
          tag kubernetes.*
          read_from_head "true"
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
          path "#{ENV['AUDIT_FILE'] || '/var/log/audit/audit.log'}"
          pos_file "#{ENV['AUDIT_POS_FILE'] || '/var/log/audit/audit.log.pos'}"
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

        # Openshift audit logs
        <source>
          @type tail
          @id openshift-audit-input
          @label @MEASURE
          path /var/log/oauth-apiserver/audit.log,/var/log/openshift-apiserver/audit.log
          pos_file /var/log/oauth-apiserver.audit.log
          tag openshift-audit.log
          <parse>
            @type json
            time_key requestReceivedTimestamp
            # In case folks want to parse based on the requestReceivedTimestamp key
            keep_time_key true
            time_format %Y-%m-%dT%H:%M:%S.%N%z
          </parse>
        </source>
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
        <match journal>
          @type relabel
          @label @INGRESS
        </match>
        <match *audit.log>
          @type relabel
          @label @INGRESS
         </match>
        <match kubernetes.**>
          @type relabel
          @label @CONCAT
        </match>
      </label>
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
            kubernetes_url 'https://kubernetes.default.svc'
            cache_size '1000'
            watch 'false'
            use_journal 'nil'
            ssl_partial_chain 'true'
          </filter>

          <filter kubernetes.journal.**>
            @type parse_json_field
            merge_json_log 'false'
            preserve_json_log 'true'
            json_fields 'log,MESSAGE'
          </filter>

          <filter kubernetes.var.log.containers.**>
            @type parse_json_field
            merge_json_log 'false'
            preserve_json_log 'true'
            json_fields 'log,MESSAGE'
          </filter>

          <filter kubernetes.var.log.containers.eventrouter-** kubernetes.var.log.containers.cluster-logging-eventrouter-**>
            @type parse_json_field
            merge_json_log true
            preserve_json_log true
            json_fields 'log,MESSAGE'
          </filter>

          <filter **kibana**>
            @type record_transformer
            enable_ruby
            <record>
              log ${record['err'] || record['msg'] || record['MESSAGE'] || record['log']}
            </record>
            remove_keys req,res,msg,name,level,v,pid,err
          </filter>

		  <filter k8s-audit.log**>
			@type record_modifier
			<record>
			  k8s_audit_level ${record['level']}
			  level info
			</record>
		  </filter>
		  <filter openshift-audit.log**>
			@type record_modifier
			<record>
			  openshift_audit_level ${record['level']}
			  level info
			</record>
		  </filter>

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
              tag "kubernetes.var.log.containers.eventrouter-** kubernetes.var.log.containers.cluster-logging-eventrouter-** k8s-audit.log** openshift-audit.log**"
              type k8s_json_file
              remove_keys log,stream,CONTAINER_ID_FULL,CONTAINER_NAME
              process_kubernetes_events 'true'
            </formatter>
            <formatter>
              tag "kubernetes.var.log.containers**"
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
              tag "linux-audit.log** k8s-audit.log** openshift-audit.log**"
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

          <filter **>
            @type elasticsearch_genid_ext
            hash_id_key viaq_msg_id
            alt_key kubernetes.event.metadata.uid
            alt_tags 'kubernetes.var.log.containers.logging-eventrouter-*.** kubernetes.var.log.containers.eventrouter-*.** kubernetes.var.log.containers.cluster-logging-eventrouter-*.** kubernetes.journal.container._default_.kubernetes-event'
          </filter>

          # Relabel specific source tags to specific intermediary labels for copy processing
          # Earlier matchers remove logs so they don't fall through to later ones.
          # A log source matcher may be null if no pipeline wants that type of log.
          <match **_default_** **_kube-*_** **_openshift-*_** **_openshift_** journal.** system.var.log**>
            @type relabel
            @label @_INFRASTRUCTURE
          </match>

          <match kubernetes.**>
            @type relabel
            @label @_APPLICATION
          </match>

          <match linux-audit.log** k8s-audit.log** openshift-audit.log**>
            @type relabel
            @label @_AUDIT
          </match>

          <match **>
            @type stdout
          </match>

        </label>

        # Relabel specific sources (e.g. logs.apps) to multiple pipelines
        <label @_APPLICATION>
          <match **>
            @type copy
            <store>
              @type relabel
              @label @_LEGACY_SECUREFORWARD
            </store>
          </match>
        </label>
        <label @_AUDIT>
          <match **>
            @type copy
            <store>
              @type relabel
              @label @_LEGACY_SECUREFORWARD
            </store>
          </match>
        </label>
        <label @_INFRASTRUCTURE>
          <match **>
            @type copy
            <store>
              @type relabel
              @label @_LEGACY_SECUREFORWARD
            </store>
          </match>
        </label>

        # Ship logs to specific outputs

        <label @_LEGACY_SECUREFORWARD>
          <match **>
            @type copy
            #include legacy secure-forward.conf
            @include /etc/fluent/configs.d/secure-forward/secure-forward.conf
          </match>
        </label>
      `

			results, err := generator.Generate(clfSpec, nil, fwSpec)
			Expect(err).Should(Succeed())
			Expect(results).To(EqualTrimLines(legacyConf))
		})
	})

	Context("based on legacy syslog method", func() {
		BeforeEach(func() {
			clfSpec = &loggingv1.ClusterLogForwarderSpec{}
			generator, err = NewConfigGenerator(false, true, false)
			Expect(err).To(BeNil())
		})

		It("should produce well formed legacy output label config", func() {
			legacyConf := `
        ## CLO GENERATED CONFIGURATION ###
        # This file is a copy of the fluentd configuration entrypoint
        # which should normally be supplied in a configmap.

        <system>
          log_level "#{ENV['LOG_LEVEL'] || 'warn'}"
        </system>

        # In each section below, pre- and post- includes don't include anything initially;
        # they exist to enable future additions to openshift conf as needed.

        ## sources
        ## ordered so that syslog always runs last...
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

        # This is considered experimental by the repo
        <source>
          @type prometheus_output_monitor
          <labels>
            hostname ${hostname}
          </labels>
        </source>
        #journal logs to gather node
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
            path '/var/log/journal_pos.json'
          </storage>
          matches "#{ENV['JOURNAL_FILTERS_JSON'] || '[]'}"
          tag journal
          read_from_head "#{if (val = ENV.fetch('JOURNAL_READ_FROM_HEAD','')) && (val.length > 0); val; else 'false'; end}"
        </source>
        # container logs
        <source>
          @type tail
          @id container-input
          path "/var/log/containers/*.log"
          exclude_path ["/var/log/containers/fluentd-*_openshift-logging_*.log", "/var/log/containers/elasticsearch-*_openshift-logging_*.log", "/var/log/containers/kibana-*_openshift-logging_*.log"]
          pos_file "/var/log/es-containers.log.pos"
          refresh_interval 5
          rotate_wait 5
          tag kubernetes.*
          read_from_head "true"
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
          path "#{ENV['AUDIT_FILE'] || '/var/log/audit/audit.log'}"
          pos_file "#{ENV['AUDIT_POS_FILE'] || '/var/log/audit/audit.log.pos'}"
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

        # Openshift audit logs
        <source>
          @type tail
          @id openshift-audit-input
          @label @MEASURE
          path /var/log/oauth-apiserver/audit.log,/var/log/openshift-apiserver/audit.log
          pos_file /var/log/oauth-apiserver.audit.log
          tag openshift-audit.log
          <parse>
            @type json
            time_key requestReceivedTimestamp
            # In case folks want to parse based on the requestReceivedTimestamp key
            keep_time_key true
            time_format %Y-%m-%dT%H:%M:%S.%N%z
          </parse>
        </source>
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
        <match journal>
          @type relabel
          @label @INGRESS
        </match>
        <match *audit.log>
          @type relabel
          @label @INGRESS
         </match>
        <match kubernetes.**>
          @type relabel
          @label @CONCAT
        </match>
      </label>
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
            kubernetes_url 'https://kubernetes.default.svc'
            cache_size '1000'
            watch 'false'
            use_journal 'nil'
            ssl_partial_chain 'true'
          </filter>

          <filter kubernetes.journal.**>
            @type parse_json_field
            merge_json_log 'false'
            preserve_json_log 'true'
            json_fields 'log,MESSAGE'
          </filter>

          <filter kubernetes.var.log.containers.**>
            @type parse_json_field
            merge_json_log 'false'
            preserve_json_log 'true'
            json_fields 'log,MESSAGE'
          </filter>

          <filter kubernetes.var.log.containers.eventrouter-** kubernetes.var.log.containers.cluster-logging-eventrouter-**>
            @type parse_json_field
            merge_json_log true
            preserve_json_log true
            json_fields 'log,MESSAGE'
          </filter>

          <filter **kibana**>
            @type record_transformer
            enable_ruby
            <record>
              log ${record['err'] || record['msg'] || record['MESSAGE'] || record['log']}
            </record>
            remove_keys req,res,msg,name,level,v,pid,err
          </filter>

		  <filter k8s-audit.log**>
			@type record_modifier
			<record>
			  k8s_audit_level ${record['level']}
			  level info
			</record>
		  </filter>
		  <filter openshift-audit.log**>
			@type record_modifier
			<record>
			  openshift_audit_level ${record['level']}
			  level info
			</record>
		  </filter>
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
              tag "kubernetes.var.log.containers.eventrouter-** kubernetes.var.log.containers.cluster-logging-eventrouter-** k8s-audit.log** openshift-audit.log**"
              type k8s_json_file
              remove_keys log,stream,CONTAINER_ID_FULL,CONTAINER_NAME
              process_kubernetes_events 'true'
            </formatter>
            <formatter>
              tag "kubernetes.var.log.containers**"
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
              tag "linux-audit.log** k8s-audit.log** openshift-audit.log**"
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

          <filter **>
            @type elasticsearch_genid_ext
            hash_id_key viaq_msg_id
            alt_key kubernetes.event.metadata.uid
            alt_tags 'kubernetes.var.log.containers.logging-eventrouter-*.** kubernetes.var.log.containers.eventrouter-*.** kubernetes.var.log.containers.cluster-logging-eventrouter-*.** kubernetes.journal.container._default_.kubernetes-event'
          </filter>

          # Relabel specific source tags to specific intermediary labels for copy processing
					# Earlier matchers remove logs so they don't fall through to later ones.
					# A log source matcher may be null if no pipeline wants that type of log.

          <match **_default_** **_kube-*_** **_openshift-*_** **_openshift_** journal.** system.var.log**>
            @type relabel
            @label @_INFRASTRUCTURE
          </match>

          <match kubernetes.**>
            @type relabel
            @label @_APPLICATION
          </match>

          <match linux-audit.log** k8s-audit.log** openshift-audit.log**>
            @type relabel
            @label @_AUDIT
          </match>

          <match **>
            @type stdout
          </match>

        </label>

        # Relabel specific sources (e.g. logs.apps) to multiple pipelines
        <label @_APPLICATION>
          <match **>
            @type copy



            <store>
              @type relabel
              @label @_LEGACY_SYSLOG
            </store>
          </match>
        </label>
        <label @_AUDIT>
          <match **>
            @type copy



            <store>
              @type relabel
              @label @_LEGACY_SYSLOG
            </store>
          </match>
        </label>
        <label @_INFRASTRUCTURE>
          <match **>
            @type copy



            <store>
              @type relabel
              @label @_LEGACY_SYSLOG
            </store>
          </match>
        </label>

        # Ship logs to specific outputs


        <label @_LEGACY_SYSLOG>
          <match **>
            @type copy
            #include legacy Syslog
            @include /etc/fluent/configs.d/syslog/syslog.conf
          </match>
        </label>
      `

			results, err := generator.Generate(clfSpec, nil, fwSpec)
			Expect(err).Should(Succeed())
			Expect(results).To(EqualTrimLines(legacyConf))
		})
	})

	Context("based on legacy syslog and fluentd forward method", func() {
		BeforeEach(func() {
			clfSpec = &loggingv1.ClusterLogForwarderSpec{}
			generator, err = NewConfigGenerator(true, true, false)
			Expect(err).To(BeNil())
		})

		It("should produce well formed legacy output label config", func() {
			legacyConf := `
        ## CLO GENERATED CONFIGURATION ###
        # This file is a copy of the fluentd configuration entrypoint
        # which should normally be supplied in a configmap.

        <system>
          log_level "#{ENV['LOG_LEVEL'] || 'warn'}"
        </system>

        # In each section below, pre- and post- includes don't include anything initially;
        # they exist to enable future additions to openshift conf as needed.

        ## sources
        ## ordered so that syslog always runs last...
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

        # This is considered experimental by the repo
        <source>
          @type prometheus_output_monitor
          <labels>
            hostname ${hostname}
          </labels>
        </source>
        #journal logs to gather node
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
            path '/var/log/journal_pos.json'
          </storage>
          matches "#{ENV['JOURNAL_FILTERS_JSON'] || '[]'}"
          tag journal
          read_from_head "#{if (val = ENV.fetch('JOURNAL_READ_FROM_HEAD','')) && (val.length > 0); val; else 'false'; end}"
        </source>
        # container logs
        <source>
          @type tail
          @id container-input
          path "/var/log/containers/*.log"
          exclude_path ["/var/log/containers/fluentd-*_openshift-logging_*.log", "/var/log/containers/elasticsearch-*_openshift-logging_*.log", "/var/log/containers/kibana-*_openshift-logging_*.log"]
          pos_file "/var/log/es-containers.log.pos"
          refresh_interval 5
          rotate_wait 5
          tag kubernetes.*
          read_from_head "true"
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
          path "#{ENV['AUDIT_FILE'] || '/var/log/audit/audit.log'}"
          pos_file "#{ENV['AUDIT_POS_FILE'] || '/var/log/audit/audit.log.pos'}"
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

        # Openshift audit logs
        <source>
          @type tail
          @id openshift-audit-input
          @label @MEASURE
          path /var/log/oauth-apiserver/audit.log,/var/log/openshift-apiserver/audit.log
          pos_file /var/log/oauth-apiserver.audit.log
          tag openshift-audit.log
          <parse>
            @type json
            time_key requestReceivedTimestamp
            # In case folks want to parse based on the requestReceivedTimestamp key
            keep_time_key true
            time_format %Y-%m-%dT%H:%M:%S.%N%z
          </parse>
        </source>
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
        <match journal>
          @type relabel
          @label @INGRESS
        </match>
        <match *audit.log>
          @type relabel
          @label @INGRESS
         </match>
        <match kubernetes.**>
          @type relabel
          @label @CONCAT
        </match>
      </label>
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
            kubernetes_url 'https://kubernetes.default.svc'
            cache_size '1000'
            watch 'false'
            use_journal 'nil'
            ssl_partial_chain 'true'
          </filter>

          <filter kubernetes.journal.**>
            @type parse_json_field
            merge_json_log 'false'
            preserve_json_log 'true'
            json_fields 'log,MESSAGE'
          </filter>

          <filter kubernetes.var.log.containers.**>
            @type parse_json_field
            merge_json_log 'false'
            preserve_json_log 'true'
            json_fields 'log,MESSAGE'
          </filter>

          <filter kubernetes.var.log.containers.eventrouter-** kubernetes.var.log.containers.cluster-logging-eventrouter-**>
            @type parse_json_field
            merge_json_log true
            preserve_json_log true
            json_fields 'log,MESSAGE'
          </filter>

          <filter **kibana**>
            @type record_transformer
            enable_ruby
            <record>
              log ${record['err'] || record['msg'] || record['MESSAGE'] || record['log']}
            </record>
            remove_keys req,res,msg,name,level,v,pid,err
          </filter>

		  <filter k8s-audit.log**>
			@type record_modifier
			<record>
			  k8s_audit_level ${record['level']}
			  level info
			</record>
		  </filter>
		  <filter openshift-audit.log**>
			@type record_modifier
			<record>
			  openshift_audit_level ${record['level']}
			  level info
			</record>
		  </filter>

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
              tag "kubernetes.var.log.containers.eventrouter-** kubernetes.var.log.containers.cluster-logging-eventrouter-** k8s-audit.log** openshift-audit.log**"
              type k8s_json_file
              remove_keys log,stream,CONTAINER_ID_FULL,CONTAINER_NAME
              process_kubernetes_events 'true'
            </formatter>
            <formatter>
              tag "kubernetes.var.log.containers**"
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
              tag "linux-audit.log** k8s-audit.log** openshift-audit.log**"
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

          <filter **>
            @type elasticsearch_genid_ext
            hash_id_key viaq_msg_id
            alt_key kubernetes.event.metadata.uid
            alt_tags 'kubernetes.var.log.containers.logging-eventrouter-*.** kubernetes.var.log.containers.eventrouter-*.** kubernetes.var.log.containers.cluster-logging-eventrouter-*.** kubernetes.journal.container._default_.kubernetes-event'
          </filter>

          # Relabel specific source tags to specific intermediary labels for copy processing
          # Earlier matchers remove logs so they don't fall through to later ones.
          # A log source matcher may be null if no pipeline wants that type of log.
          <match **_default_** **_kube-*_** **_openshift-*_** **_openshift_** journal.** system.var.log**>
            @type relabel
            @label @_INFRASTRUCTURE
          </match>

          <match kubernetes.**>
            @type relabel
            @label @_APPLICATION
          </match>

          <match linux-audit.log** k8s-audit.log** openshift-audit.log**>
            @type relabel
            @label @_AUDIT
          </match>

          <match **>
            @type stdout
          </match>

        </label>

        # Relabel specific sources (e.g. logs.apps) to multiple pipelines
        <label @_APPLICATION>
          <match **>
            @type copy


            <store>
              @type relabel
              @label @_LEGACY_SECUREFORWARD
            </store>

            <store>
              @type relabel
              @label @_LEGACY_SYSLOG
            </store>
          </match>
        </label>
        <label @_AUDIT>
          <match **>
            @type copy


            <store>
              @type relabel
              @label @_LEGACY_SECUREFORWARD
            </store>

            <store>
              @type relabel
              @label @_LEGACY_SYSLOG
            </store>
          </match>
        </label>
        <label @_INFRASTRUCTURE>
          <match **>
            @type copy


            <store>
              @type relabel
              @label @_LEGACY_SECUREFORWARD
            </store>

            <store>
              @type relabel
              @label @_LEGACY_SYSLOG
            </store>
          </match>
        </label>

        # Ship logs to specific outputs

        <label @_LEGACY_SECUREFORWARD>
          <match **>
            @type copy
            #include legacy secure-forward.conf
            @include /etc/fluent/configs.d/secure-forward/secure-forward.conf
          </match>
        </label>

        <label @_LEGACY_SYSLOG>
          <match **>
            @type copy
            #include legacy Syslog
            @include /etc/fluent/configs.d/syslog/syslog.conf
          </match>
        </label>
      `

			results, err := generator.Generate(clfSpec, nil, fwSpec)
			Expect(err).Should(Succeed())
			Expect(results).To(EqualTrimLines(legacyConf))
		})
	})
})
