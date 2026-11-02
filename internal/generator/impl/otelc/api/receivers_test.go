package api_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	api "github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/receivers"
	"gopkg.in/yaml.v3"
)

var _ = Describe("Receivers", func() {
	Context("FileLog Receiver", func() {
		It("should marshal to YAML correctly", func() {
			receiver := receivers.NewFileLog("", "/var/log/*.log")
			receiver.StartAt = "beginning"
			receiver.Encoding = "utf-8"
			receiver.MaxConcurrentFiles = 512
			receiver.Attributes = map[string]interface{}{
				"log.type": "application",
			}

			data, err := yaml.Marshal(receiver)
			Expect(err).To(BeNil())
			Expect(string(data)).To(ContainSubstring("include:"))
			Expect(string(data)).To(ContainSubstring("/var/log/*.log"))
			Expect(string(data)).To(ContainSubstring("start_at: beginning"))
			Expect(string(data)).To(ContainSubstring("encoding: utf-8"))
			Expect(string(data)).To(ContainSubstring("max_concurrent_files: 512"))
		})

		It("should unmarshal from YAML correctly", func() {
			yamlData := `
include:
  - /var/log/app/*.log
  - /var/log/service/*.log
exclude:
  - /var/log/app/*.debug
start_at: beginning
encoding: utf-8
max_concurrent_files: 256
attributes:
  log.type: application
  environment: production
multiline:
  line_start_pattern: "^\\d{4}-\\d{2}-\\d{2}"
operators:
  - type: json_parser
    id: json_parse
  - type: regex_parser
    id: regex_parse
    regex: "^(?P<timestamp>\\d+) (?P<level>\\w+)"
`
			var receiver receivers.FileLog
			err := yaml.Unmarshal([]byte(yamlData), &receiver)
			Expect(err).To(BeNil())
			Expect(receiver.Include).To(HaveLen(2))
			Expect(receiver.Include[0]).To(Equal("/var/log/app/*.log"))
			Expect(receiver.Exclude).To(HaveLen(1))
			Expect(receiver.StartAt).To(Equal("beginning"))
			Expect(receiver.Encoding).To(Equal("utf-8"))
			Expect(receiver.MaxConcurrentFiles).To(Equal(256))
			Expect(receiver.Attributes).To(HaveKeyWithValue("log.type", "application"))
			Expect(receiver.Multiline).ToNot(BeNil())
			Expect(receiver.Multiline.LineStartPattern).To(Equal("^\\d{4}-\\d{2}-\\d{2}"))
			Expect(receiver.Operators).To(HaveLen(2))
		})
	})

	Context("Receivers Map", func() {
		It("should unmarshal multiple receivers from YAML", func() {
			yamlData := `
filelog/app:
  include:
    - /var/log/app/*.log
  start_at: beginning
filelog/system:
  include:
    - /var/log/system/*.log
  start_at: end
  max_concurrent_files: 100
`
			var receiversMap api.Receivers
			err := yaml.Unmarshal([]byte(yamlData), &receiversMap)
			Expect(err).To(BeNil())
			Expect(receiversMap).To(HaveLen(2))

			appReceiver, ok := receiversMap["filelog/app"]
			Expect(ok).To(BeTrue())
			appFileLog, ok := appReceiver.(*receivers.FileLog)
			Expect(ok).To(BeTrue())
			Expect(appFileLog.Include).To(Equal([]string{"/var/log/app/*.log"}))
			Expect(appFileLog.StartAt).To(Equal("beginning"))

			sysReceiver, ok := receiversMap["filelog/system"]
			Expect(ok).To(BeTrue())
			sysFileLog, ok := sysReceiver.(*receivers.FileLog)
			Expect(ok).To(BeTrue())
			Expect(sysFileLog.Include).To(Equal([]string{"/var/log/system/*.log"}))
			Expect(sysFileLog.StartAt).To(Equal("end"))
			Expect(sysFileLog.MaxConcurrentFiles).To(Equal(100))
		})
	})
})
