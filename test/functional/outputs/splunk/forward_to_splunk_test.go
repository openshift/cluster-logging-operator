package splunk

import (
	"fmt"
	"time"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/helpers/splunk"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	obstestruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
	"k8s.io/apimachinery/pkg/api/resource"

	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	v1 "k8s.io/api/core/v1"
)

const SplunkSecretName = "splunk-secret"

var _ = Describe("Forwarding to Splunk", func() {
	var (
		framework    *functional.CollectorFunctionalFramework
		secret       *v1.Secret
		hecSecretKey = *internalobs.NewSecretReference(constants.SplunkHECTokenKey, SplunkSecretName)
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		secret = runtime.NewSecret(framework.Namespace, SplunkSecretName,
			map[string][]byte{
				"hecToken": functional.HecToken,
			},
		)
	})

	AfterEach(func() {
		framework.Cleanup()
	})

	It("should accept application logs", func() {
		obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication).
			ToSplunkOutput(hecSecretKey, func(output *obs.OutputSpec) {
				output.Splunk.Index = "main"
			})
		framework.Secrets = append(framework.Secrets, secret)
		Expect(framework.Deploy()).To(BeNil())

		// Wait for splunk to be ready
		splunk.WaitOnSplunk(framework)

		// Write app logs
		timestamp := "2020-11-04T18:13:59.061892+00:00"
		applicationLogLine := functional.NewCRIOLogMessage(timestamp, "This is my test message", false)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 2)).To(BeNil())

		// Read app logs
		logs, err := framework.ReadLogsByTypeFromSplunk(string(obs.InputTypeApplication))
		Expect(err).To(BeNil(), "Expected no errors getting logs from splunk")
		Expect(logs).ToNot(BeEmpty())

		// Parse the logs
		var appLogs []types.ApplicationLog
		jsonString := fmt.Sprintf("[%s]", strings.Join(logs, ","))
		err = types.ParseLogsFrom(jsonString, &appLogs, false)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")

		outputTestLog := appLogs[0]
		Expect(outputTestLog.LogType).To(Equal(string(obs.InputTypeApplication)))
	})

	It("should not send application logs more than 64Ki", func() {
		obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication, func(spec *obs.InputSpec) {
				spec.Name = "small-msg-app"
				spec.Application.Tuning = &obs.ContainerInputTuningSpec{
					MaxMessageSize: utils.GetPtr(resource.MustParse("64Ki")),
				}
			}).
			ToSplunkOutput(hecSecretKey, func(output *obs.OutputSpec) {
				output.Splunk.Index = "main"
			})
		framework.Secrets = append(framework.Secrets, secret)
		Expect(framework.Deploy()).To(BeNil())

		// Wait for splunk to be ready
		splunk.WaitOnSplunk(framework)

		// Write app logs
		timestamp := "2020-11-04T18:13:59.061892+00:00"
		applicationLogLine := functional.NewCRIOLogMessage(timestamp, "This is my test message", false)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 1)).To(BeNil())

		// Write large app logs
		Expect(framework.WriteApplicationLogOfSizeAsPartials(65 * 1024)).To(BeNil())

		//one more normal app log
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 1)).To(BeNil())

		// Parse the logs
		var appLogs []types.ApplicationLog
		logs, err := framework.ReadLogsByTypeFromSplunk(string(obs.InputTypeApplication))
		Expect(err).To(BeNil())
		jsonString := fmt.Sprintf("[%s]", strings.Join(logs, ","))
		err = types.ParseLogsFrom(jsonString, &appLogs, false)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")

		Expect(appLogs).To(HaveLen(2))
		collectorLog, err := framework.ReadCollectorLogs()
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		Expect(collectorLog).To(
			And(
				ContainSubstring("Found line that exceeds max_merged_line_bytes; discarding."),
				ContainSubstring("configured_limit=65536"),
			),
		)
	})

	It("should accept audit logs without timestamp unexpected type warning (see: https://issues.redhat.com/browse/LOG-4672)", func() {
		obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeAudit).
			ToSplunkOutput(hecSecretKey, func(output *obs.OutputSpec) {
				output.Splunk.Index = "main"
			})
		framework.Secrets = append(framework.Secrets, secret)
		Expect(framework.Deploy()).To(BeNil())

		// Wait for splunk to be ready
		splunk.WaitOnSplunk(framework)

		// Write audit logs
		timestamp, _ := time.Parse(time.RFC3339Nano, "2024-04-16T09:46:19.116+00:00")
		auditLogLine := functional.NewAuditHostLog(timestamp)
		writeAuditLogs := framework.WriteMessagesToAuditLog(auditLogLine, 1)
		Expect(writeAuditLogs).To(BeNil(), "Expect no errors writing audit logs")

		// Read audit logs
		logs, err := framework.ReadLogsByTypeFromSplunk(string(obs.InputTypeAudit))
		Expect(err).To(BeNil(), "Expected no errors getting logs from splunk")
		Expect(logs).ToNot(BeEmpty())

		// Parse the logs
		var auditLogs []types.AuditLog
		jsonString := fmt.Sprintf("[%s]", strings.Join(logs, ","))
		err = types.ParseLogsFrom(jsonString, &auditLogs, false)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		Expect(len(auditLogs)).To(Equal(1), "Expected one audit log")

		Expect(auditLogs[0].LogType).To(Equal(string(obs.InputTypeAudit)), "Expected audit log type")
		Expect(auditLogs[0].Level).To(Equal("default"), "Expected audit log level to default")

		collectorLog, err := framework.ReadCollectorLogs()
		Expect(err).To(BeNil(), "Expected no errors reading the collector logs")
		Expect(collectorLog).ToNot(BeEmpty(), "Expected collector logs to not be empty")
		tsWarn := "Timestamp was an unexpected type. Deferring to Splunk to set the timestamp"
		Expect(strings.Contains(collectorLog, tsWarn)).To(BeFalse(), "Expected collector logs to NOT contain timestamp unexpected type warning")
	})

	Context("splunk index", func() {
		DescribeTable("with user defined indices", func(index, expIndex string) {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToSplunkOutput(hecSecretKey, func(output *obs.OutputSpec) {
					if index != "" {
						output.Splunk.Index = index
					}
				})
			framework.Secrets = append(framework.Secrets, secret)
			framework.Labels["slash/test.dot"] = "application"
			Expect(framework.Deploy()).To(BeNil())

			// Wait for splunk to be ready
			splunk.WaitOnSplunk(framework)

			// Write app logs
			timestamp := "2020-11-04T18:13:59.061892+00:00"
			applicationLogLine := functional.NewCRIOLogMessage(timestamp, "This is my test message", false)
			Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 2)).To(BeNil())

			// Read app logs
			logs, err := framework.ReadAppLogsByIndexFromSplunk(expIndex)
			Expect(err).To(BeNil(), "Expected no errors getting logs from splunk")
			Expect(logs).ToNot(BeEmpty())

			// Parse the logs
			var appLogs []types.ApplicationLog
			jsonString := fmt.Sprintf("[%s]", strings.Join(logs, ","))
			err = types.ParseLogsFrom(jsonString, &appLogs, false)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")

			outputTestLog := appLogs[0]
			Expect(outputTestLog.LogType).To(Equal(string(obs.InputTypeApplication)))
		},
			Entry("should send logs to spec'd static index in Splunk", functional.SplunkIndexName, functional.SplunkIndexName),
			Entry("should send logs to spec'd dynamic index in Splunk", `{.log_type||"missing"}`, "application"),
			Entry("should send logs to spec'd static + dynamic index in Splunk", `foo-{.log_type||"missing"}`, "foo-application"),
			Entry("should send logs to spec'd static + label with dot/slash index in Splunk", `foo-{.kubernetes.labels."slash/test.dot"||"missing"}`, "foo-application"),
			Entry("should send logs to spec'd static + fallback value's index in Splunk if field is missing", `foo-{.missing||"application"}`, "foo-application"),
			Entry("should send logs to default index in Splunk when no index is defined", "", functional.SplunkDefaultIndex))
	})

	Context("tuning parameters", func() {
		DescribeTable("with compression settings", func(compression string) {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToSplunkOutput(hecSecretKey, func(output *obs.OutputSpec) {
					output.Splunk.Tuning = &obs.SplunkTuningSpec{
						Compression: compression,
					}
					output.Splunk.Index = "main"
				})
			framework.Secrets = append(framework.Secrets, secret)
			Expect(framework.Deploy()).To(BeNil())

			// Wait for splunk to be ready
			splunk.WaitOnSplunk(framework)

			// Write app logs
			timestamp := "2020-11-04T18:13:59.061892+00:00"
			applicationLogLine := functional.NewCRIOLogMessage(timestamp, "This is my test message", false)
			Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 2)).To(BeNil())

			// Read app logs
			logs, err := framework.ReadLogsByTypeFromSplunk(string(obs.InputTypeApplication))

			Expect(err).To(BeNil(), "expected no errors getting logs from splunk")
			Expect(logs).ToNot(BeEmpty())

			// Parse the logs
			var appLogs []types.ApplicationLog
			jsonString := fmt.Sprintf("[%s]", strings.Join(logs, ","))
			err = types.ParseLogsFrom(jsonString, &appLogs, false)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")
			outputTestLog := appLogs[0]
			Expect(outputTestLog.LogType).To(Equal(string(obs.InputTypeApplication)))

		},
			Entry("should pass with gzip", "gzip"),
			Entry("should pass with no compression", ""),
		)
	})

})
