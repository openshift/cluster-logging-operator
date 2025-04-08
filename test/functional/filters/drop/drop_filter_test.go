package drop

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

var _ = Describe("[Functional][Filters][Drop] Drop filter", func() {
	const (
		dropFilterName = "myDrop"
	)

	var (
		f *functional.CollectorFunctionalFramework
	)

	AfterEach(func() {
		f.Cleanup()
	})

	Describe("when drop filter is spec'd", func() {
		It("should drop logs that have `error` in its message OR logs with messages that doesn't include `information` AND includes `debug`", func() {
			f = functional.NewCollectorFunctionalFramework()

			testruntime.NewClusterLogForwarderBuilder(f.Forwarder).
				FromInput(obs.InputTypeApplication).
				WithFilter(dropFilterName, func(spec *obs.FilterSpec) {
					spec.Type = obs.FilterTypeDrop
					spec.DropTestsSpec = []obs.DropTest{
						{
							DropConditions: []obs.DropCondition{
								{
									Field:   ".message",
									Matches: "error",
								},
							},
						},
						{
							DropConditions: []obs.DropCondition{
								{
									Field:      ".message",
									NotMatches: "information",
								},
								{
									Field:   ".message",
									Matches: "debug",
								},
							},
						},
					}
				}).
				ToElasticSearchOutput()

			Expect(f.Deploy()).To(BeNil())
			msg := functional.NewFullCRIOLogMessage(functional.CRIOTime(time.Now()), "my error message")
			Expect(f.WriteMessagesToApplicationLog(msg, 1)).To(BeNil())
			msg2 := functional.NewFullCRIOLogMessage(functional.CRIOTime(time.Now()), "information message")
			Expect(f.WriteMessagesToApplicationLog(msg2, 1)).To(BeNil())
			msg3 := functional.NewFullCRIOLogMessage(functional.CRIOTime(time.Now()), "debug message")
			Expect(f.WriteMessagesToApplicationLog(msg3, 1)).To(BeNil())
			Expect(f.WritesApplicationLogs(5)).To(Succeed())

			logs, err := f.ReadApplicationLogsFrom(string(obs.OutputTypeElasticsearch))
			Expect(err).To(BeNil(), "Error fetching logs from %s: %v", obs.OutputTypeElasticsearch, err)
			Expect(logs).To(Not(BeEmpty()), "Exp. logs to be forwarded to %s", obs.OutputTypeElasticsearch)
			hasInfoMessage := false
			for _, msg := range logs {
				Expect(msg.ViaQCommon.Message).ToNot(Equal("my error message"))
				Expect(msg.ViaQCommon.Message).ToNot(Equal("debug message"))
				if msg.ViaQCommon.Message == "information message" {
					hasInfoMessage = true
				}
			}
			Expect(hasInfoMessage).To(BeTrue())
		})

		It("should drop logs that have `.responseStatus.code` not equals 403", func() {
			f = functional.NewCollectorFunctionalFramework()

			testruntime.NewClusterLogForwarderBuilder(f.Forwarder).
				FromInput(obs.InputTypeAudit).
				WithFilter(dropFilterName, func(spec *obs.FilterSpec) {
					spec.Type = obs.FilterTypeDrop
					spec.DropTestsSpec = []obs.DropTest{
						{
							DropConditions: []obs.DropCondition{
								{
									Field:      ".responseStatus.code",
									NotMatches: "403",
								},
							},
						},
					}
				}).
				ToElasticSearchOutput()

			Expect(f.Deploy()).To(BeNil())

			Expect(f.WriteMessagesToOpenshiftAuditLog(makeLog(403), 10)).To(BeNil())
			Expect(f.WriteMessagesToOpenshiftAuditLog(makeLog(404), 10)).To(BeNil())
			Expect(f.WriteMessagesToOpenshiftAuditLog(makeLog(200), 10)).To(BeNil())

			logs, err := f.ReadAuditLogsFrom(string(obs.OutputTypeElasticsearch))
			Expect(err).To(BeNil(), "Error fetching logs from %s: %v", obs.OutputTypeElasticsearch, err)
			Expect(logs).To(Not(BeEmpty()), "Exp. logs to be forwarded to %s", obs.OutputTypeElasticsearch)
			Expect(len(logs)).To(BeEquivalentTo(10))
			var auditLogs []types.OpenshiftAuditLog
			err = types.StrictlyParseLogs(utils.ToJsonLogs(logs), &auditLogs)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs: %v", logs)
			for _, auditLog := range auditLogs {
				Expect(auditLog.ResponseStatus.Code).To(Equal(403))
			}
		})

	})

})

func makeLog(code int) string {
	now := functional.CRIOTime(time.Now())
	auditLogLine := fmt.Sprintf(`{"kind":"Event","requestReceivedTimestamp":"%s","level":"Metadata", "responseStatus":{"code":%d}}`, now, code)
	return auditLogLine
}
