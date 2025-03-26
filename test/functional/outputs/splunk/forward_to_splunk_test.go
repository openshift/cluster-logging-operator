package splunk

import (
	"fmt"
	"github.com/onsi/ginkgo/config"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	obstestruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
	"regexp"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	v1 "k8s.io/api/core/v1"
	"strings"
)

var _ = Describe("Forwarding to Splunk", func() {
	const splunkSecretName = "splunk-secret"
	var (
		framework      *functional.CollectorFunctionalFramework
		secret         *v1.Secret
		hecSecretKey   = *internalobs.NewSecretReference(constants.SplunkHECTokenKey, splunkSecretName)
		regexpSource   = regexp.MustCompile(`"source"\s*:\s*"([^"]+)"`)
		regexpHost     = regexp.MustCompile(`"host"\s*:\s*"([^"]+)"`)
		regexOpenshift = regexp.MustCompile(`{"openshift":{"cluster_id":".*","sequence":[0-9]+}}`)
		regexMessage   = regexp.MustCompile(`{"message":".*"}`)
	)

	BeforeEach(func() {
		config.DefaultReporterConfig.SlowSpecThreshold = 120
		framework = functional.NewCollectorFunctionalFramework()
		secret = runtime.NewSecret(framework.Namespace, splunkSecretName,
			map[string][]byte{
				"hecToken": functional.HecToken,
			},
		)
	})

	AfterEach(func() {
		framework.Cleanup()
	})

	waitOnSplunk := func() {
		time.Sleep(5 * time.Second)
		Eventually(func() string {
			// Run the Splunk CLI status command to check if splunkd is running
			output, err := framework.ReadSplunkStatus(framework.Namespace, framework.Name)
			if err != nil {
				return output
			}
			return output
		}, 90*time.Second, 3*time.Second).Should(SatisfyAll(
			ContainSubstring("splunkd is running"),
			ContainSubstring("splunk helpers are running"),
		))
	}

	It("should accept application logs", func() {
		obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeApplication).
			ToSplunkOutput(hecSecretKey, func(output *obs.OutputSpec) {
				output.Splunk.Index = "main"
			})
		framework.Secrets = append(framework.Secrets, secret)
		Expect(framework.Deploy()).To(BeNil())

		// Wait for splunk to be ready
		waitOnSplunk()

		// Write app logs
		timestamp := "2020-11-04T18:13:59.061892+00:00"
		applicationLogLine := functional.NewCRIOLogMessage(timestamp, "This is my test message", false)
		Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 2)).To(BeNil())

		// Read app logs
		logs, err := framework.ReadLogsByTypeFromSplunk(framework.Namespace, framework.Name, string(obs.InputTypeApplication))
		Expect(err).To(BeNil(), "Expected no errors getting logs from splunk")
		Expect(logs).ToNot(BeEmpty())

		// Parse the logs
		var appLogs []types.ApplicationLog
		jsonString := fmt.Sprintf("[%s]", strings.Join(logs, ","))
		err = types.ParseLogsFrom(jsonString, &appLogs, false)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")

		outputTestLog := appLogs[0]
		Expect(outputTestLog.LogType).To(Equal(string(obs.InputTypeApplication)))
	})

	It("should accept audit logs without timestamp unexpected type warning (see: https://issues.redhat.com/browse/LOG-4672)", func() {
		obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(obs.InputTypeAudit).
			ToSplunkOutput(hecSecretKey, func(output *obs.OutputSpec) {
				output.Splunk.Index = "main"
			})
		framework.Secrets = append(framework.Secrets, secret)
		Expect(framework.Deploy()).To(BeNil())

		// Wait for splunk to be ready
		waitOnSplunk()

		// Write audit logs
		timestamp, _ := time.Parse(time.RFC3339Nano, "2024-04-16T09:46:19.116+00:00")
		auditLogLine := functional.NewAuditHostLog(timestamp)
		writeAuditLogs := framework.WriteMessagesToAuditLog(auditLogLine, 1)
		Expect(writeAuditLogs).To(BeNil(), "Expect no errors writing audit logs")

		// Read audit logs
		logs, err := framework.ReadLogsByTypeFromSplunk(framework.Namespace, framework.Name, string(obs.InputTypeAudit))
		Expect(err).To(BeNil(), "Expected no errors getting logs from splunk")
		Expect(logs).ToNot(BeEmpty())

		// Parse the logs
		var auditLogs []types.AuditLog
		jsonString := fmt.Sprintf("[%s]", strings.Join(logs, ","))
		err = types.ParseLogsFrom(jsonString, &auditLogs, false)
		Expect(err).To(BeNil(), "Expected no errors parsing the logs")
		Expect(len(auditLogs)).To(Equal(1), "Expected one audit log")

		Expect(auditLogs[0].LogType).To(Equal(string(obs.InputTypeAudit)), "Expected audit log type")
		Expect(auditLogs[0].Level).To(Equal("default"), "Expected audit log level to default")

		collectorLog, err := framework.ReadCollectorLogs()
		Expect(err).To(BeNil(), "Expected no errors reading the collector logs")
		Expect(collectorLog).ToNot(BeEmpty(), "Expected collector logs to not be empty")
		tsWarn := "Timestamp was an unexpected type. Deferring to Splunk to set the timestamp"
		Expect(strings.Contains(collectorLog, tsWarn)).To(BeFalse(), "Expected collector logs to NOT contain timestamp unexpected type warning")
	})

	Context("splunk index", func() {
		DescribeTable("with user defined indices", func(index, expIndex string) {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToSplunkOutput(hecSecretKey, func(output *obs.OutputSpec) {
					if index != "" {
						output.Splunk.Index = index
					}
				})
			framework.Secrets = append(framework.Secrets, secret)
			framework.Labels["slash/test.dot"] = "application"
			Expect(framework.Deploy()).To(BeNil())

			// Wait for splunk to be ready
			waitOnSplunk()

			// Write app logs
			timestamp := "2020-11-04T18:13:59.061892+00:00"
			applicationLogLine := functional.NewCRIOLogMessage(timestamp, "This is my test message", false)
			Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 2)).To(BeNil())

			// Read app logs
			logs, err := framework.ReadAppLogsByIndexFromSplunk(framework.Namespace, framework.Name, expIndex)
			Expect(err).To(BeNil(), "Expected no errors getting logs from splunk")
			Expect(logs).ToNot(BeEmpty())

			// Parse the logs
			var appLogs []types.ApplicationLog
			jsonString := fmt.Sprintf("[%s]", strings.Join(logs, ","))
			err = types.ParseLogsFrom(jsonString, &appLogs, false)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")

			outputTestLog := appLogs[0]
			Expect(outputTestLog.LogType).To(Equal(string(obs.InputTypeApplication)))
		},
			Entry("should send logs to spec'd static index in Splunk", functional.SplunkIndexName, functional.SplunkIndexName),
			Entry("should send logs to spec'd dynamic index in Splunk", `{.log_type||"missing"}`, "application"),
			Entry("should send logs to spec'd static + dynamic index in Splunk", `foo-{.log_type||"missing"}`, "foo-application"),
			Entry("should send logs to spec'd static + label with dot/slash index in Splunk", `foo-{.kubernetes.labels."slash/test.dot"||"missing"}`, "foo-application"),
			Entry("should send logs to spec'd static + fallback value's index in Splunk if field is missing", `foo-{.missing||"application"}`, "foo-application"),
			Entry("should send logs to default index in Splunk when no index is defined", "", functional.SplunkDefaultIndex))
	})

	Context("tuning parameters", func() {
		DescribeTable("with compression settings", func(compression string) {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToSplunkOutput(hecSecretKey, func(output *obs.OutputSpec) {
					output.Splunk.Tuning = &obs.SplunkTuningSpec{
						Compression: compression,
					}
					output.Splunk.Index = "main"
				})
			framework.Secrets = append(framework.Secrets, secret)
			Expect(framework.Deploy()).To(BeNil())

			// Wait for splunk to be ready
			waitOnSplunk()

			// Write app logs
			timestamp := "2020-11-04T18:13:59.061892+00:00"
			applicationLogLine := functional.NewCRIOLogMessage(timestamp, "This is my test message", false)
			Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 2)).To(BeNil())

			// Read app logs
			logs, err := framework.ReadLogsByTypeFromSplunk(framework.Namespace, framework.Name, string(obs.InputTypeApplication))

			Expect(err).To(BeNil(), "expected no errors getting logs from splunk")
			Expect(logs).ToNot(BeEmpty())

			// Parse the logs
			var appLogs []types.ApplicationLog
			jsonString := fmt.Sprintf("[%s]", strings.Join(logs, ","))
			err = types.ParseLogsFrom(jsonString, &appLogs, false)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")
			outputTestLog := appLogs[0]
			Expect(outputTestLog.LogType).To(Equal(string(obs.InputTypeApplication)))

		},
			Entry("should pass with gzip", "gzip"),
			Entry("should pass with no compression", ""),
		)
	})

	Context("splunk metadata", func() {
		It("should accept indexed fields", func() {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToSplunkOutput(hecSecretKey, func(output *obs.OutputSpec) {
					output.Splunk.IndexedFields = []obs.FieldPath{`.log_type`, `.openshift.sequence`, `.kubernetes.annotations."openshift.io/scc"`}
				})
			framework.Secrets = append(framework.Secrets, secret)
			Expect(framework.Deploy()).To(BeNil())

			// Wait for splunk to be ready
			waitOnSplunk()

			// Write app logs
			timestamp := "2020-11-04T18:13:59.061892+00:00"
			applicationLogLine := functional.NewCRIOLogMessage(timestamp, "This is my test message", false)
			Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 2)).To(BeNil())

			// Read app logs
			logs, err := framework.ReadLogsByTypeFromSplunk(framework.Namespace, framework.Name, string(obs.InputTypeApplication))
			Expect(err).To(BeNil(), "Expected no errors getting logs from splunk")
			Expect(logs).ToNot(BeEmpty())

			// Parse the logs
			var appLogs []types.ApplicationLog
			jsonString := fmt.Sprintf("[%s]", strings.Join(logs, ","))
			Expect(strings.Count(jsonString, "openshift_sequence")).To(Equal(2))
			Expect(strings.Count(jsonString, "kubernetes_annotations_openshift_io_scc")).To(Equal(2))

			err = types.ParseLogsFrom(jsonString, &appLogs, false)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")

			outputTestLog := appLogs[0]
			Expect(outputTestLog.LogType).To(Equal(string(obs.InputTypeApplication)))
			Expect(outputTestLog.Openshift.Sequence).To(BeEmpty()) //.openshift.sequence must be removed because it was set as root-level field openshift_sequence
			Expect(outputTestLog.Openshift.ClusterID).To(Equal("functional"))
			_, exist := outputTestLog.Kubernetes.Annotations["openshift.io/scc"]
			Expect(exist).To(BeFalse()) // .kubernetes.annotations."openshift.io/scc" must be removed because was set as root-level field kubernetes_annotations_openshift_io_scc

			//checking Splunk index fields with 'index=${index} | stats count by ${field_name}'
			//not direct checking to be sure fields sent as indexed
			//if filed is indexed will return counts 2*logs_count or logs_count otherwise
			output, err := framework.ReadStatsForFieldByIndexFromSplunk(framework.Namespace, framework.Name, "main", "kubernetes_annotations_openshift_io_scc", "json")
			Expect(err).To(BeNil(), "Expected no errors getting field stats")
			Expect(strings.Contains(output, "kubernetes_annotations_openshift_io_scc"))

			Expect(strings.HasSuffix(output, `"result":{"kubernetes_annotations_openshift_io_scc":"node-exporter","count":"4"}}`)).To(BeTrue())

			output, err = framework.ReadStatsForFieldByIndexFromSplunk(framework.Namespace, framework.Name, "main", "log_type", "json")
			Expect(err).To(BeNil(), "Expected no errors getting field stats")
			Expect(strings.HasSuffix(output, `"result":{"log_type":"application","count":"4"}}`)).To(BeTrue())

			output, err = framework.ReadStatsForFieldByIndexFromSplunk(framework.Namespace, framework.Name, "main", "log_source", "json")
			Expect(err).To(BeNil(), "Expected no errors getting field stats")
			Expect(strings.HasSuffix(output, `"result":{"log_source":"container","count":"2"}}`)).To(BeTrue())

		})

		It("should send correct hostname", func() {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToSplunkOutput(hecSecretKey, func(output *obs.OutputSpec) {})
			framework.Secrets = append(framework.Secrets, secret)
			Expect(framework.Deploy()).To(BeNil())

			// Wait for splunk to be ready
			waitOnSplunk()

			// Write app logs
			timestamp := "2020-11-04T18:13:59.061892+00:00"
			applicationLogLine := functional.NewCRIOLogMessage(timestamp, "This is my test message", false)
			Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 2)).To(BeNil())

			// Read app logs
			logs, err := framework.ReadLogsByTypeFromSplunk(framework.Namespace, framework.Name, string(obs.InputTypeApplication))
			Expect(err).To(BeNil(), "Expected no errors getting logs from splunk")
			Expect(logs).ToNot(BeEmpty())

			// Parse the logs
			var appLogs []types.ApplicationLog
			jsonString := fmt.Sprintf("[%s]", strings.Join(logs, ","))

			err = types.ParseLogsFrom(jsonString, &appLogs, false)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")

			outputTestLog := appLogs[0]
			Expect(outputTestLog.LogType).To(Equal(string(obs.InputTypeApplication)))

			result, err := framework.ReadFieldByIndexFromSplunk(framework.Namespace, framework.Name, "main", "host", "json")
			Expect(err).To(BeNil(), "Expected no errors getting logs from splunk")

			for _, v := range result {
				matches := regexpHost.FindStringSubmatch(v)
				Expect(matches[1]).To(Equal(outputTestLog.Hostname), "Expected to find match for host")
			}

		})

		DescribeTable("with user defined source", func(source, expSource string, inputType obs.InputType) {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(inputType).
				ToSplunkOutput(hecSecretKey, func(output *obs.OutputSpec) {
					if source != "" {
						output.Splunk.Source = source
					}
				})
			framework.Secrets = append(framework.Secrets, secret)
			if inputType == obs.InputTypeApplication {
				framework.Labels["slash/test.dot"] = "application"
			}
			Expect(framework.Deploy()).To(BeNil())

			// Wait for splunk to be ready
			waitOnSplunk()

			if inputType == obs.InputTypeApplication {
				if expSource == "" {
					expSource = strings.Join([]string{framework.Namespace, framework.Pod.Name, "collector"}, "_")
				}
				// Write app logs
				timestamp := "2020-11-04T18:13:59.061892+00:00"
				applicationLogLine := functional.NewCRIOLogMessage(timestamp, "This is my test message", false)
				Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 2)).To(BeNil())
			}

			if inputType == obs.InputTypeAudit {
				if expSource == "" {
					expSource = "auditd"
				}
				// Write audit logs
				timestamp, _ := time.Parse(time.RFC3339Nano, "2024-04-16T09:46:19.116+00:00")
				auditLogLine := functional.NewAuditHostLog(timestamp)
				writeAuditLogs := framework.WriteMessagesToAuditLog(auditLogLine, 1)
				Expect(writeAuditLogs).To(BeNil(), "Expect no errors writing audit logs")
			}

			if inputType == obs.InputTypeInfrastructure {
				if expSource == "" {
					expSource = "google-chrome.desktop"
				}
				// Write journal logs
				logline := functional.NewJournalLog(0, "journal message", "functional")
				Expect(framework.WriteMessagesToInfraJournalLog(logline, 1)).To(BeNil())
			}

			// Read logs
			logs, err := framework.ReadLogsByTypeFromSplunk(framework.Namespace, framework.Name, string(inputType))
			Expect(err).To(BeNil(), "Expected no errors getting logs from splunk")
			Expect(logs).ToNot(BeEmpty())

			result, err := framework.ReadFieldByIndexFromSplunk(framework.Namespace, framework.Name, "main", "source", "json")
			Expect(err).To(BeNil(), "Expected no errors getting field from splunk")
			for _, v := range result {
				matches := regexpSource.FindStringSubmatch(v)
				Expect(matches[1]).To(Equal(expSource), "Expected to find match for source")
			}

		},
			Entry("should send application logs with dynamic source field", `{.log_type||"missing"}`, "application", obs.InputTypeApplication),
			Entry("should send application logs with static + dynamic source field", `foo-{.log_type||"missing"}`, "foo-application", obs.InputTypeApplication),
			Entry("should send application logs with static + label with dot/slash source field", `foo-{.kubernetes.labels."slash/test.dot"||"missing"}`, "foo-application", obs.InputTypeApplication),
			Entry("should send application logs with static + fallback value's source field", `foo-{.missing||"application"}`, "foo-application", obs.InputTypeApplication),
			Entry("should send application logs with default source ", "", "", obs.InputTypeApplication),
			Entry("should send audit logs with default source ", "", "", obs.InputTypeAudit),
			Entry("should send journal logs with default source ", "", "", obs.InputTypeInfrastructure))

		DescribeTable("with user defined payloadKey", func(payloadKey, expSourceType string, expression *regexp.Regexp) {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToSplunkOutput(hecSecretKey, func(output *obs.OutputSpec) {
					if payloadKey != "" {
						output.Splunk.PayloadKey = obs.FieldPath(payloadKey)
					}
				})
			framework.Secrets = append(framework.Secrets, secret)

			Expect(framework.Deploy()).To(BeNil())

			// Wait for splunk to be ready
			waitOnSplunk()

			// Write app logs
			timestamp := "2020-11-04T18:13:59.061892+00:00"
			applicationLogLine := functional.NewCRIOLogMessage(timestamp, "This is my test message", false)
			Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 2)).To(BeNil())

			// Read logs
			logs, err := framework.ReadAppLogsByIndexFromSplunk(framework.Namespace, framework.Name, "main")
			Expect(err).To(BeNil(), "Expected no errors getting logs from splunk")
			Expect(logs).ToNot(BeEmpty())
			Expect(len(logs)).To(Equal(2))

			if expression != nil {
				for _, v := range logs {
					Expect(expression.MatchString(v)).To(BeTrue())
				}
			} else {
				//Parse the logs
				var appLogs []types.ApplicationLog
				jsonString := fmt.Sprintf("[%s]", strings.Join(logs, ","))
				err = types.ParseLogsFrom(jsonString, &appLogs, false)
				Expect(err).To(BeNil(), "Expected no errors parsing the logs")
			}

			result, err := framework.ReadFieldByIndexFromSplunk(framework.Namespace, framework.Name, "main", "sourcetype", "json")
			Expect(err).To(BeNil(), "Expected no errors getting field from splunk")
			for _, v := range result {
				Expect(strings.HasSuffix(v, fmt.Sprintf(`"result":{"sourcetype":"%s"}}`, expSourceType))).To(BeTrue())
			}
		},
			Entry("should send only 'openshift' payload of application logs", ".openshift", "_json", regexOpenshift),
			Entry("should send only 'message' payload of application logs", ".message", "generic_single_line", regexMessage),
			Entry("should send full application log if payload not found", ".not_found", "_json", nil))
	})
})
