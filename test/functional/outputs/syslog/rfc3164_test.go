package syslog

import (
	"fmt"
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
				ToSyslogOutput(obs.SyslogRFC3164, func(spec *obs.OutputSpec) {
					spec.Syslog.Facility = `{.facility_key||"notfound"}`
					spec.Syslog.Severity = `{.severity_key||"notfound"}`
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
		})
	})

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

		expMatch := fmt.Sprintf(".*:\\s*%s", payload)
		if enrichment == obs.EnrichmentTypeKubernetesMinimal {
			expMatch = fmt.Sprintf(`namespace_name=.*, container_name=collector, pod_name=functional, message=%s`, payload)
		}
		Expect(logs[0]).To(MatchRegexp(expMatch), fmt.Sprintf("Exp. message source info to be added. EnrichmentType=%v", enrichment))
	},
		Entry("should enrich application logs with container source info", obs.InputTypeApplication, obs.EnrichmentTypeKubernetesMinimal),
		Entry("should do nothing additional to the application logs", obs.InputTypeApplication, obs.EnrichmentTypeNone),
	)
})
