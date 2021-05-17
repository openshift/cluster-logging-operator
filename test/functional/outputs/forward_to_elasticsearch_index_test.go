package outputs

import (
	"fmt"

	"github.com/ViaQ/logerr/log"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"github.com/openshift/cluster-logging-operator/test/matchers"
	"github.com/openshift/cluster-logging-operator/test/runtime"
)

var _ = Describe("[Functional][Outputs][ElasticSearch][Index] FluentdForward Output to specific ElasticSearch index", func() {

	const (
		elasticSearchTag   = "7.10.1"
		elasticSearchImage = "elasticsearch:" + elasticSearchTag
		StructuredIndexKey = "structuredIndexKey"
		IndexValue         = "myindex"
		AppIndex           = "app-write"
	)

	var (
		framework                 *functional.FluentdFunctionalFramework
		addElasticSearchContainer runtime.PodBuilderVisitor

		newVisitor = func(f *functional.FluentdFunctionalFramework) runtime.PodBuilderVisitor {
			return func(b *runtime.PodBuilder) error {
				log.V(2).Info("Adding ElasticSearch output container", "name", logging.OutputTypeElasticsearch)
				b.AddLabels(map[string]string{
					StructuredIndexKey: IndexValue,
				}).
					AddContainer(logging.OutputTypeElasticsearch, elasticSearchImage).
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
		framework = functional.NewFluentdFunctionalFramework()
		addElasticSearchContainer = newVisitor(framework)
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	Context("when sending to ElasticSearch "+elasticSearchTag+" protocol", func() {
		It("should send logs to structuredIndexName", func() {
			StructuredIndexName := "myindex"
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToOutputWithVisitor(func(spec *logging.OutputSpec) {
					spec.Elasticsearch = &logging.Elasticsearch{
						StructuredIndexName: StructuredIndexName,
					}
				}, logging.OutputTypeElasticsearch)
			Expect(framework.DeployWithVisitor(addElasticSearchContainer)).To(BeNil())
			Expect(framework.WritesApplicationLogs(1)).To(BeNil())
			raw, err := framework.GetLogsFromElasticSearchIndex(logging.OutputTypeElasticsearch, StructuredIndexName)
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
	Context("when sending to ElasticSearch "+elasticSearchTag+" protocol", func() {
		It("should send to k8s label structuredIndexKey", func() {
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToOutputWithVisitor(func(spec *logging.OutputSpec) {
					spec.Elasticsearch = &logging.Elasticsearch{
						StructuredIndexKey: fmt.Sprintf("kubernetes.labels.%s", StructuredIndexKey),
					}
				}, logging.OutputTypeElasticsearch)
			Expect(framework.DeployWithVisitor(addElasticSearchContainer)).To(BeNil())

			Expect(framework.WritesApplicationLogs(1)).To(BeNil())
			raw, err := framework.GetLogsFromElasticSearchIndex(logging.OutputTypeElasticsearch, IndexValue)
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
		It("should send to openshift label structuredIndexKey", func() {
			clfb := functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToOutputWithVisitor(func(spec *logging.OutputSpec) {
					spec.Elasticsearch = &logging.Elasticsearch{
						StructuredIndexKey: fmt.Sprintf("openshift.labels.%s", StructuredIndexKey),
					}
				}, logging.OutputTypeElasticsearch)
			clfb.Forwarder.Spec.Pipelines[0].Labels = map[string]string{
				StructuredIndexKey: IndexValue,
			}
			Expect(framework.DeployWithVisitor(addElasticSearchContainer)).To(BeNil())

			Expect(framework.WritesApplicationLogs(1)).To(BeNil())
			raw, err := framework.GetLogsFromElasticSearchIndex(logging.OutputTypeElasticsearch, IndexValue)
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
		Context("if elasticsearch index not configured", func() {
			It("should send logs to app-write", func() {
				functional.NewClusterLogForwarderBuilder(framework.Forwarder).
					FromInput(logging.InputNameApplication).
					ToOutputWithVisitor(func(spec *logging.OutputSpec) {
					}, logging.OutputTypeElasticsearch)
				Expect(framework.DeployWithVisitor(addElasticSearchContainer)).To(BeNil())

				Expect(framework.WritesApplicationLogs(1)).To(BeNil())
				raw, err := framework.GetLogsFromElasticSearchIndex(logging.OutputTypeElasticsearch, AppIndex)
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
		Context("if elasticsearch structuredIndexKey wrongly configured", func() {
			It("should send logs to app-write", func() {
				functional.NewClusterLogForwarderBuilder(framework.Forwarder).
					FromInput(logging.InputNameApplication).
					ToOutputWithVisitor(func(spec *logging.OutputSpec) {
						spec.Elasticsearch = &logging.Elasticsearch{
							StructuredIndexKey: "junk",
						}
					}, logging.OutputTypeElasticsearch)
				Expect(framework.DeployWithVisitor(addElasticSearchContainer)).To(BeNil())

				Expect(framework.WritesApplicationLogs(1)).To(BeNil())
				raw, err := framework.GetLogsFromElasticSearchIndex(logging.OutputTypeElasticsearch, AppIndex)
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
})
