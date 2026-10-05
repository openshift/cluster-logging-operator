package drop

import (
	"fmt"
	"strings"
	"time"

	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"

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

			verifyLogs := func() bool {
				logs, err := f.ReadApplicationLogsFrom(string(obs.OutputTypeElasticsearch))
				if err != nil || len(logs) == 0 {
					return false
				}
				hasInfoMessage := false
				for _, msg := range logs {
					// Should not have dropped messages
					if msg.Message == "my error message" || msg.Message == "debug message" {
						return false
					}
					if msg.Message == "information message" {
						hasInfoMessage = true
					}
				}
				return hasInfoMessage
			}
			Eventually(verifyLogs, 2*time.Minute, 10*time.Second).Should(BeTrue(), "Expected to find 'information message' and not find dropped messages")
			Consistently(verifyLogs, 30*time.Second, 5*time.Second).Should(BeTrue(), "Dropped messages must remain absent after all writes")
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

			// Keep all writes below the reader's ten-result limit, including dropped records.
			Expect(f.WriteMessagesToOpenshiftAuditLog(makeLog(403), 2)).To(BeNil())
			Expect(f.WriteMessagesToOpenshiftAuditLog(makeLog(404), 2)).To(BeNil())
			Expect(f.WriteMessagesToOpenshiftAuditLog(makeLog(200), 2)).To(BeNil())

			verifyLogs := func() bool {
				logs, err := f.ReadAuditLogsFrom(string(obs.OutputTypeElasticsearch))
				if err != nil || len(logs) == 0 {
					return false
				}
				// Should have exactly two logs (both with code 403).
				if len(logs) != 2 {
					return false
				}
				var auditLogs []types.OpenshiftAuditLog
				err = types.StrictlyParseLogs(utils.ToJsonLogs(logs), &auditLogs)
				if err != nil {
					return false
				}
				// All logs should have responseStatus.code == 403
				for _, auditLog := range auditLogs {
					if auditLog.ResponseStatus.Code != 403 {
						return false
					}
				}
				return true
			}
			Eventually(verifyLogs, 2*time.Minute, 10*time.Second).Should(BeTrue(), "Expected exactly two audit logs with responseStatus.code=403")
			Consistently(verifyLogs, 30*time.Second, 5*time.Second).Should(BeTrue(), "Dropped audit records must remain absent after all writes")
		})

		DescribeTable("should apply olderThan and field conditions", func(inputType obs.InputType, useInfrastructureNamespace bool) {
			const olderThan = "2026-09-16T00:00:00Z"

			options := []client.TestOption{}
			if useInfrastructureNamespace {
				options = append(options, client.UseInfraNamespaceTestOption)
			}
			f = functional.NewCollectorFunctionalFramework(options...)
			testruntime.NewClusterLogForwarderBuilder(f.Forwarder).
				FromInput(inputType).
				WithFilter(dropFilterName, func(spec *obs.FilterSpec) {
					spec.Type = obs.FilterTypeDrop
					spec.DropTestsSpec = []obs.DropTest{
						{
							DropConditions: []obs.DropCondition{
								{OlderThan: olderThan},
								{Field: ".message", Matches: "drop-historical-"},
							},
						},
					}
				}).
				ToElasticSearchOutput()

			Expect(f.Deploy()).To(Succeed())

			cutoff, err := time.Parse(time.RFC3339, olderThan)
			Expect(err).NotTo(HaveOccurred())
			newCRIRecord := func(eventTime time.Time, message string) string {
				return functional.NewFullCRIOLogMessage(functional.CRIOTime(eventTime), message)
			}
			var (
				sourceName string
				write      func(string, int) error
				record     func(time.Time, string) string
				read       func() ([]string, error)
			)
			switch inputType {
			case obs.InputTypeApplication:
				sourceName, write, record = "application", f.WriteMessagesToApplicationLog, newCRIRecord
				read = func() ([]string, error) {
					logs, err := f.ReadApplicationLogsFrom(string(obs.OutputTypeElasticsearch))
					if err != nil {
						return nil, err
					}
					messages := make([]string, 0, len(logs))
					for _, log := range logs {
						messages = append(messages, log.Message)
					}
					return messages, nil
				}
			case obs.InputTypeInfrastructure:
				sourceName, write, record = "infrastructure-container", f.WriteMessagesToInfraContainerLog, newCRIRecord
				read = func() ([]string, error) {
					return f.ReadInfrastructureLogsFrom(string(obs.OutputTypeElasticsearch))
				}
			case obs.InputTypeAudit:
				sourceName, write = "auditd", f.WriteMessagesToAuditLog
				record = func(eventTime time.Time, message string) string {
					return functional.NewAuditHostLog(eventTime) + " " + message
				}
				read = func() ([]string, error) {
					return f.ReadAuditLogsFrom(string(obs.OutputTypeElasticsearch))
				}
			}

			// Each index receives only four records, even if the filter drops none.
			for _, entry := range []struct {
				eventTime time.Time
				message   string
			}{
				{cutoff.Add(-time.Second), "drop-historical-" + sourceName + "-older"},
				{cutoff, "drop-historical-" + sourceName + "-equal"},
				{cutoff.Add(-time.Second), "keep-historical-" + sourceName},
				// This retained record is last so the source must drain before absence is checked.
				{cutoff.Add(time.Second), "drop-historical-" + sourceName + "-newer"},
			} {
				Expect(write(record(entry.eventTime, entry.message), 1)).To(Succeed())
			}

			verifyRecords := func(records []string) error {
				received := strings.Join(records, "\n")
				for _, expected := range []string{
					"drop-historical-" + sourceName + "-equal",
					"drop-historical-" + sourceName + "-newer",
					"keep-historical-" + sourceName,
				} {
					if !strings.Contains(received, expected) {
						return fmt.Errorf("expected retained record %q in collector output", expected)
					}
				}
				dropped := "drop-historical-" + sourceName + "-older"
				if strings.Contains(received, dropped) {
					return fmt.Errorf("expected dropped record %q to be absent from collector output", dropped)
				}
				return nil
			}
			verifyLogs := func() error {
				logs, err := read()
				if err != nil {
					return err
				}
				return verifyRecords(logs)
			}
			Eventually(verifyLogs, 2*time.Minute, 10*time.Second).Should(Succeed())
			Consistently(verifyLogs, 30*time.Second, 5*time.Second).Should(Succeed(), "Dropped records must remain absent after the source's final record arrives")
		},
			Entry("application records", obs.InputTypeApplication, false),
			Entry("infrastructure container records", obs.InputTypeInfrastructure, true),
			Entry("auditd records", obs.InputTypeAudit, false),
		)

	})

})

func makeLog(code int) string {
	now := functional.CRIOTime(time.Now())
	auditLogLine := fmt.Sprintf(`{"kind":"Event","requestReceivedTimestamp":"%s","level":"Metadata", "responseStatus":{"code":%d}}`, now, code)
	return auditLogLine
}
