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

const truncationMarker = "..TRUNCATED"

func verifyTruncation(actualMessage, expectedOriginal string, maxSize *resource.Quantity, expectTruncated bool) {
	if expectTruncated {
		Expect(actualMessage).To(ContainSubstring(truncationMarker), "Expected truncation marker in message")

		if maxSize != nil {
			Expect(len(actualMessage)).To(BeNumerically("<=", int(maxSize.Value())+len(truncationMarker)),
				"Truncated message size exceeds configured limit + marker length")
		}
	} else {
		Expect(actualMessage).NotTo(ContainSubstring(truncationMarker), "Message should not contain truncation marker")
		Expect(actualMessage).To(Equal(expectedOriginal), "Message content does not match expected payload")
	}
}

// setupAppLogForwarder for application log
func setupAppLogForwarder(framework *functional.CollectorFunctionalFramework, maxSize *resource.Quantity) {
	testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
		FromInput(obs.InputTypeApplication, func(spec *obs.InputSpec) {
			spec.Name = "app-tuning-input"
			spec.Application = &obs.Application{}
			if maxSize != nil {
				spec.Application.Tuning = &obs.ContainerInputTuningSpec{
					MaxMessageSize: utils.GetPtr(maxSize.DeepCopy()),
				}
			}
		}).ToHttpOutput()

	Expect(framework.Deploy()).To(Succeed(), "Expected Log Forwarder deployment to succeed")
}

// setupInfraLogForwarder for infrastructure log
func setupInfraLogForwarder(framework *functional.CollectorFunctionalFramework, maxSize *resource.Quantity) {
	testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
		FromInput(obs.InputTypeInfrastructure, func(spec *obs.InputSpec) {
			spec.Name = "infra-tuning-input"
			spec.Infrastructure = &obs.Infrastructure{
				Sources: []obs.InfrastructureSource{obs.InfrastructureSourceContainer},
			}
			if maxSize != nil {
				spec.Infrastructure.Tuning = &obs.InfrastructureInputTuningSpec{
					Container: &obs.ContainerInputTuningSpec{
						MaxMessageSize: new(maxSize.DeepCopy()),
					},
				}
			}
		}).ToHttpOutput()

	Expect(framework.Deploy()).To(Succeed(), "Expected Log Forwarder deployment to succeed")
}

// writePartialCRIOLogs generates and writes an interleaved stream of partial logs followed by a final log
func writePartialCRIOLogs(numPartialLines, lineSize int, writeFn func(string, int) error) string {
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
	var expected strings.Builder
	expected.Grow((numPartialLines + 1) * lineSize) // Pre-allocate to prevent multiple resizing allocations

	part := strings.Repeat("p", lineSize)
	for range numPartialLines {
		expected.WriteString(part)
		logLine := functional.NewCRIOLogMessage(timestamp, part, true)
		Expect(writeFn(logLine, 1)).To(Succeed(), "Expected successful partial log write")
	}

	finalPart := strings.Repeat("f", lineSize)
	expected.WriteString(finalPart)
	finalLine := functional.NewCRIOLogMessage(timestamp, finalPart, false)
	Expect(writeFn(finalLine, 1)).To(Succeed(), "Expected successful final log write")

	return expected.String()
}

var _ = Describe("[Functional][LogForwarding][Inputs] Container Log Truncation", func() {
	var framework *functional.CollectorFunctionalFramework

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
	})

	AfterEach(func() {
		if framework != nil {
			framework.Cleanup()
			framework = nil
		}
	})

	Context("Application Logs", func() {
		DescribeTable("single message size limits",
			func(maxSize *resource.Quantity, messageSize int, expectTruncated bool) {
				setupAppLogForwarder(framework, maxSize)

				message := strings.Repeat("x", messageSize)
				timestamp := time.Now().UTC().Format(time.RFC3339Nano)
				logLine := functional.NewCRIOLogMessage(timestamp, message, false)
				Expect(framework.WriteMessagesToApplicationLog(logLine, 1)).To(Succeed())

				logs, err := framework.ReadApplicationLogsFrom(string(obs.OutputTypeHTTP))
				Expect(err).To(BeNil(), "Expected no errors reading logs")
				Expect(logs).To(HaveLen(1), "Expected exactly one log record")

				record := logs[0]
				verifyTruncation(record.Message, message, maxSize, expectTruncated)
				Expect(record.Kubernetes).NotTo(BeNil(), "Expected Kubernetes metadata to be present")
				Expect(record.Kubernetes.NamespaceName).NotTo(BeEmpty(), "Expected Kubernetes NamespaceName to be populated")
			},
			Entry("with 2KB limit, 4KB message (truncated)", resource.NewQuantity(2048, resource.BinarySI), 4096, true),
			Entry("with 4KB limit, 2KB message (not truncated)", resource.NewQuantity(4096, resource.BinarySI), 2048, false),
			Entry("with default limit, normal message", nil, 512, false),
		)

		DescribeTable("partial messages (merged lines) limits",
			func(maxSize *resource.Quantity, numPartialLines int, lineSize int, expectTruncated bool) {
				setupAppLogForwarder(framework, maxSize)

				expectedMsg := writePartialCRIOLogs(numPartialLines, lineSize, framework.WriteMessagesToApplicationLog)

				logs, err := framework.ReadApplicationLogsFrom(string(obs.OutputTypeHTTP))
				Expect(err).To(BeNil(), "Expected no errors reading logs")
				Expect(logs).To(HaveLen(1), "Expected exactly one log record")

				verifyTruncation(logs[0].Message, expectedMsg, maxSize, expectTruncated)
			},
			Entry("with 2KB limit, 5 partial lines (truncated)", resource.NewQuantity(2048, resource.BinarySI), 5, 512, true),
			Entry("with 4KB limit, 3 partial lines (not truncated)", resource.NewQuantity(4096, resource.BinarySI), 3, 512, false),
			Entry("with default limit, 2 partial lines (normal, not truncated)", nil, 2, 256, false),
		)
	})

	Context("Infrastructure Container Logs", func() {
		BeforeEach(func() {
			if framework != nil {
				framework.Cleanup()
			}
			framework = functional.NewCollectorFunctionalFramework(client.UseInfraNamespaceTestOption)
		})

		DescribeTable("infrastructure partial messages limits",
			func(maxSize *resource.Quantity, numPartialLines int, lineSize int, expectTruncated bool) {
				setupInfraLogForwarder(framework, maxSize)

				expectedMsg := writePartialCRIOLogs(numPartialLines, lineSize, framework.WriteMessagesToInfraContainerLog)

				rawLogs, err := framework.ReadInfrastructureLogsFrom(string(obs.OutputTypeHTTP))
				Expect(err).To(BeNil(), "Expected no errors reading infrastructure logs")
				Expect(rawLogs).To(HaveLen(1), "Expected exactly one infrastructure log record")

				var logs []types.InfraContainerLog
				err = types.ParseLogsFrom(utils.ToJsonLogs(rawLogs), &logs, false)
				Expect(err).To(BeNil(), "Expected no errors parsing infrastructure logs")
				Expect(logs).ToNot(BeEmpty(), "Expected parsed infrastructure logs to not be empty")

				record := logs[0]
				verifyTruncation(record.Message, expectedMsg, maxSize, expectTruncated)
				Expect(record.LogType).To(Equal(string(obs.InputTypeInfrastructure)), "Expected log_type to match Infrastructure")
			},
			Entry("with 2KB limit, 5 partial lines (truncated)", resource.NewQuantity(2048, resource.BinarySI), 5, 512, true),
			Entry("with 4KB limit, 3 partial lines (not truncated)", resource.NewQuantity(4096, resource.BinarySI), 3, 512, false),
			Entry("with default limit, 2 partial lines (normal, not truncated)", nil, 2, 256, false),
		)
	})
})
