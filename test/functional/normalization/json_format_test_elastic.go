package normalization

import (
	"encoding/json"
	"github.com/ViaQ/logerr/log"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"github.com/openshift/cluster-logging-operator/test/matchers"
	"github.com/openshift/cluster-logging-operator/test/runtime"
	"strings"
)

var _ = Describe("[Functional][Outputs][ElasticSearch] FluentdForward Output to ElasticSearch", func() {

	const (
		elasticSearchTag   = "7.10.1"
		elasticSearchImage = "elasticsearch:" + elasticSearchTag

		// json message
		jsonMsg = "{\\\"name\\\":\\\"fred\\\",\\\"home\\\":\\\"bedrock\\\"}"
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

	)

	BeforeEach(func() {

		framework = functional.NewFluentdFunctionalFramework()
		functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication).
			ToElasticSearchOutput()
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	Context("when sending to ElasticSearch "+elasticSearchTag+" protocol", func() {
		FIt("should  be compatible", func() {

			framework.SetFluentConfigFileName("fluentd_structured_configuration_to_elastic.txt")
			addElasticSearchContainer := newVisitor(framework)
			Expect(framework.DeployWithVisitor(addElasticSearchContainer)).To(BeNil())
			Expect(framework.WritesJsonApplicationLogs(jsonMsg, 1)).To(BeNil())

			raw, err := framework.GetLogsFromElasticSearch(logging.OutputTypeElasticsearch, logging.InputNameJsonApplication)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(raw).To(Not(BeEmpty()))

			// Compare to expected template
			outputLogTemplate := functional.NewLogTemplate()
			outputLogTemplate.Message = strings.Replace(jsonMsg, "\\", "", -1)

			// Verify structured field
			var jsonRaw []map[string]interface{}
			err = json.Unmarshal([]byte(raw), &jsonRaw)
			Expect(err).To(BeNil(), "Expected no errors Unmarshal Log")
			jsonMessageMarshal,err  := json.Marshal(jsonRaw[0]["structured"])
			Expect(err).To(BeNil(), "Expected no errors Marshal Message")
			Expect(string(jsonMessageMarshal)).To(MatchJSON(outputLogTemplate.Message))

			// using custom elastic index index
			outputLogTemplate.ViaqIndexName = "*"

			// Verify all other fields
			logs, err := types.ParseLogs(raw)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")
			outputTestLog := logs[0]
			Expect(outputTestLog).To(matchers.FitLogFormatTemplate(outputLogTemplate))
		})
	})
})
