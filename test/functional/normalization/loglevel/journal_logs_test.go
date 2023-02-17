package loglevel

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testfw "github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"k8s.io/apimachinery/pkg/util/sets"
)

var _ = Describe("[functional][normalization][loglevel] tests for message format of journal logs", func() {

	var (
		framework *functional.CollectorFunctionalFramework
	)

	const (
		expReadFail = "failread"
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFrameworkUsingCollector(testfw.LogCollectionType)
		functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameInfrastructure).
			ToFluentForwardOutput().
			FromInput(logging.InputNameAudit).
			ToElasticSearchOutput()
		Expect(framework.Deploy()).To(BeNil())
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	DescribeTable("when evaluating a journal log entry", func(priority int, expLevel string, options ...string) {
		// Write log line as input to collector
		logline := functional.NewJournalLog(priority, "*", "*")
		Expect(framework.WriteMessagesToInfraJournalLog(logline, 1)).To(BeNil())

		// Read line from Log Forward output
		raw, err := framework.ReadInfrastructureLogsFrom(logging.OutputTypeElasticsearch)
		if sets.NewString(options...).Has(expReadFail) {
			Expect(err).To(Not(BeNil()), "Exp. to not find any logs")
			return
		}

		Expect(err).To(BeNil(), "Expected no errors reading the logs")

		// Parse log line
		var logs []types.JournalLog
		err = types.ParseLogsFrom(utils.ToJsonLogs(raw), &logs, false)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")

		// Compare to expected template
		outputTestLog := logs[0]
		Expect(outputTestLog.Level).To(Equal(expLevel))
	},
		Entry("should recognize an emerg message", 0, "emerg"),
		Entry("should recognize an alert message", 1, "alert"),
		Entry("should recognize a crit message", 2, "crit"),
		Entry("should recognize an err message", 3, "err"),
		Entry("should recognize a warning message", 4, "warning"),
		Entry("should recognize a notice message", 5, "notice"),
		Entry("should recognize an info message", 6, "info"),
		Entry("should drop debug messages", 7, "debug", expReadFail),
	)
})
