package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("Generating external syslog server output store config blocks", func() {

	var (
		err           error
		outputs       []logging.OutputSpec
		forwarderSpec *logging.ForwarderSpec
		generator     *ConfigGenerator
	)
	Context("based on old syslog plugin", func() {

		BeforeEach(func() {
			generator, err = NewConfigGenerator(false, false, true)
			Expect(err).To(BeNil())
		})

		tcpConf := `<label @SYSLOG_RECEIVER>
		<filter **>
		  @type parse_json_field
		  json_fields  message
		  merge_json_log false
		  replace_json_log true
		</filter>
		<match **>
		@type copy
		<store>
			@type syslog_buffered
			@id syslog_receiver
			remote_syslog sl.svc.messaging.cluster.local
			port 9654
			hostname ${hostname}
			facility user
			severity debug
		</store>
		</match>
	</label>`

		udpConf := `<label @SYSLOG_RECEIVER>
		<filter **>
		  @type parse_json_field
		  json_fields  message
		  merge_json_log false
		  replace_json_log true
		</filter>
		<match **>
		@type copy
		<store>
			@type syslog
			@id syslog_receiver
			remote_syslog sl.svc.messaging.cluster.local
			port 9654
			hostname ${hostname}
			facility user
			severity debug
		</store>
		</match>
	</label>`

		Context("for protocol-less endpoint", func() {
			BeforeEach(func() {
				outputs = []logging.OutputSpec{
					{
						Type: "syslog",
						Name: "syslog-receiver",
						URL:  "tcp://sl.svc.messaging.cluster.local:9654",
					},
				}
			})
			It("should produce well formed output label config", func() {
				results, err := generator.generateOutputLabelBlocks(outputs, nil, forwarderSpec)
				Expect(err).To(BeNil())
				Expect(len(results)).To(Equal(1))
				Expect(results[0]).To(EqualTrimLines(tcpConf))
			})
		})

		Context("for tcp endpoint", func() {
			BeforeEach(func() {
				outputs = []logging.OutputSpec{
					{
						Type: "syslog",
						Name: "syslog-receiver",
						URL:  "tcp://sl.svc.messaging.cluster.local:9654",
					},
				}
			})
			It("should produce well formed output label config", func() {
				results, err := generator.generateOutputLabelBlocks(outputs, nil, forwarderSpec)
				Expect(err).To(BeNil())
				Expect(len(results)).To(Equal(1))
				Expect(results[0]).To(EqualTrimLines(tcpConf))
			})
		})

		Context("for udp endpoint", func() {
			BeforeEach(func() {
				outputs = []logging.OutputSpec{
					{
						Type: "syslog",
						Name: "syslog-receiver",
						URL:  "udp://sl.svc.messaging.cluster.local:9654",
					},
				}
			})
			It("should produce well formed output label config", func() {
				results, err := generator.generateOutputLabelBlocks(outputs, nil, forwarderSpec)
				Expect(err).To(BeNil())
				Expect(len(results)).To(Equal(1))
				Expect(results[0]).To(EqualTrimLines(udpConf))
			})
		})
	})

	Context("based on new syslog plugin", func() {
		BeforeEach(func() {
			generator, err = NewConfigGenerator(false, false, false)
			Expect(err).To(BeNil())
		})
		tcpConf := `<label @SYSLOG_RECEIVER>
  <filter **>
	@type parse_json_field
	json_fields  message
	merge_json_log false
	replace_json_log true
  </filter>
  <match **>
    @type copy
    <store>
    	@type remote_syslog
    	@id syslog_receiver
    	host sl.svc.messaging.cluster.local
    	port 9654
    	rfc rfc5424
    	facility user
        severity debug
    	protocol tcp
    	packet_size 4096
		hostname "#{ENV['NODE_NAME']}"
		timeout 60
		timeout_exception true
	    keep_alive true
        keep_alive_idle 75
        keep_alive_cnt 9
        keep_alive_intvl 7200
		<buffer >
        @type file
        path '/var/lib/fluentd/syslog_receiver'
        flush_mode interval
        flush_interval 1s
        flush_thread_count 2
        flush_at_shutdown true
        retry_type exponential_backoff
        retry_wait 1s
        retry_max_interval 60s
        retry_forever true
        queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
        total_limit_size "#{ENV['TOTAL_LIMIT_SIZE'] ||  8589934592 }" #8G
        chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
        overflow_action block
      </buffer>
    </store>
  </match>
</label>`
		udpConf := `<label @SYSLOG_RECEIVER>
  <filter **>
	@type parse_json_field
	json_fields  message
	merge_json_log false
	replace_json_log true
  </filter>
  <match **>
    @type copy
    <store>
    	@type remote_syslog
    	@id syslog_receiver
    	host sl.svc.messaging.cluster.local
    	port 9654
    	rfc rfc5424
    	facility user
        severity debug
    	protocol udp
    	packet_size 4096
		hostname "#{ENV['NODE_NAME']}"
        <buffer >
        @type file
        path '/var/lib/fluentd/syslog_receiver'
        flush_mode interval
        flush_interval 1s
        flush_thread_count 2
        flush_at_shutdown true
        retry_type exponential_backoff
        retry_wait 1s
        retry_max_interval 60s
        retry_forever true
        queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
        total_limit_size "#{ENV['TOTAL_LIMIT_SIZE'] ||  8589934592 }" #8G
        chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
        overflow_action block
      </buffer>
    </store>
  </match>
</label>`
		tcpWithTLSConf := `<label @SYSLOG_RECEIVER>
  <filter **>
	@type parse_json_field
	json_fields  message
	merge_json_log false
	replace_json_log true
  </filter>
  <match **>
    @type copy
    <store>
    	@type remote_syslog
    	@id syslog_receiver
    	host sl.svc.messaging.cluster.local
    	port 9654
    	rfc rfc5424
    	facility user
        severity debug
    	protocol tcp
    	packet_size 4096
		hostname "#{ENV['NODE_NAME']}"
        tls true
        ca_file '/var/run/ocp-collector/secrets/some-secret/ca-bundle.crt'
        verify_mode true
        timeout 60
        timeout_exception true
        keep_alive true
        keep_alive_idle 75
        keep_alive_cnt 9
        keep_alive_intvl 7200
        <buffer >
    		@type file
        path '/var/lib/fluentd/syslog_receiver'
        flush_mode interval
        flush_interval 1s
        flush_thread_count 2
        flush_at_shutdown true
        retry_type exponential_backoff
        retry_wait 1s
        retry_max_interval 60s
        retry_forever true
        queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
        total_limit_size "#{ENV['TOTAL_LIMIT_SIZE'] ||  8589934592 }" #8G
        chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
        overflow_action block
      </buffer>
    </store>
  </match>
</label>`
		udpWithTLSConf := `<label @SYSLOG_RECEIVER>
  <filter **>
	@type parse_json_field
	json_fields  message
	merge_json_log false
	replace_json_log true
  </filter>
  <match **>
    @type copy
    <store>
    	@type remote_syslog
    	@id syslog_receiver
    	host sl.svc.messaging.cluster.local
    	port 9654
    	rfc rfc5424
    	facility user
        severity debug
    	protocol udp
    	packet_size 4096
		hostname "#{ENV['NODE_NAME']}"
        tls true
        ca_file '/var/run/ocp-collector/secrets/some-secret/ca-bundle.crt'
        verify_mode true
        <buffer >
        @type file
        path '/var/lib/fluentd/syslog_receiver'
        flush_mode interval
        flush_interval 1s
        flush_thread_count 2
        flush_at_shutdown true
        retry_type exponential_backoff
        retry_wait 1s
        retry_max_interval 60s
        retry_forever true
        queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
        total_limit_size "#{ENV['TOTAL_LIMIT_SIZE'] ||  8589934592 }" #8G
        chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
        overflow_action block
      </buffer>
    </store>
  </match>
</label>`

		Context("for tcp endpoint", func() {
			Context("with TLS disabled", func() {
				BeforeEach(func() {
					outputs = []logging.OutputSpec{
						{
							Type: "syslog",
							Name: "syslog-receiver",
							URL:  "tcp://sl.svc.messaging.cluster.local:9654",
						},
					}
				})
				It("should produce well formed output label config", func() {
					results, err := generator.generateOutputLabelBlocks(outputs, nil, forwarderSpec)
					Expect(err).To(BeNil())
					Expect(len(results)).To(Equal(1))
					Expect(results[0]).To(EqualTrimLines(tcpConf))
				})
			})
			Context("with TLS enabled", func() {
				BeforeEach(func() {
					outputs = []logging.OutputSpec{
						{
							Type: "syslog",
							Name: "syslog-receiver",
							URL:  "tls://sl.svc.messaging.cluster.local:9654",
							Secret: &logging.OutputSecretSpec{
								Name: "some-secret",
							},
						},
					}
				})
				It("should produce well formed output label config", func() {
					results, err := generator.generateOutputLabelBlocks(outputs, nil, forwarderSpec)
					Expect(err).To(BeNil())
					Expect(len(results)).To(Equal(1))
					Expect(results[0]).To(EqualTrimLines(tcpWithTLSConf))
				})
			})
			Context("with AddLogSource flag", func() {
				syslogConfWithAddSource := `<label @SYSLOG_RECEIVER>
				  <filter **>
					@type parse_json_field
					json_fields  message
					merge_json_log false
					replace_json_log true
				  </filter>
				  <filter **>
					@type record_modifier
					<record>
					  kubernetes_info ${if record.has_key?('kubernetes'); record['kubernetes']; else {}; end}
					  namespace_info  ${if record['kubernetes_info'] != nil && record['kubernetes_info'] != {}; "namespace_name=" + record['kubernetes_info']['namespace_name']; else nil; end}
					  pod_info        ${if record['kubernetes_info'] != nil && record['kubernetes_info'] != {}; "pod_name=" + record['kubernetes_info']['pod_name']; else nil; end}
					  container_info  ${if record['kubernetes_info'] != nil && record['kubernetes_info'] != {}; "container_name=" + record['kubernetes_info']['container_name']; else nil; end}
					  msg_key         ${if record.has_key?('message') && record['message'] != nil; record['message']; else nil; end}
					  msg_info        ${if record['msg_key'] != nil && record['msg_key'].is_a?(Hash); require 'json'; "message="+record['message'].to_json; elsif record['msg_key'] != nil; "message="+record['message']; else nil; end}
					  message         ${if record['msg_key'] != nil && record['kubernetes_info'] != nil && record['kubernetes_info'] != {}; record['namespace_info'] + ", " + record['container_info'] + ", " + record['pod_info'] + ", " + record['msg_info']; else record['message']; end}
					</record>
					  remove_keys kubernetes_info, namespace_info, pod_info, container_info, msg_key, msg_info
				  </filter>
				  <match **>
					@type copy
					<store>
					  @type remote_syslog
					  @id syslog_receiver
					  host sl.svc.messaging.cluster.local
					  port 9654
					  rfc rfc5424
					  facility user
					  severity debug
					  protocol tcp
					  packet_size 4096
					  hostname "#{ENV['NODE_NAME']}"
					  tls true
					  ca_file '/var/run/ocp-collector/secrets/some-secret/ca-bundle.crt'
					  verify_mode true
					  timeout 60
					  timeout_exception true
					  keep_alive true
					  keep_alive_idle 75
					  keep_alive_cnt 9
					  keep_alive_intvl 7200
					  <buffer >
						@type file
						path '/var/lib/fluentd/syslog_receiver'
						flush_mode interval
						flush_interval 1s
						flush_thread_count 2
						flush_at_shutdown true
						retry_type exponential_backoff
						retry_wait 1s
						retry_max_interval 60s
						retry_forever true
						queued_chunks_limit_size "#{ENV['BUFFER_QUEUE_LIMIT'] || '32' }"
						total_limit_size "#{ENV['TOTAL_LIMIT_SIZE'] ||  8589934592 }" #8G
						chunk_limit_size "#{ENV['BUFFER_SIZE_LIMIT'] || '8m'}"
						overflow_action block
					  </buffer>
					</store>
				  </match>
				</label>`
				BeforeEach(func() {
					outputs = []logging.OutputSpec{
						{
							Type: "syslog",
							Name: "syslog-receiver",
							URL:  "tls://sl.svc.messaging.cluster.local:9654",
							Secret: &logging.OutputSecretSpec{
								Name: "some-secret",
							},
							OutputTypeSpec: logging.OutputTypeSpec{
								Syslog: &logging.Syslog{
									AddLogSource: true,
									RFC:          "RFC5424",
								},
							},
						},
					}
				})
				It("should produce config to copy log source information to log message", func() {
					results, err := generator.generateOutputLabelBlocks(outputs, nil, forwarderSpec)
					Expect(err).To(BeNil())
					Expect(len(results)).To(Equal(1))
					Expect(results[0]).To(EqualTrimLines(syslogConfWithAddSource))
				})
			})
		})

		Context("for udp endpoint", func() {
			Context("with TLS disabled", func() {
				BeforeEach(func() {
					outputs = []logging.OutputSpec{
						{
							Type: "syslog",
							Name: "syslog-receiver",
							URL:  "udp://sl.svc.messaging.cluster.local:9654",
						},
					}
				})
				It("should produce well formed output label config", func() {
					results, err := generator.generateOutputLabelBlocks(outputs, nil, forwarderSpec)
					Expect(err).To(BeNil())
					Expect(len(results)).To(Equal(1))
					Expect(results[0]).To(EqualTrimLines(udpConf))
				})
			})
			Context("with TLS enabled", func() {
				BeforeEach(func() {
					outputs = []logging.OutputSpec{
						{
							Type: "syslog",
							Name: "syslog-receiver",
							URL:  "udps://sl.svc.messaging.cluster.local:9654",
							Secret: &logging.OutputSecretSpec{
								Name: "some-secret",
							},
						},
					}
				})
				It("should produce well formed output label config", func() {
					results, err := generator.generateOutputLabelBlocks(outputs, nil, forwarderSpec)
					Expect(err).To(BeNil())
					Expect(len(results)).To(Equal(1))
					Expect(results[0]).To(EqualTrimLines(udpWithTLSConf))
				})
			})
		})
	})
})
