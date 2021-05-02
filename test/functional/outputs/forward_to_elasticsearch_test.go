package outputs

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("[Functional][Outputs][ElasticSearch] FluentdForward Output to ElasticSearch", func() {

	var (
		framework *functional.CollectorFunctionalFramework

		newVisitor = func(f *functional.CollectorFunctionalFramework) runtime.PodBuilderVisitor {
			return func(b *runtime.PodBuilder) error {
				log.V(2).Info("Adding ElasticSearch output container", "name", logging.OutputTypeElasticsearch)
				b.AddContainer(logging.OutputTypeElasticsearch, elasticSearchImage).
					AddEnvVar("discovery.type", "single-node").
					AddRunAsUser(2000).
					End()
				return nil
			}
		}

		// Template expected as output Log
		outputLogTemplate = functional.NewApplicationLogTemplate()
	)

	BeforeEach(func() {

		framework = functional.NewCollectorFunctionalFramework()
		addElasticSearchContainer := newVisitor(framework)
		functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication).
			ToElasticSearchOutput()
		Expect(framework.Deploy()).To(BeNil())
		Expect(framework.WritesApplicationLogs(1)).To(BeNil())
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	Context("when sending to ElasticSearch "+functional.ElasticSearchTag+" protocol", func() {
		It("should  be compatible", func() {
			raw, err := framework.GetLogsFromElasticSearch(logging.OutputTypeElasticsearch, logging.InputNameApplication)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(raw).To(Not(BeEmpty()))

			// Parse log line
			var logs []types.ApplicationLog
			err = types.StrictlyParseLogs(raw, &logs)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")
			// Compare to expected template
			outputTestLog := logs[0]
			outputLogTemplate.ViaqIndexName = ""
			Expect(outputTestLog).To(matchers.FitLogFormatTemplate(outputLogTemplate))
		})
	})
})
