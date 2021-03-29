package normalization

import (
	"github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/test/runtime"

	"github.com/openshift/cluster-logging-operator/test/helpers/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/functional"
)

var _ = Describe("[Normalization] CRI-O normalization", func() {

	const (
		elasticSearchTag   = "7.10.1"
		elasticSearchImage = "elasticsearch:" + elasticSearchTag
	)

	var (
		framework *functional.FluentdFunctionalFramework

		newVisitor = func(f *functional.FluentdFunctionalFramework) runtime.PodBuilderVisitor {
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
		//outputLogTemplate = functional.NewApplicationLogTemplate()
	)

	BeforeEach(func() {
		framework = functional.NewFluentdFunctionalFramework()
		addElasticSearchContainer := newVisitor(framework)
		functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication).
			ToElasticSearchOutput()
		Expect(framework.DeployWithVisitor(addElasticSearchContainer)).To(BeNil())
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	It("send long log message (more than 8192 split on CRI-O)", func() {
		//write partial log
		Expect(framework.WritesPartialApplicationLog(8192)).To(BeNil())
		//finalize previous with F
		Expect(framework.WritesNApplicationLogsOfSize(1, 1000))

		raw, err := framework.GetLogsFromElasticSearch(logging.OutputTypeElasticsearch, logging.InputNameApplication)
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		logs, err := types.ParseLogs(raw)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		Expect(len(logs[0].Message)).Should(Equal(9192))
	})

})
