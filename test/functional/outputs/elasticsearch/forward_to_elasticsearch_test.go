package elasticsearch

import (
	"sort"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"github.com/openshift/cluster-logging-operator/test/matchers"
	obstestruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
)

var _ = Describe("[Functional][Outputs][ElasticSearch] Logforwarding to ElasticSearch", func() {

	var (
		framework *functional.CollectorFunctionalFramework

		// Template expected as output Log
		outputLogTemplate = functional.NewApplicationLogTemplate()
	)

	Context("and writing log messages", func() {
		BeforeEach(func() {
			outputLogTemplate.ViaqIndexName = "app-write"
			framework = functional.NewCollectorFunctionalFramework()
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToElasticSearchOutput()
			Expect(framework.Deploy()).To(BeNil())
		})
		AfterEach(func() {
			framework.Cleanup()
		})
		It("should work well for valid utf-8 and replace not utf-8", func() {
			timestamp := functional.CRIOTime(time.Now())
			ukr := "привіт "
			jp := "こんにちは "
			ch := "你好"
			msg := functional.NewCRIOLogMessage(timestamp, ukr+jp+ch, false)
			Expect(framework.WriteMessagesToApplicationLog(msg, 1)).To(BeNil())

			Expect(framework.WriteMessagesWithNotUTF8SymbolsToLog()).To(BeNil())

			raw, err := framework.GetLogsFromElasticSearch(string(obs.OutputTypeElasticsearch), string(obs.InputTypeApplication))
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(raw).To(Not(BeEmpty()))
			// Parse log line
			var logs []types.ApplicationLog
			err = types.StrictlyParseLogsFromSlice(raw, &logs)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")
			Expect(len(logs)).To(Equal(2))
			//sort log by time before matching
			sort.Slice(logs, func(i, j int) bool {
				return logs[i].Timestamp.Before(logs[j].Timestamp)
			})

			Expect(logs[0].Message).To(Equal(ukr + jp + ch))
			Expect(logs[1].Message).To(Equal("������������"))
		})
	})

	DescribeTable("should be compatible with version", func(version functional.ElasticsearchVersion) {
		var secret *corev1.Secret
		var auth obs.HTTPAuthentication
		framework := functional.NewCollectorFunctionalFramework()
		if version > functional.ElasticsearchVersion7 {
			secret = runtime.NewSecret(framework.Namespace, "mysecret", map[string][]byte{
				constants.ClientUsername: []byte("admin"),
				constants.ClientPassword: []byte("elasticadmin"),
			})
			framework.Secrets = append(framework.Secrets, secret)
			auth = obs.HTTPAuthentication{
				Username: &obs.SecretReference{
					Key:        constants.ClientUsername,
					SecretName: "mysecret",
				},
				Password: &obs.SecretReference{
					Key:        constants.ClientPassword,
					SecretName: "mysecret",
				},
			}
		}
		outputLogTemplate.ViaqIndexName = "app-write"

		obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication).
			ToElasticSearchOutput(func(output *obs.OutputSpec) {
				output.Elasticsearch.Version = int(version)
				output.Elasticsearch.Authentication = &auth
			})
		defer framework.Cleanup()
		Expect(framework.Deploy()).To(BeNil())

		Expect(framework.WritesApplicationLogs(1)).To(BeNil())
		raw, err := framework.ReadLogsFrom(string(obs.OutputTypeElasticsearch), string(obs.InputTypeApplication))
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		Expect(raw).To(Not(BeEmpty()))

		// Parse log line
		var logs []types.ApplicationLog
		err = types.StrictlyParseLogsFromSlice(raw, &logs)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		// Compare to expected template
		outputTestLog := logs[0]
		outputLogTemplate.ViaqIndexName = ""
		Expect(outputTestLog).To(matchers.FitLogFormatTemplate(outputLogTemplate))
	},
		Entry("Elasticsearch v6", functional.ElasticsearchVersion6),
		Entry("Elasticsearch v7", functional.ElasticsearchVersion7),
		Entry("Elasticsearch v8", functional.ElasticsearchVersion8),
	)

	Context("with tuning parameters", func() {
		DescribeTable("with compression", func(compression string) {
			outputLogTemplate.ViaqIndexName = "app-write"
			framework = functional.NewCollectorFunctionalFramework()
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToElasticSearchOutput(func(output *obs.OutputSpec) {
					output.Elasticsearch.Tuning = &obs.ElasticsearchTuningSpec{
						Compression: compression,
					}
				})
			defer framework.Cleanup()
			Expect(framework.Deploy()).To(BeNil())

			// Write app logs
			timestamp := "2020-11-04T18:13:59.061892+00:00"
			applicationLogLine := functional.NewCRIOLogMessage(timestamp, "This is my test message", false)
			Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 2)).To(BeNil())

			raw, err := framework.ReadLogsFrom(string(obs.OutputTypeElasticsearch), string(obs.InputTypeApplication))

			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(raw).To(Not(BeEmpty()))

		},
			Entry("should pass with gzip", "gzip"),
			Entry("should pass with zlib", "zlib"),
			Entry("should pass with no compression", "none"))
	})

	DescribeTable("user defined index", func(index, expIndex string) {
		framework = functional.NewCollectorFunctionalFramework()
		obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication).
			ToElasticSearchOutput(func(output *obs.OutputSpec) {
				output.Elasticsearch.Index = index
			})
		defer framework.Cleanup()
		Expect(framework.Deploy()).To(BeNil())

		// Write app logs
		timestamp := "2020-11-04T18:13:59.061892+00:00"
		applicationLogLine := functional.NewCRIOLogMessage(timestamp, "This is my test message", false)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 2)).To(BeNil())

		raw, err := framework.GetLogsFromElasticSearchIndex(string(obs.OutputTypeElasticsearch), expIndex)
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		Expect(raw).To(Not(BeEmpty()))
	},
		Entry("should write to defined static index", "custom-index", "custom-index"),
		Entry("should write to defined dynamic index", `{.log_type||"none"}`, "application"),
		Entry("should write to defined static + dynamic index", `foo-{.log_type||"none"}`, "foo-application"),
		Entry("should write to defined static + fallback value if field is missing", `foo-{.missing||"none"}`, "foo-none"),
	)

	Context("elasticsearch authentication", func() {
		AfterEach(func() {
			framework.Cleanup()
		})

		It("should authenticate with username and password", func() {
			password := "4xpXpbq&rmPCF576N$Bz"
			secret := runtime.NewSecret("", "mysecret", map[string][]byte{
				constants.ClientUsername: []byte(functional.ElasticUsername),
				constants.ClientPassword: []byte(password),
			})

			outputLogTemplate.ViaqIndexName = "app-write"
			framework = functional.NewCollectorFunctionalFramework()
			framework.Secrets = append(framework.Secrets, secret)

			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToElasticSearchOutput(func(output *obs.OutputSpec) {
					output.Elasticsearch.Authentication = &obs.HTTPAuthentication{
						Username: &obs.SecretReference{
							Key:        constants.ClientUsername,
							SecretName: "mysecret",
						},
						Password: &obs.SecretReference{
							Key:        constants.ClientPassword,
							SecretName: "mysecret",
						},
					}
					output.Elasticsearch.Index = `{.log_type||"none"}`
				})

			esVisitor := func(b *runtime.PodBuilder) error {
				return framework.AddESOutputWithBasicSecurity(password, b, framework.Forwarder.Spec.Outputs[0])
			}

			Expect(framework.DeployWithVisitor(esVisitor)).To(BeNil())

			Expect(framework.WritesApplicationLogs(1)).To(BeNil())
			raw, err := framework.GetLogsFromElasticSearchIndex(string(obs.OutputTypeElasticsearch), string(obs.InputTypeApplication),
				functional.Option{Name: "username", Value: functional.ElasticUsername},
				functional.Option{Name: "password", Value: password})

			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(raw).To(Not(BeEmpty()))

			// Parse log line
			var logs []types.ApplicationLog
			err = types.StrictlyParseLogsFromSlice(raw, &logs)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")
			// Compare to expected template
			outputTestLog := logs[0]
			outputLogTemplate.ViaqIndexName = ""
			Expect(outputTestLog).To(matchers.FitLogFormatTemplate(outputLogTemplate))
		})

		It("should authenticate with custom bearer token", func() {
			secret := runtime.NewSecret("", "mysecret", map[string][]byte{
				constants.ClientUsername: []byte(functional.ElasticUsername),
				constants.ClientPassword: []byte(functional.ElasticPassword),
			})

			outputLogTemplate.ViaqIndexName = "app-write"
			framework = functional.NewCollectorFunctionalFramework()

			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToElasticSearchOutput(func(output *obs.OutputSpec) {
					output.Elasticsearch.URL = "http://elasticsearch:9200"
					output.Elasticsearch.Authentication = &obs.HTTPAuthentication{
						Token: &obs.BearerToken{
							From: obs.BearerTokenFromSecret,
							Secret: &obs.BearerTokenSecretKey{
								Name: secret.Name,
								Key:  constants.TokenKey,
							},
						},
					}
					output.Elasticsearch.Index = `{.log_type||"none"}`
				})

			// 1. Deploy Pod with ES container with token service enabled along with service
			Expect(framework.DeployESTokenPodWithService()).To(BeNil())

			// 2. Generate and get token from ES pod
			token, err := framework.GenerateESAccessToken(string(obs.OutputTypeElasticsearch))
			Expect(err).To(BeNil())
			Expect(token).ToNot(BeEmpty())

			// 3. Add token to secret
			cleanToken, err := strconv.Unquote(token)
			Expect(err).To(BeNil())

			secret.Data[constants.TokenKey] = []byte(cleanToken)
			framework.Secrets = append(framework.Secrets, secret)

			// 4. Deploy collector
			Expect(framework.DeployWithVisitor(nil)).To(BeNil())

			Expect(framework.WritesApplicationLogs(1)).To(BeNil())
			raw, err := framework.GetLogsFromElasticSearchIndex(string(obs.OutputTypeElasticsearch), string(obs.InputTypeApplication),
				functional.Option{Name: constants.TokenKey, Value: cleanToken})

			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(raw).To(Not(BeEmpty()))

			// Parse log line
			var logs []types.ApplicationLog
			err = types.StrictlyParseLogsFromSlice(raw, &logs)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")
			// Compare to expected template
			outputTestLog := logs[0]
			outputLogTemplate.ViaqIndexName = ""
			Expect(outputTestLog).To(matchers.FitLogFormatTemplate(outputLogTemplate))
		})
	})
})
