package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/sets"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	. "github.com/openshift/cluster-logging-operator/test"
)

var _ = Describe("generating source", func() {

	var (
		generator *ConfigGenerator
		err       error
		results   []string
	)

	BeforeEach(func() {
		generator, err = NewConfigGenerator(false, false, true)
		Expect(err).To(BeNil())
	})

	Context("for only logs-app source", func() {
		BeforeEach(func() {
			results, err = generator.generateSource(sets.NewString(string(logging.LogSourceTypeApp)), sets.NewString())
			Expect(err).To(BeNil())
			Expect(len(results) == 1).To(BeTrue())
		})

		It("should produce a container config", func() {
			Expect(results[0]).To(EqualTrimLines(`# container logs
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
		  `))
		})
	})

	Context("for only logs-infra source", func() {
		BeforeEach(func() {
			results, err = generator.generateSource(sets.NewString(string(logging.LogSourceTypeInfra)), sets.NewString())
			Expect(err).To(BeNil())
			Expect(len(results) == 1).To(BeTrue())
		})

		It("should produce a journal config", func() {
			Expect(results[0]).To(EqualTrimLines(`
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
		  `))
		})
	})

	Context("for only logs-audit source", func() {
		BeforeEach(func() {
			results, err = generator.generateSource(sets.NewString(string(logging.LogSourceTypeAudit)), sets.NewString())
			Expect(err).To(BeNil())
			Expect(len(results)).To(Equal(3))
		})

		It("should produce configs for the audit logs", func() {
			Expect(results[0]).To(EqualTrimLines(`
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
		  `))
			Expect(results[1]).To(EqualTrimLines(`
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
		  `))
			Expect(results[2]).To(EqualTrimLines(`
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
		  `))
		})
	})

	Context("for all log sources", func() {

		BeforeEach(func() {
			results, err = generator.generateSource(sets.NewString(string(logging.LogSourceTypeApp), string(logging.LogSourceTypeInfra), string(logging.LogSourceTypeAudit)), sets.NewString())
			Expect(err).To(BeNil())
			Expect(len(results)).To(Equal(5))
		})
		Context("for journal input", func() {

			It("should produce a config with no exclusions", func() {
				Expect(results[0]).To(EqualTrimLines(`
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
			</source>`))
			})
		})

		Context("for container inputs", func() {

			It("should produce a config", func() {
				Expect(results[1]).To(EqualTrimLines(`# container logs
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
			  `))
			})
		})

		Context("for audit inputs", func() {

			It("should produce a config with no exclusions", func() {
				Expect(results[2]).To(EqualTrimLines(`
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
		    `))
				Expect(results[3]).To(EqualTrimLines(`
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
		    `))
				Expect(results[4]).To(EqualTrimLines(`
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
		    `))
			})
		})
	})

})
