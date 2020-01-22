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
		generator, err = NewConfigGenerator(false, false)
		Expect(err).To(BeNil())
	})

	Context("based on syslog plugin", func() {
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

		Context("for tcp endpoint", func() {
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

		Context("for udp endpoint", func() {
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
	})
})
