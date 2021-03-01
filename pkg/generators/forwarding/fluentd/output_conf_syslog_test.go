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
