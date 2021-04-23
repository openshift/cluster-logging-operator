package normalization

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"github.com/openshift/cluster-logging-operator/test/matchers"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/functional"
)

//Fast test for checking reassembly logic for split log by CRI-O.
//CRI-O split long string on parts. Part marked by 'P' letter and finished with 'F'.
//Example:
// 2021-03-31T12:59:28.573159188+00:00 stdout P First line of log entry
// 2021-03-31T12:59:28.573159188+00:00 stdout P Second line of the log entry
// 2021-03-31T12:59:28.573159188+00:00 stdout F Last line of the log entry
//
// Here we will emulate CRI-O split by direct writing formatted content
var _ = Describe("Reassembly split by CRI-O logs ", func() {

	var (
		framework *functional.FluentdFunctionalFramework
		timestamp = "2021-03-31T12:59:28.573159188+00:00"
	)

	BeforeEach(func() {
		framework = functional.NewFluentdFunctionalFramework()
		functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication).
			ToFluentForwardOutput()
		Expect(framework.Deploy()).To(BeNil())
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	It("should handle a single split log message", func() {
		//write partial log
		msg := functional.NewCRIOLogMessage(timestamp, "May ", true)
		matchers.ExpectOK(framework.WriteMessagesToApplicationLog(msg, 1),
			"Expected no errors writing the logs")
		//write partial log
		msg = functional.NewCRIOLogMessage(timestamp, "the force ", true)
		matchers.ExpectOK(framework.WriteMessagesToApplicationLog(msg, 1),
			"Expected no errors writing the logs")
		//write partial log
		msg = functional.NewCRIOLogMessage(timestamp, "be with ", true)
		matchers.ExpectOK(framework.WriteMessagesToApplicationLog(msg, 1),
			"Expected no errors writing the logs")
		//write final part of log entry
		msg = functional.NewCRIOLogMessage(timestamp, "you", false)
		matchers.ExpectOK(framework.WriteMessagesToApplicationLog(msg, 1),
			"Expected no errors writing the logs")
		raw, err := framework.ReadApplicationLogsFrom(logging.OutputTypeFluentdForward)
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		logs, err := types.ParseLogs(utils.ToJsonLogs(raw))
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		Expect(logs[0].Message).Should(Equal("May the force be with you"))
	})

	It("should handle a split log followed by a full log", func() {
		//write partial log
		msg := functional.NewCRIOLogMessage(timestamp, "Run, ", true)
		matchers.ExpectOK(framework.WriteMessagesToApplicationLog(msg, 1),
			"Expected no errors writing the logs")
		//write partial log
		msg = functional.NewCRIOLogMessage(timestamp, "Forest, ", true)
		matchers.ExpectOK(framework.WriteMessagesToApplicationLog(msg, 1),
			"Expected no errors writing the logs")
		//write final part of log entry
		msg = functional.NewCRIOLogMessage(timestamp, "Run!", false)
		matchers.ExpectOK(framework.WriteMessagesToApplicationLog(msg, 1),
			"Expected no errors writing the logs")
		//write single-line log entry
		msg = functional.NewCRIOLogMessage(timestamp, "Freedom!!!", false)
		matchers.ExpectOK(framework.WriteMessagesToApplicationLog(msg, 1),
			"Expected no errors writing the logs")

		raw, err := framework.ReadApplicationLogsFrom(logging.OutputTypeFluentdForward)
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		logs, err := types.ParseLogs(utils.ToJsonLogs(raw))
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		Expect(logs[0].Message).Should(Equal("Run, Forest, Run!"))
		Expect(logs[1].Message).Should(Equal("Freedom!!!"))
	})
})
