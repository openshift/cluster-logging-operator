package inputs

import (
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
	"k8s.io/apimachinery/pkg/api/resource"
)

func verifyTruncation(actualMessage, expectedOriginal string, maxSize *resource.Quantity, expectTruncated bool) {
	const marker = "..TRUNCATED"
	if expectTruncated {
		Expect(actualMessage).To(ContainSubstring(marker), "Expected truncation marker")
		Expect(len(actualMessage)).To(BeNumerically("<=", int(maxSize.Value())+len(marker)),
			"Truncated message size exceeds configured limit + marker length")
	} else {
		Expect(actualMessage).NotTo(ContainSubstring(marker), "Message should not contain truncation marker")
		Expect(actualMessage).To(Equal(expectedOriginal), "Message content does not match expected payload")
	}
}

var _ = Describe("[Functional][LogForwarding][Inputs] Log Truncation", func() {
	var framework *functional.CollectorFunctionalFramework

	AfterEach(func() {
		if framework != nil {
			framework.Cleanup()
			framework = nil
		}
	})

	Context("Application Logs", func() {
		BeforeEach(func() {
			framework = functional.NewCollectorFunctionalFramework()
		})

		DescribeTable("single message size limits",
			func(maxSize *resource.Quantity, messageSize int, expectTruncated bool) {
				testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
					FromInput(obs.InputTypeApplication, func(spec *obs.InputSpec) {
						spec.Name = "app-tuning-input"
						spec.Application = &obs.Application{}
						if maxSize != nil {
							spec.Application.Tuning = &obs.ContainerInputTuningSpec{
								MaxMessageSize: utils.GetPtr(maxSize.DeepCopy()),
							}
						}
					}).ToElasticSearchOutput()

				Expect(framework.Deploy()).To(Succeed())

				message := strings.Repeat("x", messageSize)
				timestamp := time.Now().UTC().Format(time.RFC3339Nano)
				logLine := functional.NewCRIOLogMessage(timestamp, message, false)
				Expect(framework.WriteMessagesToApplicationLog(logLine, 1)).To(Succeed())

				logs, err := framework.ReadApplicationLogsFrom(string(obs.OutputTypeElasticsearch))
				Expect(err).To(BeNil(), "Expected no errors reading logs")
				Expect(logs).To(HaveLen(1), "Expected exactly one log record")

				verifyTruncation(logs[0].Message, message, maxSize, expectTruncated)
				Expect(logs[0].Kubernetes).NotTo(BeNil())
				Expect(logs[0].Kubernetes.NamespaceName).NotTo(BeEmpty())
			},
			Entry("with 2KB limit, 4KB message (truncated)",
				resource.NewQuantity(2048, resource.BinarySI), 4096, true),
			Entry("with 4KB limit, 2KB message (not truncated)",
				resource.NewQuantity(4096, resource.BinarySI), 2048, false),
			Entry("with default limit, normal message",
				nil, 512, false),
		)

		DescribeTable("partial messages (merged lines) limits",
			func(maxSize *resource.Quantity, numPartialLines int, lineSize int, expectTruncated bool) {
				testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
					FromInput(obs.InputTypeApplication, func(spec *obs.InputSpec) {
						spec.Name = "app-tuning-input"
						spec.Application = &obs.Application{}
						if maxSize != nil {
							spec.Application.Tuning = &obs.ContainerInputTuningSpec{
								MaxMessageSize: utils.GetPtr(maxSize.DeepCopy()),
							}
						}
					}).ToElasticSearchOutput()

				Expect(framework.Deploy()).To(Succeed())

				timestamp := time.Now().UTC().Format(time.RFC3339Nano)
				var expected strings.Builder
				for range numPartialLines {
					part := strings.Repeat("p", lineSize)
					expected.WriteString(part)
					logLine := functional.NewCRIOLogMessage(timestamp, part, true)
					Expect(framework.WriteMessagesToApplicationLog(logLine, 1)).To(Succeed())
				}
				finalPart := strings.Repeat("f", lineSize)
				expected.WriteString(finalPart)
				finalLine := functional.NewCRIOLogMessage(timestamp, finalPart, false)
				Expect(framework.WriteMessagesToApplicationLog(finalLine, 1)).To(Succeed())

				logs, err := framework.ReadApplicationLogsFrom(string(obs.OutputTypeElasticsearch))
				Expect(err).To(BeNil(), "Expected no errors reading logs")
				Expect(logs).To(HaveLen(1), "Expected exactly one log record")

				verifyTruncation(logs[0].Message, expected.String(), maxSize, expectTruncated)
			},
			Entry("with 2KB limit, 5 partial lines (truncated)",
				resource.NewQuantity(2048, resource.BinarySI), 5, 512, true),
			Entry("with 4KB limit, 3 partial lines (not truncated)",
				resource.NewQuantity(4096, resource.BinarySI), 3, 512, false),
			Entry("with default limit, 2 partial lines (normal, not truncated)",
				nil, 2, 256, false),
		)
	})

	Context("Infrastructure Container Logs", func() {
		BeforeEach(func() {
			framework = functional.NewCollectorFunctionalFramework(client.UseInfraNamespaceTestOption)
		})

		DescribeTable("infrastructure partial messages limits",
			func(maxSize *resource.Quantity, numPartialLines int, lineSize int, expectTruncated bool) {
				testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
					FromInput(obs.InputTypeInfrastructure, func(spec *obs.InputSpec) {
						spec.Name = "infra-tuning-input"
						infra := &obs.Infrastructure{
							Sources: []obs.InfrastructureSource{obs.InfrastructureSourceContainer},
						}
						if maxSize != nil {
							infra.Tuning = &obs.InfrastructureInputTuningSpec{
								Container: &obs.ContainerInputTuningSpec{
									MaxMessageSize: utils.GetPtr(maxSize.DeepCopy()),
								},
							}
						}
						spec.Infrastructure = infra
					}).ToElasticSearchOutput()

				Expect(framework.Deploy()).To(Succeed())

				timestamp := time.Now().UTC().Format(time.RFC3339Nano)
				var expected strings.Builder
				for range numPartialLines {
					part := strings.Repeat("p", lineSize)
					expected.WriteString(part)
					logLine := functional.NewCRIOLogMessage(timestamp, part, true)
					Expect(framework.WriteMessagesToInfraContainerLog(logLine, 1)).To(Succeed())
				}
				finalPart := strings.Repeat("f", lineSize)
				expected.WriteString(finalPart)
				finalLine := functional.NewCRIOLogMessage(timestamp, finalPart, false)
				Expect(framework.WriteMessagesToInfraContainerLog(finalLine, 1)).To(Succeed())

				rawLogs, err := framework.ReadInfrastructureLogsFrom(string(obs.OutputTypeElasticsearch))
				Expect(err).To(BeNil(), "Expected no errors reading infrastructure logs")
				Expect(rawLogs).To(HaveLen(1), "Expected exactly one infrastructure log record")

				var logs []types.JournalLog
				err = types.StrictlyParseLogs(utils.ToJsonLogs(rawLogs), &logs)
				Expect(err).To(BeNil(), "Expected no errors parsing infrastructure logs")
				record := logs[0]
				verifyTruncation(record.Message, expected.String(), maxSize, expectTruncated)
				Expect(record.LogType).To(Equal(string(obs.InputTypeInfrastructure)))
			},
			Entry("with 2KB limit, 5 partial lines (truncated)",
				resource.NewQuantity(2048, resource.BinarySI), 5, 512, true),
			Entry("with 4KB limit, 3 partial lines (not truncated)",
				resource.NewQuantity(4096, resource.BinarySI), 3, 512, false),
			Entry("with default limit, 2 partial lines (normal, not truncated)",
				nil, 2, 256, false),
		)
	})
})
