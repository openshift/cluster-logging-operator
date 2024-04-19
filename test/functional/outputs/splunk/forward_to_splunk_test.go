package splunk

import (
	"fmt"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"

	v1 "k8s.io/api/core/v1"
)

var _ = Describe("Forwarding to Splunk", func() {
	var (
		framework *functional.CollectorFunctionalFramework
		secret    *v1.Secret
	)
	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFrameworkUsingCollector(logging.LogCollectionTypeVector)
		secret = runtime.NewSecret(framework.Namespace, "splunk-secret",
			map[string][]byte{
				"hecToken": []byte(string(functional.HecToken)),
			},
		)
	})

	AfterEach(func() {
		framework.Cleanup()
	})

	It("should accept application logs", func() {

		testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication).
			ToSplunkOutput()
		framework.Secrets = append(framework.Secrets, secret)
		Expect(framework.Deploy()).To(BeNil())

		// Wait for splunk to be ready
		time.Sleep(90 * time.Second)

		// Write app logs
		timestamp := "2020-11-04T18:13:59.061892+00:00"
		applicationLogLine := functional.NewCRIOLogMessage(timestamp, "This is my test message", false)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 2)).To(BeNil())

		// Read app logs
		logs, err := framework.ReadLogsByTypeFromSplunk(framework.Namespace, framework.Name, logging.InputNameApplication)
		Expect(err).To(BeNil(), "Expected no errors getting logs from splunk")
		Expect(logs).ToNot(BeEmpty())

		// Parse the logs
		var appLogs []types.ApplicationLog
		jsonString := fmt.Sprintf("[%s]", strings.Join(logs, ","))
		err = types.ParseLogsFrom(jsonString, &appLogs, false)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")

		outputTestLog := appLogs[0]
		Expect(outputTestLog.LogType).To(Equal(logging.InputNameApplication))
	})

	It("should accept audit logs without timestamp unexpected type warning (see: https://issues.redhat.com/browse/LOG-4672)", func() {
		testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameAudit).
			ToSplunkOutput()
		framework.Secrets = append(framework.Secrets, secret)
		Expect(framework.Deploy()).To(BeNil())

		// Wait for splunk to be ready
		time.Sleep(90 * time.Second)

		// Write audit logs
		timestamp, _ := time.Parse(time.RFC3339Nano, "2024-04-16T09:46:19.116+00:00")
		auditLogLine := functional.NewAuditHostLog(timestamp)
		writeAuditLogs := framework.WriteMessagesToAuditLog(auditLogLine, 1)
		Expect(writeAuditLogs).To(BeNil(), "Expect no errors writing audit logs")

		// Read audit logs
		logs, err := framework.ReadLogsByTypeFromSplunk(framework.Namespace, framework.Name, logging.InputNameAudit)
		Expect(err).To(BeNil(), "Expected no errors getting logs from splunk")
		Expect(logs).ToNot(BeEmpty())

		// Parse the logs
		var auditLogs []types.AuditLog
		jsonString := fmt.Sprintf("[%s]", strings.Join(logs, ","))
		err = types.ParseLogsFrom(jsonString, &auditLogs, false)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		Expect(len(auditLogs)).To(Equal(1), "Expected one audit log")

		Expect(auditLogs[0].LogType).To(Equal(logging.InputNameAudit), "Expected audit log type")
		Expect(auditLogs[0].Level).To(Equal("default"), "Expected audit log level to default")

		collectorLog, err := framework.ReadCollectorLogs()
		Expect(err).To(BeNil(), "Expected no errors reading the collector logs")
		Expect(collectorLog).ToNot(BeEmpty(), "Expected collector logs to not be empty")
		tsWarn := "Timestamp was an unexpected type. Deferring to Splunk to set the timestamp"
		Expect(strings.Contains(collectorLog, tsWarn)).To(BeFalse(), "Expected collector logs to NOT contain timestamp unexpected type warning")
	})

	Context("with custom indexes", func() {
		withIndexName := func(spec *logging.OutputSpec) {
			spec.Splunk = &logging.Splunk{
				IndexName: functional.SplunkIndexName,
			}
		}

		withIndexKey := func(spec *logging.OutputSpec) {
			spec.Splunk = &logging.Splunk{
				IndexKey: "log_type",
			}
		}

		withFakeIndexKey := func(spec *logging.OutputSpec) {
			spec.Splunk = &logging.Splunk{
				IndexKey: "kubernetes.foo_key",
			}
		}

		It("should send logs to spec'd indexName in Splunk", func() {

			testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToOutputWithVisitor(withIndexName, logging.OutputTypeSplunk)
			framework.Secrets = append(framework.Secrets, secret)
			Expect(framework.Deploy()).To(BeNil())

			// Wait for splunk to be ready
			time.Sleep(90 * time.Second)

			// Write app logs
			timestamp := "2020-11-04T18:13:59.061892+00:00"
			applicationLogLine := functional.NewCRIOLogMessage(timestamp, "This is my test message", false)
			Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 2)).To(BeNil())

			// Read app logs
			logs, err := framework.ReadAppLogsByIndexFromSplunk(framework.Namespace, framework.Name, functional.SplunkIndexName)
			Expect(err).To(BeNil(), "Expected no errors getting logs from splunk")
			Expect(logs).ToNot(BeEmpty())

			// Parse the logs
			var appLogs []types.ApplicationLog
			jsonString := fmt.Sprintf("[%s]", strings.Join(logs, ","))
			err = types.ParseLogsFrom(jsonString, &appLogs, false)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")

			outputTestLog := appLogs[0]
			Expect(outputTestLog.LogType).To(Equal(logging.InputNameApplication))
		})

		It("should send logs to spec'd indexKey in Splunk", func() {

			testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToOutputWithVisitor(withIndexKey, logging.OutputTypeSplunk)
			framework.Secrets = append(framework.Secrets, secret)
			Expect(framework.Deploy()).To(BeNil())

			// Wait for splunk to be ready
			time.Sleep(90 * time.Second)

			// Write app logs
			timestamp := "2020-11-04T18:13:59.061892+00:00"
			applicationLogLine := functional.NewCRIOLogMessage(timestamp, "This is my test message", false)
			Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 2)).To(BeNil())

			// Read app logs from index = "application"
			logs, err := framework.ReadAppLogsByIndexFromSplunk(framework.Namespace, framework.Name, logging.InputNameApplication)
			Expect(err).To(BeNil(), "Expected no errors getting logs from splunk")
			Expect(logs).ToNot(BeEmpty())

			// Parse the logs
			var appLogs []types.ApplicationLog
			jsonString := fmt.Sprintf("[%s]", strings.Join(logs, ","))
			err = types.ParseLogsFrom(jsonString, &appLogs, false)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")

			outputTestLog := appLogs[0]
			Expect(outputTestLog.LogType).To(Equal(logging.InputNameApplication))
		})

		It("should send logs to default index if spec'd indexKey is not available", func() {

			testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToOutputWithVisitor(withFakeIndexKey, logging.OutputTypeSplunk)
			framework.Secrets = append(framework.Secrets, secret)
			Expect(framework.Deploy()).To(BeNil())

			// Wait for splunk to be ready
			time.Sleep(90 * time.Second)

			// Write app logs
			timestamp := "2020-11-04T18:13:59.061892+00:00"
			applicationLogLine := functional.NewCRIOLogMessage(timestamp, "This is my test message", false)
			Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 2)).To(BeNil())

			// Read app logs from default index in splunk. Without config, default is "main"
			logs, err := framework.ReadAppLogsByIndexFromSplunk(framework.Namespace, framework.Name, functional.SplunkDefaultIndex)
			Expect(err).To(BeNil(), "Expected no errors getting logs from splunk")
			Expect(logs).ToNot(BeEmpty())

			// Parse the logs
			var appLogs []types.ApplicationLog
			jsonString := fmt.Sprintf("[%s]", strings.Join(logs, ","))
			err = types.ParseLogsFrom(jsonString, &appLogs, false)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")

			outputTestLog := appLogs[0]
			Expect(outputTestLog.LogType).To(Equal(logging.InputNameApplication))
		})
	})

	Context("tuning parameters", func() {
		var (
			compVisitFunc func(spec *logging.OutputSpec)
		)

		DescribeTable("with compression settings", func(compression string) {

			compVisitFunc = func(spec *logging.OutputSpec) {
				spec.Tuning = &logging.OutputTuningSpec{
					Compression: compression,
				}
			}

			testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToOutputWithVisitor(compVisitFunc, logging.OutputTypeSplunk)
			framework.Secrets = append(framework.Secrets, secret)
			Expect(framework.Deploy()).To(BeNil())

			// Wait for splunk to be ready
			time.Sleep(90 * time.Second)

			// Write app logs
			timestamp := "2020-11-04T18:13:59.061892+00:00"
			applicationLogLine := functional.NewCRIOLogMessage(timestamp, "This is my test message", false)
			Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 2)).To(BeNil())

			// Read app logs
			logs, err := framework.ReadLogsByTypeFromSplunk(framework.Namespace, framework.Name, logging.InputNameApplication)

			Expect(err).To(BeNil(), "expected no errors getting logs from splunk")
			Expect(logs).ToNot(BeEmpty())

			// Parse the logs
			var appLogs []types.ApplicationLog
			jsonString := fmt.Sprintf("[%s]", strings.Join(logs, ","))
			err = types.ParseLogsFrom(jsonString, &appLogs, false)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")
			outputTestLog := appLogs[0]
			Expect(outputTestLog.LogType).To(Equal(logging.InputNameApplication))

		},
			Entry("should pass with gzip", "gzip"),
			Entry("should pass with no compression", ""),
		)
	})
})
