package syslog

import (
	"testing"

	"github.com/openshift/cluster-logging-operator/internal/generator"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Generating external syslog server output store config blocks", func() {

	var (
		outputs []logging.OutputSpec
		g       generator.Generator
		op      generator.Options
		secret  *corev1.Secret
	)
	Context("based on old syslog plugin", func() {

		BeforeEach(func() {
			g = generator.MakeGenerator()
			op = generator.Options{
				generator.UseOldRemoteSyslogPlugin: "",
			}
		})

		tcpConf := `
<label @SYSLOG_RECEIVER>
  <filter **>
    @type parse_json_field
    json_fields  message
    merge_json_log false
    replace_json_log true
  </filter>
  
  <match **>
    @type syslog_buffered
    @id syslog_receiver
    remote_syslog sl.svc.messaging.cluster.local
    port 9654
    hostname "#{ENV['NODE_NAME']}"
    facility user
    severity debug
  </match>
</label>
`

		udpConf := `
<label @SYSLOG_RECEIVER>
  <filter **>
    @type parse_json_field
    json_fields  message
    merge_json_log false
    replace_json_log true
  </filter>
  
  <match **>
    @type syslog
    @id syslog_receiver
    remote_syslog sl.svc.messaging.cluster.local
    port 9654
    hostname "#{ENV['NODE_NAME']}"
    facility user
    severity debug
  </match>
</label>
`

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
				c := Conf(nil, secret, outputs[0], op)
				results, err := g.GenerateConf(c...)
				Expect(err).To(BeNil())
				Expect(results).To(EqualTrimLines(tcpConf))
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
				c := Conf(nil, nil, outputs[0], op)
				results, err := g.GenerateConf(c...)
				Expect(err).To(BeNil())
				Expect(results).To(EqualTrimLines(tcpConf))
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
				c := Conf(nil, nil, outputs[0], op)
				results, err := g.GenerateConf(c...)
				Expect(err).To(BeNil())
				Expect(results).To(EqualTrimLines(udpConf))
			})
		})
	})

	Context("based on new syslog plugin", func() {
		BeforeEach(func() {
			g = generator.MakeGenerator()
			secret = nil
		})
		tcpConf := `
<label @SYSLOG_RECEIVER>
  <filter **>
    @type parse_json_field
    json_fields  message
    merge_json_log false
    replace_json_log true
  </filter>
  
  <match **>
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
    <format>
      @type json
    </format>
    <buffer>
      @type file
      path '/var/lib/fluentd/syslog_receiver'
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
`
		udpConf := `
<label @SYSLOG_RECEIVER>
  <filter **>
    @type parse_json_field
    json_fields  message
    merge_json_log false
    replace_json_log true
  </filter>
  
  <match **>
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
    <format>
      @type json
    </format>
    <buffer>
      @type file
      path '/var/lib/fluentd/syslog_receiver'
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
`
		//nolint:gosec
		tcpWithTLSConf := `
<label @SYSLOG_RECEIVER>
  <filter **>
    @type parse_json_field
    json_fields  message
    merge_json_log false
    replace_json_log true
  </filter>
  
  <match **>
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
    client_cert_key '/var/run/ocp-collector/secrets/some-secret/tls.key'
    client_cert '/var/run/ocp-collector/secrets/some-secret/tls.crt'
    ca_file '/var/run/ocp-collector/secrets/some-secret/ca-bundle.crt'
    client_cert_key_password "#{File.exists?('/var/run/ocp-collector/secrets/some-secret/passphrase') ? open('/var/run/ocp-collector/secrets/some-secret/passphrase','r') do |f|f.read end : ''}"
    timeout 60
    timeout_exception true
    keep_alive true
    keep_alive_idle 75
    keep_alive_cnt 9
    keep_alive_intvl 7200
    <format>
      @type json
    </format>
    <buffer>
      @type file
      path '/var/lib/fluentd/syslog_receiver'
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
`
		//nolint:gosec
		udpWithTLSConf := `
<label @SYSLOG_RECEIVER>
  <filter **>
    @type parse_json_field
    json_fields  message
    merge_json_log false
    replace_json_log true
  </filter>
  
  <match **>
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
    <format>
      @type json
    </format>
    <buffer>
      @type file
      path '/var/lib/fluentd/syslog_receiver'
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
`

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
					c := Conf(nil, secret, outputs[0], nil)
					results, err := g.GenerateConf(c...)
					Expect(err).To(BeNil())
					Expect(results).To(EqualTrimLines(tcpConf))
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
					secret = &corev1.Secret{
						Data: map[string][]byte{
							"tls.crt":       []byte("my-tls"),
							"tls.key":       []byte("my-tls-key"),
							"ca-bundle.crt": []byte("my-bundle"),
							"passphrase":    []byte("my-tls-passphrase"),
						},
					}
				})
				It("should produce well formed output label config", func() {
					c := Conf(nil, secret, outputs[0], nil)
					results, err := g.GenerateConf(c...)
					Expect(err).To(BeNil())
					Expect(results).To(EqualTrimLines(tcpWithTLSConf))
				})
			})
			Context("with TLS enabled and Hostname Verify", func() {
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
					secret = &corev1.Secret{
						Data: map[string][]byte{
							"ca-bundle.crt":          []byte("junk"),
							"syslog_hostname_verify": []byte("true"),
						},
					}
				})
				It("should produce security config with host name verification", func() {
					c := SecurityConfig(outputs[0], secret)
					results, err := g.GenerateConf(c...)
					Expect(err).To(BeNil())
					Expect(results).To(EqualTrimLines(`
						tls true
						verify_mode 1 #VERIFY_NONE:0, VERIFY_PEER:1
						ca_file '/var/run/ocp-collector/secrets/some-secret/ca-bundle.crt'
					`))
				})
				It("should produce security config with host name verification disabled", func() {
					secret.Data[SyslogHostnameVerify] = []byte("false")
					c := SecurityConfig(outputs[0], secret)
					results, err := g.GenerateConf(c...)
					Expect(err).To(BeNil())
					Expect(results).To(EqualTrimLines(`
						tls true
						verify_mode 0 #VERIFY_NONE:0, VERIFY_PEER:1
						ca_file '/var/run/ocp-collector/secrets/some-secret/ca-bundle.crt'
					`))
				})
			})
			Context("with AddLogSource flag", func() {
				syslogConfWithAddSource := `
<label @SYSLOG_RECEIVER>
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
      namespace_info ${if record['kubernetes_info'] != nil && record['kubernetes_info'] != {}; "namespace_name=" + record['kubernetes_info']['namespace_name']; else nil; end}
      pod_info ${if record['kubernetes_info'] != nil && record['kubernetes_info'] != {}; "pod_name=" + record['kubernetes_info']['pod_name']; else nil; end}
      container_info ${if record['kubernetes_info'] != nil && record['kubernetes_info'] != {}; "container_name=" + record['kubernetes_info']['container_name']; else nil; end}
      msg_key ${if record.has_key?('message') && record['message'] != nil; record['message']; else nil; end}
      msg_info ${if record['msg_key'] != nil && record['msg_key'].is_a?(Hash); require 'json'; "message="+record['message'].to_json; elsif record['msg_key'] != nil; "message="+record['message']; else nil; end}
      message ${if record['msg_key'] != nil && record['kubernetes_info'] != nil && record['kubernetes_info'] != {}; record['namespace_info'] + ", " + record['container_info'] + ", " + record['pod_info'] + ", " + record['msg_info']; else record['message']; end}
      systemd_info ${if record.has_key?('systemd') && record['systemd']['t'].has_key?('PID'); record['systemd']['u']['SYSLOG_IDENTIFIER'] += "[" + record['systemd']['t']['PID'] + "]"; else {}; end}
    </record>
    remove_keys kubernetes_info, namespace_info, pod_info, container_info, msg_key, msg_info, systemd_info
  </filter>
  
  <match journal.** system.var.log**>
    @type remote_syslog
    @id syslog_receiver_journal
    host sl.svc.messaging.cluster.local
    port 9654
    rfc rfc5424
    facility user
    severity debug
    appname ${$.systemd.u.SYSLOG_IDENTIFIER}
    protocol tcp
    packet_size 4096
    hostname "#{ENV['NODE_NAME']}"
    tls true
    ca_file '/var/run/ocp-collector/secrets/some-secret/ca-bundle.crt'
    timeout 60
    timeout_exception true
    keep_alive true
    keep_alive_idle 75
    keep_alive_cnt 9
    keep_alive_intvl 7200
    <format>
      @type json
    </format>
    <buffer $.systemd.u.SYSLOG_IDENTIFIER>
      @type file
      path '/var/lib/fluentd/syslog_receiver_journal'
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
    timeout 60
    timeout_exception true
    keep_alive true
    keep_alive_idle 75
    keep_alive_cnt 9
    keep_alive_intvl 7200
    <format>
      @type json
    </format>
    <buffer>
      @type file
      path '/var/lib/fluentd/syslog_receiver'
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
`
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
					secret = &corev1.Secret{
						Data: map[string][]byte{
							"ca-bundle.crt": []byte("junk"),
						},
					}
				})
				It("should produce config to copy log source information to log message", func() {
					c := Conf(nil, secret, outputs[0], nil)
					results, err := g.GenerateConf(c...)
					Expect(err).To(BeNil())
					Expect(results).To(EqualTrimLines(syslogConfWithAddSource))
				})
			})
			Context("with AddLogSource flag and AppName field", func() {
				syslogConfWithAddSource := `
<label @SYSLOG_RECEIVER>
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
      namespace_info ${if record['kubernetes_info'] != nil && record['kubernetes_info'] != {}; "namespace_name=" + record['kubernetes_info']['namespace_name']; else nil; end}
      pod_info ${if record['kubernetes_info'] != nil && record['kubernetes_info'] != {}; "pod_name=" + record['kubernetes_info']['pod_name']; else nil; end}
      container_info ${if record['kubernetes_info'] != nil && record['kubernetes_info'] != {}; "container_name=" + record['kubernetes_info']['container_name']; else nil; end}
      msg_key ${if record.has_key?('message') && record['message'] != nil; record['message']; else nil; end}
      msg_info ${if record['msg_key'] != nil && record['msg_key'].is_a?(Hash); require 'json'; "message="+record['message'].to_json; elsif record['msg_key'] != nil; "message="+record['message']; else nil; end}
      message ${if record['msg_key'] != nil && record['kubernetes_info'] != nil && record['kubernetes_info'] != {}; record['namespace_info'] + ", " + record['container_info'] + ", " + record['pod_info'] + ", " + record['msg_info']; else record['message']; end}
      systemd_info ${if record.has_key?('systemd') && record['systemd']['t'].has_key?('PID'); record['systemd']['u']['SYSLOG_IDENTIFIER'] += "[" + record['systemd']['t']['PID'] + "]"; else {}; end}
    </record>
    remove_keys kubernetes_info, namespace_info, pod_info, container_info, msg_key, msg_info, systemd_info
  </filter>
  
  <match journal.** system.var.log**>
    @type remote_syslog
    @id syslog_receiver_journal
    host sl.svc.messaging.cluster.local
    port 9654
    rfc rfc5424
    facility user
    severity debug
    appname ${$.systemd.u.SYSLOG_IDENTIFIER}
    protocol tcp
    packet_size 4096
    hostname "#{ENV['NODE_NAME']}"
    tls true
    ca_file '/var/run/ocp-collector/secrets/some-secret/ca-bundle.crt'
    timeout 60
    timeout_exception true
    keep_alive true
    keep_alive_idle 75
    keep_alive_cnt 9
    keep_alive_intvl 7200
    <format>
      @type json
    </format>
    <buffer $.systemd.u.SYSLOG_IDENTIFIER>
      @type file
      path '/var/lib/fluentd/syslog_receiver_journal'
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
    @type remote_syslog
    @id syslog_receiver
    host sl.svc.messaging.cluster.local
    port 9654
    rfc rfc5424
    facility user
    severity debug
    appname app
    protocol tcp
    packet_size 4096
    hostname "#{ENV['NODE_NAME']}"
    tls true
    ca_file '/var/run/ocp-collector/secrets/some-secret/ca-bundle.crt'
    timeout 60
    timeout_exception true
    keep_alive true
    keep_alive_idle 75
    keep_alive_cnt 9
    keep_alive_intvl 7200
    <format>
      @type json
    </format>
    <buffer>
      @type file
      path '/var/lib/fluentd/syslog_receiver'
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
`
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
									AppName:      "app",
									RFC:          "RFC5424",
								},
							},
						},
					}
					secret = &corev1.Secret{
						Data: map[string][]byte{
							"ca-bundle.crt": []byte("junk"),
						},
					}
				})
				It("should produce config to copy log source information to log message", func() {
					c := Conf(nil, secret, outputs[0], nil)
					results, err := g.GenerateConf(c...)
					Expect(err).To(BeNil())
					Expect(results).To(EqualTrimLines(syslogConfWithAddSource))
				})
			})
			Context("with AddLogSource flag and rfc3164 flag", func() {
				syslogConfWithAddSource := `
<label @SYSLOG_RECEIVER>
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
      namespace_info ${if record['kubernetes_info'] != nil && record['kubernetes_info'] != {}; "namespace_name=" + record['kubernetes_info']['namespace_name']; else nil; end}
      pod_info ${if record['kubernetes_info'] != nil && record['kubernetes_info'] != {}; "pod_name=" + record['kubernetes_info']['pod_name']; else nil; end}
      container_info ${if record['kubernetes_info'] != nil && record['kubernetes_info'] != {}; "container_name=" + record['kubernetes_info']['container_name']; else nil; end}
      msg_key ${if record.has_key?('message') && record['message'] != nil; record['message']; else nil; end}
      msg_info ${if record['msg_key'] != nil && record['msg_key'].is_a?(Hash); require 'json'; "message="+record['message'].to_json; elsif record['msg_key'] != nil; "message="+record['message']; else nil; end}
      message ${if record['msg_key'] != nil && record['kubernetes_info'] != nil && record['kubernetes_info'] != {}; record['namespace_info'] + ", " + record['container_info'] + ", " + record['pod_info'] + ", " + record['msg_info']; else record['message']; end}
      systemd_info ${if record.has_key?('systemd') && record['systemd']['t'].has_key?('PID'); record['systemd']['u']['SYSLOG_IDENTIFIER'] += "[" + record['systemd']['t']['PID'] + "]"; else {}; end}
    </record>
    remove_keys kubernetes_info, namespace_info, pod_info, container_info, msg_key, msg_info, systemd_info
  </filter>
  
  <match journal.** system.var.log**>
    @type remote_syslog
    @id syslog_receiver_journal
    host sl.svc.messaging.cluster.local
    port 9654
    rfc rfc3164
    facility user
    severity debug
    program ${$.systemd.u.SYSLOG_IDENTIFIER}
    protocol tcp
    packet_size 4096
    hostname "#{ENV['NODE_NAME']}"
    tls true
    ca_file '/var/run/ocp-collector/secrets/some-secret/ca-bundle.crt'
    timeout 60
    timeout_exception true
    keep_alive true
    keep_alive_idle 75
    keep_alive_cnt 9
    keep_alive_intvl 7200
    <format>
      @type json
    </format>
    <buffer $.systemd.u.SYSLOG_IDENTIFIER>
      @type file
      path '/var/lib/fluentd/syslog_receiver_journal'
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
    @type remote_syslog
    @id syslog_receiver
    host sl.svc.messaging.cluster.local
    port 9654
    rfc rfc3164
    facility user
    severity debug
    protocol tcp
    packet_size 4096
    hostname "#{ENV['NODE_NAME']}"
    tls true
    ca_file '/var/run/ocp-collector/secrets/some-secret/ca-bundle.crt'
    timeout 60
    timeout_exception true
    keep_alive true
    keep_alive_idle 75
    keep_alive_cnt 9
    keep_alive_intvl 7200
    <format>
      @type json
    </format>
    <buffer>
      @type file
      path '/var/lib/fluentd/syslog_receiver'
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
`
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
									RFC:          "RFC3164",
								},
							},
						},
					}
					secret = &corev1.Secret{
						Data: map[string][]byte{
							"ca-bundle.crt": []byte("junk"),
						},
					}
				})
				It("should produce config to copy log source information to log message", func() {
					c := Conf(nil, secret, outputs[0], nil)
					results, err := g.GenerateConf(c...)
					Expect(err).To(BeNil())
					Expect(results).To(EqualTrimLines(syslogConfWithAddSource))
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
					c := Conf(nil, nil, outputs[0], nil)
					results, err := g.GenerateConf(c...)
					Expect(err).To(BeNil())
					Expect(results).To(EqualTrimLines(udpConf))
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
					secret = &corev1.Secret{
						Data: map[string][]byte{
							"ca-bundle.crt": []byte("junk"),
						},
					}
				})
				It("should produce well formed output label config", func() {
					c := Conf(nil, secret, outputs[0], nil)
					results, err := g.GenerateConf(c...)
					Expect(err).To(BeNil())
					Expect(results).To(EqualTrimLines(udpWithTLSConf))
				})
			})
		})
	})
})

func TestFluendConfGenerator(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fluend Conf Generation")
}
