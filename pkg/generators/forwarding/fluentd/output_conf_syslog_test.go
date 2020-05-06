package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/test"
)

var _ = Describe("Generating external syslog server output store config blocks", func() {

	var (
		err       error
		outputs   []logging.OutputSpec
		generator *ConfigGenerator
	)
	Context("based on old syslog plugin", func() {

		BeforeEach(func() {
			generator, err = NewConfigGenerator(false, false, true)
			Expect(err).To(BeNil())
		})

		tcpConf := `<label @SYSLOG_RECEIVER>
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
						URL:  "sl.svc.messaging.cluster.local:9654",
					},
				}
			})
			It("should produce well formed output label config", func() {
				results, err := generator.generateOutputLabelBlocks(outputs)
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
				results, err := generator.generateOutputLabelBlocks(outputs)
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
				results, err := generator.generateOutputLabelBlocks(outputs)
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
      program fluentd
      protocol tcp
      packet_size 4096
      timeout 60
      timeout_exception true
      keep_alive true
      keep_alive_idle 75
      keep_alive_cnt 9
      keep_alive_intvl 7200
      <buffer>
        @type file
        path '/var/lib/fluentd/syslog_receiver'
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
  </match>
</label>`
		udpConf := `<label @SYSLOG_RECEIVER>
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
      program fluentd
      protocol udp
      packet_size 4096
      <buffer>
        @type file
        path '/var/lib/fluentd/syslog_receiver'
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
  </match>
</label>`
		tcpWithTLSConf := `<label @SYSLOG_RECEIVER>
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
      program fluentd
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
      <buffer>
        @type file
        path '/var/lib/fluentd/syslog_receiver'
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
  </match>
</label>`
		udpWithTLSConf := `<label @SYSLOG_RECEIVER>
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
      program fluentd
      protocol udp
      packet_size 4096
      tls true
      ca_file '/var/run/ocp-collector/secrets/some-secret/ca-bundle.crt'
      verify_mode true
      <buffer>
        @type file
        path '/var/lib/fluentd/syslog_receiver'
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
					results, err := generator.generateOutputLabelBlocks(outputs)
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
							URL:  "tcp://sl.svc.messaging.cluster.local:9654",
							Secret: &logging.OutputSecretSpec{
								Name: "some-secret",
							},
						},
					}
				})
				It("should produce well formed output label config", func() {
					results, err := generator.generateOutputLabelBlocks(outputs)
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
					results, err := generator.generateOutputLabelBlocks(outputs)
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
							URL:  "udp://sl.svc.messaging.cluster.local:9654",
							Secret: &logging.OutputSecretSpec{
								Name: "some-secret",
							},
						},
					}
				})
				It("should produce well formed output label config", func() {
					results, err := generator.generateOutputLabelBlocks(outputs)
					Expect(err).To(BeNil())
					Expect(len(results)).To(Equal(1))
					Expect(results[0]).To(EqualTrimLines(udpWithTLSConf))
				})
			})
		})
	})
})
