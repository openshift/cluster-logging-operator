package syslog

import (
	"fmt"
	"time"

	obstestruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"

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
		framework = functional.NewCollectorFunctionalFramework()
		framework.MaxReadDuration = &maxReadDuration
	})

	AfterEach(func() {
		framework.Cleanup()
	})

	// TODO: NEED TO FIX AS IT IS RELIANT ON TAG
	DescribeTable("logforwarder configured with output.syslog.tag", func(tagSpec, expTag string, requiresFluentd bool) {
		if requiresFluentd && testfw.LogCollectionType != logging.LogCollectionTypeFluentd {
			Skip("TODO: fix me for vector?Does that make sense? Test requires fluentd")
		}
		obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication).
			ToSyslogOutput(obs.SyslogRFC3164, func(output *obs.OutputSpec) {
				output.Syslog.Facility = "user"
				output.Syslog.Severity = "debug"
				output.Syslog.PayloadKey = "message"
			})
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
	// Entry("should use the value from the record and include the message", "$.message.tag_key", "rec_tag", false),
	)

	Describe("configured with values for facility,severity", func() {
		It("should use values from the record", func() {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToSyslogOutput(obs.SyslogRFC3164, func(spec *obs.OutputSpec) {
					spec.Syslog.Facility = "$.message.facility_key"
					spec.Syslog.Severity = "$.message.severity_key"
					spec.Syslog.RFC = obs.SyslogRFC3164
				})
			Expect(framework.Deploy()).To(BeNil())

			record := `{"index": 1, "timestamp": 1, "facility_key": "local0", "severity_key": "Informational"}`
			crioMessage := functional.NewFullCRIOLogMessage(functional.CRIOTime(time.Now()), record)
			Expect(framework.WriteMessagesToApplicationLog(crioMessage, 1)).To(BeNil())

			outputlogs, err := framework.ReadRawApplicationLogsFrom(string(obs.OutputTypeSyslog))
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(outputlogs).To(HaveLen(1), "Expected the receiver to receive the message")

			// 134 = Facility(local0/16)*8 + Severity(Informational/6)
			// The 1 after <134> is version, which is always set to 1
			expectedPriority := "<134>1 "
			Expect(outputlogs[0]).To(MatchRegexp(expectedPriority), "Exp to find tag in received message")
		})
	})

	// TODO: FIX AFTER ADDLOGSOURCE FINALIZED
	DescribeTable("configured to addLogSourceToMessage should add namespace, pod, container name", func(source obs.InputType) {
		writeLogs := framework.WriteMessagesToInfraContainerLog
		readLogsFrom := framework.ReadInfrastructureLogsFrom
		if source == obs.InputTypeApplication {
			writeLogs = framework.WriteMessagesToApplicationLog
			readLogsFrom = framework.ReadRawApplicationLogsFrom
		}

		obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(source).
			ToSyslogOutput(obs.SyslogRFC3164, func(spec *obs.OutputSpec) {
				spec.Syslog.Facility = "user"
				spec.Syslog.Severity = "debug"
				spec.Syslog.PayloadKey = "message"
			})
		Expect(framework.Deploy()).To(BeNil())

		crioMessage := functional.NewFullCRIOLogMessage(functional.CRIOTime(time.Now()), payload)
		Expect(writeLogs(crioMessage, 1)).To(BeNil())

		outputlogs, err := readLogsFrom(string(obs.OutputTypeSyslog))
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		Expect(outputlogs).To(HaveLen(1), "Expected the receiver to receive the message")
		expMatch := fmt.Sprintf(`namespace_name=.*, container_name=collector, pod_name=functional, message=%s`, payload)
		Expect(outputlogs[0]).To(MatchRegexp(expMatch), "Exp. message source info to be added")
	},
	// Entry("should support application logs", obs.InputTypeApplication, false),
	)
})
