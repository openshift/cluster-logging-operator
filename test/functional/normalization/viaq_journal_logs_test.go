package normalization

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	. "github.com/openshift/cluster-logging-operator/test/matchers"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
)

var _ = Describe("[functional][normalization] ViaQ message format of journal logs", func() {
	var (
		framework *functional.CollectorFunctionalFramework
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeInfrastructure).
			ToElasticSearchOutput()
		Expect(framework.Deploy()).To(BeNil())
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	It("should format ViaQ journal logs", func() {

		expLog := functional.NewJournalInfrastructureLogTemplate()

		// Write log line as input to collector
		logline := functional.NewJournalLog(2, "here is my message", "*")
		Expect(framework.WriteMessagesToInfraJournalLog(logline, 1)).To(BeNil())

		// Read line from Log Forward output
		raw, err := framework.ReadInfrastructureLogsFrom(string(obs.OutputTypeElasticsearch))
		Expect(err).To(BeNil(), "Expected no errors reading the logs")

		// Parse log line
		var logs []types.JournalLog
		err = types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")

		// Compare to expected template
		outputTestLog := logs[0]
		Expect(outputTestLog.ViaQCommon).To(FitLogFormatTemplate(expLog.ViaQCommon))
		Expect(outputTestLog.Systemd.T).NotTo(Equal(types.T{}), "Exp. to be populated with something")
		Expect(outputTestLog.Systemd.U).NotTo(Equal(types.U{}), "Exp. to be populated with something")
	})
})
