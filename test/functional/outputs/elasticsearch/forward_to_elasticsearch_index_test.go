package elasticsearch

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"github.com/openshift/cluster-logging-operator/test/matchers"
)

var _ = Describe("[Functional][Outputs][ElasticSearch][Index] FluentdForward Output to specific ElasticSearch index", func() {

	const (
		elasticSearchTag   = "7.10.1"
		LabelName          = "mytypekey"
		LabelValue         = "myindex"
		StructuredTypeName = "mytypename"
		AppIndex           = "app-write"
		jsonLog            = `
           {
			"host":"localhost",
			"labels": {
			  "client": "unknown",
			  "testname" : "json parsing"
			}
		   }`
	)

	var (
		framework         *functional.CollectorFunctionalFramework
		outputLogTemplate types.ApplicationLog
	)

	BeforeEach(func() {
		// Template expected as output Log
		outputLogTemplate = functional.NewApplicationLogTemplate()
		outputLogTemplate.ViaqIndexName = "app-write"
		framework = functional.NewCollectorFunctionalFramework()
	})
	AfterEach(func() {
		framework.Cleanup()
	})

	Context("when sending to ElasticSearch "+elasticSearchTag+" protocol", func() {
		withStructuredTypeName := func(spec *logging.OutputSpec) {
			spec.Elasticsearch = &logging.Elasticsearch{
				StructuredTypeName: StructuredTypeName,
			}
		}
		withK8sLabelsTypeKey := func(spec *logging.OutputSpec) {
			spec.Elasticsearch = &logging.Elasticsearch{
				StructuredTypeKey: fmt.Sprintf("kubernetes.labels.%s", LabelName),
			}
		}
		withOpenshiftLabelsTypeKey := func(spec *logging.OutputSpec) {
			spec.Elasticsearch = &logging.Elasticsearch{
				StructuredTypeKey: fmt.Sprintf("openshift.labels.%s", LabelName),
			}
		}
		setPodLabelsVisitor := func(pb *runtime.PodBuilder) error {
			pb.AddLabels(map[string]string{
				LabelName: LabelValue,
			})
			return nil
		}

		It("should send logs to structuredTypeName", func() {
			clfb := functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToOutputWithVisitor(withStructuredTypeName,
					logging.OutputTypeElasticsearch)
			clfb.Forwarder.Spec.Pipelines[0].Parse = "json"
			ESIndexName := fmt.Sprintf("app-%s-write", StructuredTypeName)
			Expect(framework.Deploy()).To(BeNil())

			applicationLogLine := functional.CreateAppLogFromJson(jsonLog)
			Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 10)).To(BeNil())

			raw, err := framework.GetLogsFromElasticSearchIndex(logging.OutputTypeElasticsearch, ESIndexName)
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
		It("should not send logs to structuredTypeName for infrastructure sources", func() {
			outputLogTemplate = functional.NewContainerInfrastructureLogTemplate()
			outputLogTemplate.ViaqIndexName = "infra-write"
			clfb := functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameInfrastructure).
				ToOutputWithVisitor(withStructuredTypeName,
					logging.OutputTypeElasticsearch)
			clfb.Forwarder.Spec.Pipelines[0].Parse = "json"
			Expect(framework.Deploy()).To(BeNil())

			applicationLogLine := functional.CreateAppLogFromJson(jsonLog)
			Expect(framework.WriteMessagesToInfraContainerLog(applicationLogLine, 10)).To(BeNil())

			raw, err := framework.GetLogsFromElasticSearchIndex(logging.OutputTypeElasticsearch, outputLogTemplate.ViaqIndexName)
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
		It("should not send to k8s label structuredTypeKey for infrastructure sources", func() {
			outputLogTemplate = functional.NewContainerInfrastructureLogTemplate()
			outputLogTemplate.ViaqIndexName = "infra-write"
			clfb := functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameInfrastructure).
				ToOutputWithVisitor(withK8sLabelsTypeKey, logging.OutputTypeElasticsearch)
			clfb.Forwarder.Spec.Pipelines[0].Parse = "json"
			visitors := append(framework.AddOutputContainersVisitors(), setPodLabelsVisitor)
			Expect(framework.DeployWithVisitors(visitors)).To(BeNil())

			applicationLogLine := functional.CreateAppLogFromJson(jsonLog)
			Expect(framework.WriteMessagesToInfraContainerLog(applicationLogLine, 10)).To(BeNil())
			raw, err := framework.GetLogsFromElasticSearchIndex(logging.OutputTypeElasticsearch, "infra-write")
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
		It("should send to k8s label structuredTypeKey", func() {
			clfb := functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToOutputWithVisitor(withK8sLabelsTypeKey, logging.OutputTypeElasticsearch)
			clfb.Forwarder.Spec.Pipelines[0].Parse = "json"
			ESIndexName := fmt.Sprintf("app-%s-write", LabelValue)
			visitors := append(framework.AddOutputContainersVisitors(), setPodLabelsVisitor)
			Expect(framework.DeployWithVisitors(visitors)).To(BeNil())

			applicationLogLine := functional.CreateAppLogFromJson(jsonLog)
			Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 10)).To(BeNil())
			raw, err := framework.GetLogsFromElasticSearchIndex(logging.OutputTypeElasticsearch, ESIndexName)
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
		It("should send to openshift label structuredTypeKey", func() {
			clfb := functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToOutputWithVisitor(withOpenshiftLabelsTypeKey, logging.OutputTypeElasticsearch)
			clfb.Forwarder.Spec.Pipelines[0].Labels = map[string]string{
				LabelName: LabelValue,
			}
			clfb.Forwarder.Spec.Pipelines[0].Parse = "json"
			ESIndexName := fmt.Sprintf("app-%s-write", LabelValue)
			Expect(framework.Deploy()).To(BeNil())

			applicationLogLine := functional.CreateAppLogFromJson(jsonLog)
			Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 1)).To(BeNil())
			raw, err := framework.GetLogsFromElasticSearchIndex(logging.OutputTypeElasticsearch, ESIndexName)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(raw).To(Not(BeEmpty()))

			// Parse log line
			var logs []types.ApplicationLog
			err = types.StrictlyParseLogs(raw, &logs)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")
			Expect(logs).To(Not(BeEmpty()), "Expected to find logs indexed")
			// Compare to expected template
			outputTestLog := logs[0]
			outputLogTemplate.ViaqIndexName = ""
			Expect(outputTestLog).To(matchers.FitLogFormatTemplate(outputLogTemplate))
		})
		Context("and enabling sending each container log to different indices", func() {
			It("should send one container's log as defined by the annotation and the other defined by structuredTypeKey", func() {
				clfb := functional.NewClusterLogForwarderBuilder(framework.Forwarder).
					FromInput(logging.InputNameApplication).
					ToOutputWithVisitor(func(spec *logging.OutputSpec) {
						spec.OutputTypeSpec = logging.OutputTypeSpec{
							Elasticsearch: &logging.Elasticsearch{
								EnableStructuredContainerLogs: true,
							},
						}
					}, logging.OutputTypeElasticsearch)
				clfb.Forwarder.Spec.Pipelines[0].Parse = "json"

				containerStructuredIndexValue := "bar"
				visitors := framework.AddOutputContainersVisitors()
				visitors = append(visitors, func(builder *runtime.PodBuilder) error {
					builder.AddAnnotation(fmt.Sprintf("containerType.logging.openshift.io/%s", constants.CollectorName), containerStructuredIndexValue)
					return nil
				})
				Expect(framework.DeployWithVisitors(visitors)).To(BeNil())

				applicationLogLine := functional.CreateAppLogFromJson(jsonLog)
				Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 1)).To(BeNil())
				Expect(framework.WriteMessagesToApplicationLogForContainer(applicationLogLine, "other", 1)).To(BeNil())

				for index, containerName := range map[string]string{fmt.Sprintf("app-%s-write", containerStructuredIndexValue): constants.CollectorName, "app-write": "other"} {
					raw, err := framework.GetLogsFromElasticSearchIndex(logging.OutputTypeElasticsearch, index)
					Expect(err).To(BeNil(), "Expected no errors reading the logs from index ", index)
					Expect(raw).To(Not(BeEmpty()))

					// Parse log line
					var logs []types.ApplicationLog
					err = types.StrictlyParseLogs(raw, &logs)
					Expect(err).To(BeNil(), "Expected no errors parsing the logs")
					Expect(logs).To(Not(BeEmpty()), "Expected to find a log in index", index)
					// Compare to expected template
					Expect(logs[0].Kubernetes.ContainerName).To(Equal(containerName), "Exp. to find a log entry for container", containerName, "in index", index)
				}

			})
		})

		Context("if structured type name/key not configured", func() {
			It("should send logs to app-write", func() {
				clfb := functional.NewClusterLogForwarderBuilder(framework.Forwarder).
					FromInput(logging.InputNameApplication).
					ToElasticSearchOutput()
				clfb.Forwarder.Spec.Pipelines[0].Parse = "json"
				Expect(framework.Deploy()).To(BeNil())

				applicationLogLine := functional.CreateAppLogFromJson(jsonLog)
				Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 10)).To(BeNil())
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
		Context("if elasticsearch structuredTypeKey wrongly configured", func() {
			It("should send logs to app-write", func() {
				clfb := functional.NewClusterLogForwarderBuilder(framework.Forwarder).
					FromInput(logging.InputNameApplication).
					ToOutputWithVisitor(func(spec *logging.OutputSpec) {
						spec.Elasticsearch = &logging.Elasticsearch{
							StructuredTypeKey: "junk",
						}
					}, logging.OutputTypeElasticsearch)
				clfb.Forwarder.Spec.Pipelines[0].Parse = "json"
				Expect(framework.Deploy()).To(BeNil())

				applicationLogLine := functional.CreateAppLogFromJson(jsonLog)
				Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 10)).To(BeNil())
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
		Context("if json parsing failed", func() {
			It("should send logs to app-write", func() {
				clfb := functional.NewClusterLogForwarderBuilder(framework.Forwarder).
					FromInput(logging.InputNameApplication).
					ToOutputWithVisitor(withK8sLabelsTypeKey, logging.OutputTypeElasticsearch)
				clfb.Forwarder.Spec.Pipelines[0].Parse = "json"
				Expect(framework.Deploy()).To(BeNil())

				// Write log line as input to fluentd
				invalidJson := `{"key":"v}`
				timestamp := "2020-11-04T18:13:59.061892+00:00"
				//expectedMessage := invalidJson
				applicationLogLine := functional.NewCRIOLogMessage(timestamp, invalidJson, false)
				Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 10)).To(BeNil())

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
