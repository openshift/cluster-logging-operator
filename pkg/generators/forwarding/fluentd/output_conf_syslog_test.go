package fluentd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	test "github.com/openshift/cluster-logging-operator/test"
)

var _ = Describe("Generating external syslog server output store config blocks", func() {

	var (
		err       error
		outputs   []logging.OutputSpec
		generator *ConfigGenerator
	)
	BeforeEach(func() {
		generator, err = NewConfigGenerator(false)
		Expect(err).To(BeNil())
	})

	Context("based on legacy syslog plugin", func() {
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
						Type:     logging.OutputTypeLegacySyslog,
						Name:     "syslog-receiver",
						Endpoint: "sl.svc.messaging.cluster.local:9654",
					},
				}
			})
			It("should produce well formed output label config", func() {
				results, err := generator.generateOutputLabelBlocks(outputs)
				Expect(err).To(BeNil())
				Expect(len(results)).To(Equal(1))
				test.Expect(results[0]).ToEqual(tcpConf)
			})
		})

		Context("for tcp endpoint", func() {
			BeforeEach(func() {
				outputs = []logging.OutputSpec{
					{
						Type:     logging.OutputTypeLegacySyslog,
						Name:     "syslog-receiver",
						Endpoint: "tcp://sl.svc.messaging.cluster.local:9654",
					},
				}
			})
			It("should produce well formed output label config", func() {
				results, err := generator.generateOutputLabelBlocks(outputs)
				Expect(err).To(BeNil())
				Expect(len(results)).To(Equal(1))
				test.Expect(results[0]).ToEqual(tcpConf)
			})
		})

		Context("for udp endpoint", func() {
			BeforeEach(func() {
				outputs = []logging.OutputSpec{
					{
						Type:     logging.OutputTypeLegacySyslog,
						Name:     "syslog-receiver",
						Endpoint: "udp://sl.svc.messaging.cluster.local:9654",
					},
				}
			})
			It("should produce well formed output label config", func() {
				results, err := generator.generateOutputLabelBlocks(outputs)
				Expect(err).To(BeNil())
				Expect(len(results)).To(Equal(1))
				test.Expect(results[0]).ToEqual(udpConf)
			})
		})
	})

	Context("based on the new syslog plugin", func() {
		secureConf := `<label @SYSLOG_RECEIVER>
        <match **>
           @type copy
           <store>
             @type remote_syslog
             @id syslog_receiver
             host sl.svc.messaging.cluster.local
             protocol tcp
             tls true
             ca_file '/var/run/ocp-collector/secrets/my-syslog-secret/ca.pem'
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
             @type remote_syslog
             @id syslog_receiver
             host sl.svc.messaging.cluster.local
             protocol udp
             port 9654
             hostname ${hostname}
             facility user
             severity debug
           </store>
        </match>
</label>`
		tcpConf := `<label @SYSLOG_RECEIVER>
        <match **>
           @type copy
           <store>
             @type remote_syslog
             @id syslog_receiver
             host sl.svc.messaging.cluster.local
             protocol tcp
             port 9654
             hostname ${hostname}
             facility user
             severity debug
           </store>
        </match>
</label>`

		Context("for an secure endpoint", func() {
			BeforeEach(func() {
				outputs = []logging.OutputSpec{
					{
						Type:     logging.OutputTypeSyslog,
						Name:     "syslog-receiver",
						Endpoint: "sl.svc.messaging.cluster.local:9654",
						Secret: &logging.OutputSecretSpec{
							Name: "my-syslog-secret",
						},
					},
				}
			})
			It("should produce well formed output label config", func() {
				results, err := generator.generateOutputLabelBlocks(outputs)
				Expect(err).To(BeNil())
				Expect(len(results)).To(Equal(1))
				test.Expect(results[0]).ToEqual(secureConf)
			})
		})

		Context("for a udp endpoint", func() {
			BeforeEach(func() {
				outputs = []logging.OutputSpec{
					{
						Type:     logging.OutputTypeSyslog,
						Name:     "syslog-receiver",
						Endpoint: "udp://sl.svc.messaging.cluster.local:9654",
					},
				}
			})
			It("should produce well formed output label config", func() {
				results, err := generator.generateOutputLabelBlocks(outputs)
				Expect(err).To(BeNil())
				Expect(len(results)).To(Equal(1))
				test.Expect(results[0]).ToEqual(udpConf)
			})
		})

		Context("for a tcp endpoint", func() {
			BeforeEach(func() {
				outputs = []logging.OutputSpec{
					{
						Type:     logging.OutputTypeSyslog,
						Name:     "syslog-receiver",
						Endpoint: "tcp://sl.svc.messaging.cluster.local:9654",
					},
				}
			})
			It("should produce well formed output label config", func() {
				results, err := generator.generateOutputLabelBlocks(outputs)
				Expect(err).To(BeNil())
				Expect(len(results)).To(Equal(1))
				test.Expect(results[0]).ToEqual(tcpConf)
			})
		})

		Context("for a protocol-less endpoint", func() {
			BeforeEach(func() {
				outputs = []logging.OutputSpec{
					{
						Type:     logging.OutputTypeSyslog,
						Name:     "syslog-receiver",
						Endpoint: "sl.svc.messaging.cluster.local:9654",
					},
				}
			})
			It("should produce well formed output label config", func() {
				results, err := generator.generateOutputLabelBlocks(outputs)
				Expect(err).To(BeNil())
				Expect(len(results)).To(Equal(1))
				test.Expect(results[0]).ToEqual(tcpConf)
			})
		})
	})
})
