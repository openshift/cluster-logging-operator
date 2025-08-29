package misc

import (
	"github.com/onsi/gomega/format"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/rand"
	"github.com/openshift/cluster-logging-operator/test/matchers"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
)

var _ = Describe("[Functional][Misc] ", func() {

	const WarnFileTooSmall = "Currently ignoring file too small to fingerprint"
	var (
		framework *functional.CollectorFunctionalFramework
		timestamp = "2021-03-31T12:59:28.573159188+00:00"
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication).
			ToHttpOutput()
		Expect(framework.Deploy()).To(BeNil())
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	It("should to proceed small file with new line", func() {
		msg := functional.NewCRIOLogMessage(timestamp, "A", false)
		matchers.ExpectOK(framework.WriteMessagesToApplicationLog(msg, 1),
			"Expected no errors writing the logs")

		logs, err := framework.ReadApplicationLogsFrom(string(obs.OutputTypeHTTP))
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		Expect(len(logs)).To(Equal(1))
		Expect(logs[0].Message).Should(Equal("A"))
		collectorLogs, err := framework.ReadCollectorLogs()
		Expect(err).To(BeNil(), "Expected no errors read collector logs")
		Expect(collectorLogs).ToNot(ContainSubstring(WarnFileTooSmall))
	})

	It("validate what collector emmit warning for big logs without new line", func() {
		big := rand.Word(1000)
		msg := functional.NewCRIOLogMessage(timestamp, string(big), false)
		// writing log, without new line symbol at the end
		matchers.ExpectOK(framework.WriteMessagesToLogWithoutNewLine(msg),
			"Expected no errors writing the logs")

		time.Sleep(20 * time.Second)

		format.MaxLength = 0
		collectorLogs, err := framework.ReadCollectorLogs()
		Expect(err).To(BeNil(), "Expected no errors read collector logs")
		Expect(collectorLogs).To(ContainSubstring(WarnFileTooSmall))

		// continue writing log, now with new line symbol, normal usecase
		msg = functional.NewCRIOLogMessage(timestamp, "New msg", false)
		matchers.ExpectOK(framework.WriteMessagesToApplicationLog(msg, 1),
			"Expected no errors writing the logs")

		logs, err := framework.ReadApplicationLogsFrom(string(obs.OutputTypeHTTP))
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		Expect(len(logs)).To(Equal(1))
	})

	It("validate what collector emmit warning on creating new log file for container", func() {
		matchers.ExpectOK(framework.EmulateCreationNewLogFileForContainer(), "Expected no errors")

		time.Sleep(20 * time.Second)

		format.MaxLength = 0
		collectorLogs, err := framework.ReadCollectorLogs()
		Expect(err).To(BeNil(), "Expected no errors read collector logs")
		Expect(collectorLogs).To(ContainSubstring(WarnFileTooSmall))
	})

})
