package otlp

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/otlp"
	obstestruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
)

const (
	timestamp     = "2023-08-28T12:59:28.573159188+00:00"
	timestampNano = "1693227568573159188"
)

var _ = Describe("[Functional][Outputs][OTLP] Functional tests", func() {
	var (
		framework    *functional.CollectorFunctionalFramework
		appNamespace string
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()

	})

	AfterEach(func() {
		framework.Cleanup()
	})

	Context("When forwarding logs to an otel-collector", func() {
		It("should send application logs with correct grouping and resource attributes", func() {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToOtlpOutput()

			Expect(framework.DeployWithVisitor(func(b *runtime.PodBuilder) error {
				return framework.AddOTELCollector(b, string(obs.OutputTypeOTLP))
			})).To(BeNil())

			appNamespace = framework.Pod.Namespace

			// Write message to namespace
			crioLine := functional.NewCRIOLogMessage(timestamp, "Format me to a c`ontainer message!", false)
			Expect(framework.WriteMessagesToNamespace(crioLine, appNamespace, 1)).To(Succeed())
			crioLine = functional.NewCRIOLogMessage(timestamp, "My second message.", false)
			Expect(framework.WriteMessagesToNamespace(crioLine, appNamespace, 1)).To(Succeed())
			crioLine = functional.NewCRIOLogMessage(timestamp, "My third and final message.", false)
			Expect(framework.WriteMessagesToNamespace(crioLine, appNamespace, 1)).To(Succeed())

			// Read logs
			raw, err := framework.ReadRawApplicationLogsFrom(string(obs.OutputTypeOTLP))
			Expect(err).To(BeNil(), "Expected no errors reading the logs for type")
			Expect(raw).ToNot(BeEmpty())

			logs, err := otlp.ParseLogs(raw[0])
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")

			// Access the first (and only) ResourceLog
			// Inspect resource attributes
			resourceLog := logs.ResourceLogs[0]
			resource := resourceLog.Resource
			namespaceName := resource.FindStringValue(resource.NamespaceNameAttribute())
			Expect(namespaceName).To(Equal(appNamespace), "Expect namespace name to exist in resource attributes")

			logSource := resource.FindStringValue(resource.LogSourceAttribute())
			Expect(logSource).To(Equal("container"))

			clusterId := resource.FindStringValue(resource.ClusterIDAttribute())
			Expect(clusterId).ToNot(BeEmpty(), "Expected cluster.id resource attribute")

			scopeLogs := resourceLog.ScopeLogs
			Expect(scopeLogs).ToNot(BeEmpty(), "Expected scopeLogs")
			Expect(scopeLogs).To(HaveLen(1), "Expected a single scopeLog")

			logRecords := scopeLogs[0].LogRecords
			Expect(logRecords).ToNot(BeEmpty(), "Expected log records for the scope")
			Expect(logRecords).To(HaveLen(3), "Expected same count as sent for log records")

			// Inspect the first log record for correct fields
			logRecord := logRecords[0]
			Expect(logRecord.TimeUnixNano).To(Equal(timestampNano), "Expect timestamp to be converted into unix nano")
			Expect(logRecord.SeverityNumber).To(Equal(9), "Expect severityNumber to not be parsed to 9")
			Expect(logRecord.Body.StringValue).ToNot(BeEmpty(), "Expect message to be populated")

			// Ensure we have log.type populated
			Expect(logRecord.Attributes).ToNot(BeEmpty(), "Expect logRecord attributes to be populated")
			logType := logRecord.FindStringValue(logRecord.LogTypeAttribute())
			Expect(logType).To(Equal(string(obs.InputTypeApplication)))

		})
		It("should send audit logs with correct grouping and resource attributes", func() {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeAudit).
				ToOtlpOutput()

			Expect(framework.DeployWithVisitor(func(b *runtime.PodBuilder) error {
				return framework.AddOTELCollector(b, string(obs.OutputTypeOTLP))
			})).To(BeNil())

			// Log message data
			nanoTime, _ := time.Parse(time.RFC3339Nano, timestamp)
			k8sAuditLogLine := fmt.Sprintf(`{"kind":"Event","requestReceivedTimestamp":"%s","level":"Metadata"}`, functional.CRIOTime(nanoTime))
			Expect(framework.WriteMessagesTok8sAuditLog(k8sAuditLogLine, 3)).To(BeNil())

			// Read line from Log Forward output
			raw, err := framework.ReadAuditLogsFrom(string(obs.OutputTypeOTLP))
			Expect(err).To(BeNil(), "Expected no errors reading the logs for type")
			Expect(raw).ToNot(BeEmpty())

			logs, err := otlp.ParseLogs(raw[0])
			resourceLog := logs.ResourceLogs[0]
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")

			resource := resourceLog.Resource
			logSource := resource.FindStringValue(resource.LogSourceAttribute())
			Expect(logSource).To(Equal("kubeAPI"))

			clusterId := resource.FindStringValue(resource.ClusterIDAttribute())
			Expect(clusterId).ToNot(BeEmpty(), "Expected cluster.id resource attribute")

			scopeLogs := resourceLog.ScopeLogs
			Expect(scopeLogs).ToNot(BeEmpty(), "Expected scopeLogs")
			Expect(scopeLogs).To(HaveLen(1), "Expected a single scopeLog")

			logRecords := scopeLogs[0].LogRecords
			Expect(logRecords).ToNot(BeEmpty(), "Expected log records for the scope")
			Expect(logRecords).To(HaveLen(3), "Expected same count as sent for log records")

			// Inspect the first log record for correct fields
			logRecord := logRecords[0]
			//Expect(logRecord.TimeUnixNano).To(Equal(timestampNano), "Expect timestamp to be converted into unix nano")
			Expect(logRecord.TimeUnixNano).ToNot(BeEmpty(), "Expect timestamp to be converted into unix nano")
			Expect(logRecord.Body.StringValue).ToNot(BeEmpty(), "Expect message to be populated")

			// Ensure we have log.type populated
			Expect(logRecord.Attributes).ToNot(BeEmpty(), "Expect logRecord attributes to be populated")
			logType := logRecord.FindStringValue(logRecord.LogTypeAttribute())
			Expect(logType).To(Equal(string(obs.InputTypeAudit)))
		})
		It("should send journal logs with correct resource attributes", func() {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeInfrastructure).
				ToOtlpOutput()

			Expect(framework.DeployWithVisitor(func(b *runtime.PodBuilder) error {
				return framework.AddOTELCollector(b, string(obs.OutputTypeOTLP))
			})).To(BeNil())

			// Log message data
			logline := functional.NewJournalLog(3, "*", "*")
			Expect(framework.WriteMessagesToInfraJournalLog(logline, 1)).To(BeNil())

			// Read line from Log Forward output
			raw, err := framework.ReadInfrastructureLogsFrom(string(obs.OutputTypeOTLP))
			Expect(err).To(BeNil(), "Expected no errors reading the logs for type")
			Expect(raw).ToNot(BeEmpty())

			logs, err := otlp.ParseLogs(raw[0])
			resourceLog := logs.ResourceLogs[0]
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")

			resource := resourceLog.Resource
			logSource := resource.FindStringValue(resource.LogSourceAttribute())
			Expect(logSource).To(Equal("node"))

			clusterId := resource.FindStringValue(resource.ClusterIDAttribute())
			Expect(clusterId).ToNot(BeEmpty(), "Expected cluster.id resource attribute")

			scopeLogs := resourceLog.ScopeLogs
			Expect(scopeLogs).ToNot(BeEmpty(), "Expected scopeLogs")
			Expect(scopeLogs).To(HaveLen(1), "Expected a single scopeLog")

			logRecords := scopeLogs[0].LogRecords
			Expect(logRecords).ToNot(BeEmpty(), "Expected log records for the scope")
			Expect(logRecords).To(HaveLen(1), "Expected same count as sent for log records")

			// Inspect the first log record for correct fields
			logRecord := logRecords[0]
			Expect(logRecord.TimeUnixNano).ToNot(BeEmpty(), "Expect timestamp to be converted into unix nano")
			Expect(logRecord.Body.StringValue).ToNot(BeEmpty(), "Expect message to be populated")

			// Ensure we have log.type populated
			Expect(logRecord.Attributes).ToNot(BeEmpty(), "Expect logRecord attributes to be populated")
			logType := logRecord.FindStringValue(logRecord.LogTypeAttribute())
			Expect(logType).To(Equal(string(obs.InputTypeInfrastructure)))
		})

		Context("with tuning parameters", func() {
			DescribeTable("with compression", func(compression string) {
				obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
					FromInput(obs.InputTypeApplication).
					ToOtlpOutput(func(output *obs.OutputSpec) {
						output.OTLP.Tuning = &obs.OTLPTuningSpec{
							Compression: compression,
						}
					})

				Expect(framework.DeployWithVisitor(func(b *runtime.PodBuilder) error {
					return framework.AddOTELCollector(b, string(obs.OutputTypeOTLP))
				})).To(BeNil())

				appNamespace = framework.Pod.Namespace

				// Write message to namespace
				crioLine := functional.NewCRIOLogMessage(timestamp, "Format me to a c`ontainer message!", false)
				Expect(framework.WriteMessagesToNamespace(crioLine, appNamespace, 1)).To(Succeed())
				crioLine = functional.NewCRIOLogMessage(timestamp, "My second message.", false)
				Expect(framework.WriteMessagesToNamespace(crioLine, appNamespace, 1)).To(Succeed())
				crioLine = functional.NewCRIOLogMessage(timestamp, "My third and final message.", false)
				Expect(framework.WriteMessagesToNamespace(crioLine, appNamespace, 1)).To(Succeed())

				// Read logs
				raw, err := framework.ReadRawApplicationLogsFrom(string(obs.OutputTypeOTLP))
				Expect(err).To(BeNil(), "Expected no errors reading the logs for type")
				Expect(raw).ToNot(BeEmpty())

				logs, err := otlp.ParseLogs(raw[0])
				Expect(err).To(BeNil(), "Expected no errors parsing the logs")

				// Access the first (and only) ResourceLog
				// Inspect resource attributes
				resourceLog := logs.ResourceLogs[0]
				resource := resourceLog.Resource
				namespaceName := resource.FindStringValue(resource.NamespaceNameAttribute())
				Expect(namespaceName).To(Equal(appNamespace), "Expect namespace name to exist in resource attributes")

				logSource := resource.FindStringValue(resource.LogSourceAttribute())
				Expect(logSource).To(Equal("container"))

				clusterId := resource.FindStringValue(resource.ClusterIDAttribute())
				Expect(clusterId).ToNot(BeEmpty(), "Expected cluster.id resource attribute")

				scopeLogs := resourceLog.ScopeLogs
				Expect(scopeLogs).ToNot(BeEmpty(), "Expected scopeLogs")
				Expect(scopeLogs).To(HaveLen(1), "Expected a single scopeLog")

				logRecords := scopeLogs[0].LogRecords
				Expect(logRecords).ToNot(BeEmpty(), "Expected log records for the scope")
				Expect(logRecords).To(HaveLen(3), "Expected same count as sent for log records")

				// Inspect the first log record for correct fields
				logRecord := logRecords[0]
				Expect(logRecord.TimeUnixNano).To(Equal(timestampNano), "Expect timestamp to be converted into unix nano")
				Expect(logRecord.SeverityNumber).To(Equal(9), "Expect severityNumber to not be parsed to 9")
				Expect(logRecord.Body.StringValue).ToNot(BeEmpty(), "Expect message to be populated")

				// Ensure we have log.type populated
				Expect(logRecord.Attributes).ToNot(BeEmpty(), "Expect logRecord attributes to be populated")
				logType := logRecord.FindStringValue(logRecord.LogTypeAttribute())
				Expect(logType).To(Equal(string(obs.InputTypeApplication)))
			},
				Entry("should pass with gzip", "gzip"),
				Entry("should pass with zlib", "zlib"),
				Entry("should pass with no compression", "none"))
		})
	})
})
