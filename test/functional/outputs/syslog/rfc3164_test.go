package syslog

import (
	"strings"
	"time"

	obstestruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"

	"github.com/openshift/cluster-logging-operator/test/framework/functional"
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

	Describe("configured with values for facility,severity", func() {
		It("should use values from the record", func() {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				WithParseJson().
				ToSyslogOutput(obs.SyslogRFC3164, func(spec *obs.OutputSpec) {
					spec.Syslog.Facility = `{.structured.facility_key||"notfound"}`
					spec.Syslog.Severity = `{.structured.severity_key||"notfound"}`
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
			expectedPriority := "<134>"
			Expect(outputlogs[0]).To(MatchRegexp(expectedPriority), "Exp to find tag in received message")

			collectorLogs, err := framework.ReadCollectorLogs()
			Expect(err).To(BeNil())
			Expect(collectorLogs).ToNot(ContainSubstring("VRL compilation warning"), "Expected no VRL compilation warnings in collector logs")
		})
	})

	DescribeTable("should correctly encode severity keywords in all forms", func(severity string, expectedPriority string) {
		obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication).
			ToSyslogOutput(obs.SyslogRFC3164, func(spec *obs.OutputSpec) {
				spec.Syslog.Facility = "user"
				spec.Syslog.Severity = severity
			})
		Expect(framework.Deploy()).To(BeNil())

		crioMessage := functional.NewFullCRIOLogMessage(functional.CRIOTime(time.Now()), `{"index":1}`)
		Expect(framework.WriteMessagesToApplicationLog(crioMessage, 1)).To(BeNil())

		outputlogs, err := framework.ReadRawApplicationLogsFrom(string(obs.OutputTypeSyslog))
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		Expect(outputlogs).To(HaveLen(1), "Expected the receiver to receive the message")
		Expect(outputlogs[0]).To(HavePrefix(expectedPriority), "Expected priority to match facility(user/1)*8 + severity")

		collectorLogs, err := framework.ReadCollectorLogs()
		Expect(err).To(BeNil())
		Expect(collectorLogs).ToNot(ContainSubstring("VRL compilation warning"), "Expected no VRL compilation warnings in collector logs")
		Expect(collectorLogs).ToNot(ContainSubstring("Invalid syslog severity"), "Expected no invalid severity warnings in collector logs")
	},
		// facility "user" = 1, priority = 1*8 + severity_code
		// full-form keywords (RFC 5424)
		Entry("full-form: critical", "critical", "<10>"),         // 8 + 2
		Entry("full-form: emergency", "emergency", "<8>"),        // 8 + 0
		Entry("full-form: error", "error", "<11>"),               // 8 + 3
		Entry("full-form: informational", "informational", "<14>"), // 8 + 6
		Entry("full-form: warning", "warning", "<12>"),           // 8 + 4
		// capitalized keywords (as documented in oc explain)
		Entry("capitalized: Critical", "Critical", "<10>"),         // 8 + 2
		Entry("capitalized: Emergency", "Emergency", "<8>"),        // 8 + 0
		Entry("capitalized: Informational", "Informational", "<14>"), // 8 + 6
	)

	DescribeTable("should enrich logs based upon the enrichment type", func(source obs.InputType, enrichment obs.EnrichmentType) {
		obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(source).
			ToSyslogOutput(obs.SyslogRFC3164, func(spec *obs.OutputSpec) {
				spec.Syslog.Facility = "user"
				spec.Syslog.Severity = "debug"
				spec.Syslog.PayloadKey = "{.message}"
				spec.Syslog.Enrichment = enrichment
			})
		Expect(framework.Deploy()).To(BeNil())

		crioMessage := functional.NewFullCRIOLogMessage(functional.CRIOTime(time.Now()), payload)
		Expect(framework.WriteMessagesToApplicationLog(crioMessage, 1)).To(BeNil())

		logs, err := framework.ReadRawApplicationLogsFrom(string(obs.OutputTypeSyslog))
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		Expect(logs).To(HaveLen(1), "Expected the receiver to receive the message")

		logLine := strings.TrimSpace(logs[0])
		Expect(logLine).To(MatchRegexp(`^<\d+>`))
		Expect(logLine).To(MatchRegexp(`functionalcollector:`))

		if enrichment == obs.EnrichmentTypeKubernetesMinimal {
			Expect(logLine).To(MatchRegexp(`namespace_name=[^,]+`))
			Expect(logLine).To(ContainSubstring(`container_name=collector`))
			Expect(logLine).To(ContainSubstring(`pod_name=functional`))
			Expect(logLine).To(ContainSubstring("message=" + payload))
		} else {
			Expect(logLine).To(ContainSubstring(payload))
		}
	},
		Entry("should enrich application logs with container source info", obs.InputTypeApplication, obs.EnrichmentTypeKubernetesMinimal),
		Entry("should do nothing additional to the application logs", obs.InputTypeApplication, obs.EnrichmentTypeNone),
	)
})
