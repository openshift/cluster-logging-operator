package normalization

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/matchers"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("[Functional][Normalization] OversizedMessageBehavior for merged partial logs", func() {

	var (
		framework *functional.CollectorFunctionalFramework
		timestamp = "2021-03-31T12:59:28.573159188+00:00"
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
	})

	AfterEach(func() {
		framework.Cleanup()
	})

	Context("when oversizedLineBehavior is drop (default)", func() {
		BeforeEach(func() {
			testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication, func(spec *obs.InputSpec) {
					spec.Name = "drop-app"
					spec.Application.Tuning = &obs.ContainerInputTuningSpec{
						MaxMessageSize: utils.GetPtr(resource.MustParse("8Ki")),
					}
				}).
				ToHttpOutput()
			Expect(framework.Deploy()).To(BeNil())
		})

		It("should drop oversized merged lines and forward normal logs", func() {
			msg := functional.NewCRIOLogMessage(timestamp, "short message", false)
			matchers.ExpectOK(framework.WriteMessagesToApplicationLog(msg, 1))

			// Each partial is ~1000 bytes; 16Ki total exceeds 8Ki merged limit
			Expect(framework.WriteApplicationLogOfSizeAsPartials(16 * 1024)).To(BeNil())

			msg = functional.NewCRIOLogMessage(timestamp, "another short message", false)
			matchers.ExpectOK(framework.WriteMessagesToApplicationLog(msg, 1))

			logs, err := framework.ReadApplicationLogsFrom(string(obs.OutputTypeHTTP))
			Expect(err).To(BeNil(), "Expected no errors reading logs")
			Expect(logs).To(HaveLen(2))

			collectorLog, err := framework.ReadCollectorLogs()
			Expect(err).To(BeNil())
			Expect(collectorLog).To(
				ContainSubstring("Found line that exceeds max_merged_line_bytes; discarding."),
			)
		})
	})

	Context("when oversizedLineBehavior is truncate", func() {
		BeforeEach(func() {
			testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication, func(spec *obs.InputSpec) {
					spec.Name = "truncate-app"
					spec.Application.Tuning = &obs.ContainerInputTuningSpec{
						MaxMessageSize:           utils.GetPtr(resource.MustParse("8Ki")),
						OversizedMessageBehavior: utils.GetPtr(obs.OversizedMessageBehaviorTruncate),
					}
				}).
				ToHttpOutput()
			Expect(framework.Deploy()).To(BeNil())
		})

		It("should truncate oversized merged lines instead of dropping them", func() {
			msg := functional.NewCRIOLogMessage(timestamp, "short message", false)
			matchers.ExpectOK(framework.WriteMessagesToApplicationLog(msg, 1))

			// Each partial is ~1000 bytes; 16Ki total exceeds 8Ki merged limit
			Expect(framework.WriteApplicationLogOfSizeAsPartials(16 * 1024)).To(BeNil())

			msg = functional.NewCRIOLogMessage(timestamp, "another short message", false)
			matchers.ExpectOK(framework.WriteMessagesToApplicationLog(msg, 1))

			// All three logs should be forwarded — the oversized one is truncated, not dropped
			logs, err := framework.ReadApplicationLogsFrom(string(obs.OutputTypeHTTP))
			Expect(err).To(BeNil(), "Expected no errors reading logs")
			Expect(logs).To(HaveLen(3))

			// The truncated log's message should be no longer than maxMessageSize (8Ki)
			Expect(len(logs[1].Message)).To(BeNumerically("<=", 8*1024))
			// The truncated log should end with the ..TRUNCATED suffix
			Expect(logs[1].Message).To(HaveSuffix("..TRUNCATED"))

			collectorLog, err := framework.ReadCollectorLogs()
			Expect(err).To(BeNil())
			Expect(collectorLog).To(
				ContainSubstring("Truncated line that exceeds max_merged_line_bytes."),
			)
		})
	})
})
