package syslog

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/framework/e2e"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testfw "github.com/openshift/cluster-logging-operator/test/functional"
)

var _ = Describe("[Functional][Outputs][Syslog] RFC3164 tests", func() {
	const payload = `{"index":1,"timestamp":1}`

	var (
		framework          *functional.CollectorFunctionalFramework
		maxReadDuration, _ = time.ParseDuration("30s")
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFrameworkUsingCollector(testfw.LogCollectionType)
		framework.MaxReadDuration = &maxReadDuration
	})

	AfterEach(func() {
		framework.Cleanup()
	})
	DescribeTable("logforwarder configured with output.syslog.tag", func(tagSpec, expTag string, requiresFluentd bool) {
		if requiresFluentd && testfw.LogCollectionType != logging.LogCollectionTypeFluentd {
			Skip("Test requires fluentd")
		}
		functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication).
			ToOutputWithVisitor(func(spec *logging.OutputSpec) {
				spec.Syslog.Facility = "user"
				spec.Syslog.Severity = "debug"
				spec.Syslog.PayloadKey = "message"
				spec.Syslog.RFC = e2e.RFC3164.String()
				spec.Syslog.Tag = tagSpec
			}, logging.OutputTypeSyslog)
		Expect(framework.Deploy()).To(BeNil())

		record := `{"index": 1, "timestamp": 1, "tag_key": "rec_tag"}`
		crioMessage := functional.NewFullCRIOLogMessage(functional.CRIOTime(time.Now()), record)
		Expect(framework.WriteMessagesToApplicationLog(crioMessage, 1)).To(BeNil())

		outputlogs, err := framework.ReadRawApplicationLogsFrom(logging.OutputTypeSyslog)
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		Expect(outputlogs).To(HaveLen(1), "Expected the receiver to receive the message")
		expMatch := fmt.Sprintf(`( %s )`, expTag)
		Expect(outputlogs[0]).To(MatchRegexp(expMatch), "Exp to find tag in received message")
		Expect(outputlogs[0]).To(MatchRegexp(`{"index":.*1,.*"timestamp":.*1,.*"tag_key":.*"rec_tag"}`), "Exp to find the original message in received message")
	},

		Entry("should use the value from the record and include the message", "$.message.tag_key", "rec_tag", false),
		Entry("should use the value from the complete tag and include the message", "tag", `kubernetes\.var\.log.pods\..*`, true),
		Entry("should use values from parts of the tag and include the message", "${tag[0]}#${tag[-2]}", `kubernetes#.*`, true),
	)
	Describe("configured with values for facility,severity", func() {
		It("should use values from the record", func() {
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToOutputWithVisitor(func(spec *logging.OutputSpec) {
					spec.Syslog.Facility = "$.message.facility_key"
					spec.Syslog.Severity = "$.message.severity_key"
					spec.Syslog.RFC = e2e.RFC3164.String()
					spec.Syslog.Tag = "myTag"
				}, logging.OutputTypeSyslog)
			Expect(framework.Deploy()).To(BeNil())

			record := `{"index": 1, "timestamp": 1, "facility_key": "local0", "severity_key": "Informational"}`
			crioMessage := functional.NewFullCRIOLogMessage(functional.CRIOTime(time.Now()), record)
			Expect(framework.WriteMessagesToApplicationLog(crioMessage, 1)).To(BeNil())

			outputlogs, err := framework.ReadRawApplicationLogsFrom(logging.OutputTypeSyslog)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).To(HaveLen(1), "Expected the receiver to receive the message")

			// 134 = Facility(local0/16)*8 + Severity(Informational/6)
			// The 1 after <134> is version, which is always set to 1
			expectedPriority := "<134>1 "
			Expect(outputlogs[0]).To(MatchRegexp(expectedPriority), "Exp to find tag in received message")
		})
	})

	DescribeTable("configured to addLogSourceToMessage should add namespace, pod, container name", func(source string, requiresFluentd bool) {
		if requiresFluentd && testfw.LogCollectionType != logging.LogCollectionTypeFluentd {
			Skip("Test requires fluentd")
		}
		writeLogs := framework.WriteMessagesToInfraContainerLog
		readLogsFrom := framework.ReadInfrastructureLogsFrom
		if source == logging.InputNameApplication {
			writeLogs = framework.WriteMessagesToApplicationLog
			readLogsFrom = framework.ReadRawApplicationLogsFrom
		}

		functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(source).
			ToOutputWithVisitor(func(spec *logging.OutputSpec) {
				spec.Syslog.Facility = "user"
				spec.Syslog.Severity = "debug"
				spec.Syslog.PayloadKey = "message"
				spec.Syslog.RFC = e2e.RFC3164.String()
				spec.Syslog.Tag = "mytag"
				spec.Syslog.AddLogSource = true
			}, logging.OutputTypeSyslog)
		Expect(framework.Deploy()).To(BeNil())

		crioMessage := functional.NewFullCRIOLogMessage(functional.CRIOTime(time.Now()), payload)
		Expect(writeLogs(crioMessage, 1)).To(BeNil())

		outputlogs, err := readLogsFrom(logging.OutputTypeSyslog)
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		Expect(outputlogs).To(HaveLen(1), "Expected the receiver to receive the message")
		expMatch := fmt.Sprintf(`namespace_name=.*, container_name=collector, pod_name=functional, message=%s`, payload)
		Expect(outputlogs[0]).To(MatchRegexp(expMatch), "Exp. message source info to be added")
	},

		Entry("should support infrastructure logs", logging.InputNameInfrastructure, true),
		Entry("should support application logs", logging.InputNameApplication, false),
	)
})
