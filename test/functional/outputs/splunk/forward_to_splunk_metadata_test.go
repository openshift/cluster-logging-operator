package splunk

import (
	"fmt"
	"regexp"
	"time"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/test/helpers/splunk"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	obstestruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"

	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("Forwarding to Splunk with Metadata", func() {
	var (
		framework      *functional.CollectorFunctionalFramework
		secret         *v1.Secret
		hecSecretKey   = *internalobs.NewSecretReference(constants.SplunkHECTokenKey, SplunkSecretName)
		regexpSource   = regexp.MustCompile(`"source"\s*:\s*"([^"]+)"`)
		regexpHost     = regexp.MustCompile(`"host"\s*:\s*"([^"]+)"`)
		regexOpenshift = regexp.MustCompile(`{"openshift":{"cluster_id":".*","sequence":[0-9]+}}`)
		regexMessage   = regexp.MustCompile(`{"message":".*"}`)
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		secret = runtime.NewSecret(framework.Namespace, SplunkSecretName,
			map[string][]byte{
				"hecToken": functional.HecToken,
			},
		)
	})

	AfterEach(func() {
		framework.Cleanup()
	})

	Context("splunk metadata", func() {
		It("should accept indexed fields for Application log type", func() {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToSplunkOutput(hecSecretKey, func(output *obs.OutputSpec) {
					output.Splunk.IndexedFields = []obs.FieldPath{`.log_type`, `.openshift.sequence`, `.kubernetes.annotations."openshift.io/scc"`}
				})
			framework.Secrets = append(framework.Secrets, secret)
			Expect(framework.Deploy()).To(BeNil())

			// Wait for splunk to be ready
			splunk.WaitOnSplunk(framework)

			// Write app logs
			timestamp := "2020-11-04T18:13:59.061892+00:00"
			applicationLogLine := functional.NewCRIOLogMessage(timestamp, "This is my test message", false)
			Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 2)).To(BeNil())

			// Read app logs
			logs, err := framework.ReadLogsByTypeFromSplunk(string(obs.InputTypeApplication))
			Expect(err).To(BeNil(), "Expected no errors getting logs from splunk")
			Expect(logs).ToNot(BeEmpty())

			// Parse the logs
			var appLogs []types.ApplicationLog
			jsonString := fmt.Sprintf("[%s]", strings.Join(logs, ","))
			Expect(strings.Count(jsonString, "openshift_sequence")).To(Equal(2))
			scc := "kubernetes_annotations_openshift_io_scc"
			Expect(strings.Count(jsonString, scc)).To(Equal(2))

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
			//if field is indexed will return counts 2*logs_count or logs_count otherwise
			output, err := framework.ReadStatsForFieldByIndexFromSplunk(functional.SplunkDefaultIndex, scc, "json")
			Expect(err).To(BeNil(), "Expected no errors getting field stats")
			Expect(checkFieldCount(output, scc, 4)).To(BeTrue())

			output, err = framework.ReadStatsForFieldByIndexFromSplunk(functional.SplunkDefaultIndex, "log_type", "json")
			Expect(err).To(BeNil(), "Expected no errors getting field stats")
			Expect(checkFieldCount(output, "log_type", 4)).To(BeTrue())

			output, err = framework.ReadStatsForFieldByIndexFromSplunk(functional.SplunkDefaultIndex, "log_source", "json")
			Expect(err).To(BeNil(), "Expected no errors getting field stats")
			Expect(checkFieldCount(output, "log_source", 2)).To(BeTrue())
		})

		It("should accept indexed fields for Audit log type", func() {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeAudit).
				ToSplunkOutput(hecSecretKey, func(output *obs.OutputSpec) {
					output.Splunk.IndexedFields = []obs.FieldPath{`.log_type`, `.auditID`, `.annotations."authorization.k8s.io/decision"`, `.annotations."authorization.k8s.io/reason"`}
				})
			framework.Secrets = append(framework.Secrets, secret)
			Expect(framework.Deploy()).To(BeNil())

			// Wait for splunk to be ready
			splunk.WaitOnSplunk(framework)

			// Write audit logs
			writeAuditLogs := framework.WriteK8sAuditLog(2)
			Expect(writeAuditLogs).To(BeNil(), "Expect no errors writing audit logs")

			// Read app logs
			logs, err := framework.ReadLogsByTypeFromSplunk(string(obs.InputTypeAudit))
			Expect(err).To(BeNil(), "Expected no errors getting logs from splunk")
			Expect(logs).ToNot(BeEmpty())

			decision := "annotations_authorization_k8s_io_decision"
			reason := "annotations_authorization_k8s_io_reason"

			// Parse the logs
			var appLogs []types.AuditLog
			jsonString := fmt.Sprintf("[%s]", strings.Join(logs, ","))
			Expect(strings.Count(jsonString, decision)).To(Equal(2))
			Expect(strings.Count(jsonString, reason)).To(Equal(2))

			err = types.ParseLogsFrom(jsonString, &appLogs, false)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")

			outputTestLog := appLogs[0]
			Expect(outputTestLog.LogType).To(Equal(string(obs.InputTypeAudit)))
			Expect(outputTestLog.AuditID).ToNot(BeEmpty())
			Expect(outputTestLog.Annotations.AuthorizationK8SIoDecision).To(BeEmpty()) // must be removed because was set as root-level field annotations_authorization_k8s_io_decision
			Expect(outputTestLog.Annotations.AuthorizationK8SIoReason).To(BeEmpty())

			//checking Splunk index fields with 'index=${index} | stats count by ${field_name}'
			//not direct checking to be sure fields sent as indexed
			//if field is indexed will return counts 2*logs_count or logs_count otherwise

			output, err := framework.ReadStatsForFieldByIndexFromSplunk(functional.SplunkDefaultIndex, decision, "json")
			Expect(err).To(BeNil(), "Expected no errors getting field stats")
			Expect(checkFieldCount(output, decision, 4)).To(BeTrue())

			output, err = framework.ReadStatsForFieldByIndexFromSplunk(functional.SplunkDefaultIndex, reason, "json")
			Expect(err).To(BeNil(), "Expected no errors getting field stats")
			Expect(checkFieldCount(output, reason, 4)).To(BeTrue())

			output, err = framework.ReadStatsForFieldByIndexFromSplunk(functional.SplunkDefaultIndex, "auditID", "json")
			Expect(err).To(BeNil(), "Expected no errors getting field stats")
			Expect(checkFieldCount(output, "auditID", 4)).To(BeTrue())

			output, err = framework.ReadStatsForFieldByIndexFromSplunk(functional.SplunkDefaultIndex, "log_type", "json")
			Expect(err).To(BeNil(), "Expected no errors getting field stats")
			Expect(checkFieldCount(output, "log_type", 4)).To(BeTrue())

			output, err = framework.ReadStatsForFieldByIndexFromSplunk(functional.SplunkDefaultIndex, "log_source", "json")
			Expect(err).To(BeNil(), "Expected no errors getting field stats")
			Expect(checkFieldCount(output, "log_source", 2)).To(BeTrue())

		})

		It("should send correct hostname", func() {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToSplunkOutput(hecSecretKey, func(output *obs.OutputSpec) {})
			framework.Secrets = append(framework.Secrets, secret)
			Expect(framework.Deploy()).To(BeNil())

			// Wait for splunk to be ready
			splunk.WaitOnSplunk(framework)

			// Write app logs
			timestamp := "2020-11-04T18:13:59.061892+00:00"
			applicationLogLine := functional.NewCRIOLogMessage(timestamp, "This is my test message", false)
			Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 2)).To(BeNil())

			// Read app logs
			logs, err := framework.ReadLogsByTypeFromSplunk(string(obs.InputTypeApplication))
			Expect(err).To(BeNil(), "Expected no errors getting logs from splunk")
			Expect(logs).ToNot(BeEmpty())

			// Parse the logs
			var appLogs []types.ApplicationLog
			jsonString := fmt.Sprintf("[%s]", strings.Join(logs, ","))

			err = types.ParseLogsFrom(jsonString, &appLogs, false)
			Expect(err).To(BeNil(), "Expected no errors parsing the logs")

			outputTestLog := appLogs[0]
			Expect(outputTestLog.LogType).To(Equal(string(obs.InputTypeApplication)))

			result, err := framework.ReadFieldByIndexFromSplunk(functional.SplunkDefaultIndex, "host", "json")
			Expect(err).To(BeNil(), "Expected no errors getting logs from splunk")

			for _, v := range result {
				matches := regexpHost.FindStringSubmatch(v)
				Expect(matches[1]).To(Equal(outputTestLog.Hostname), "Expected to find match for host")
			}
		})

		It("should send correct hostname with payloadKey settings", func() {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToSplunkOutput(hecSecretKey, func(output *obs.OutputSpec) {
					output.Splunk.PayloadKey = ".message"
				})
			framework.VisitConfig = func(conf string) string {
				return strings.Replace(conf, `._internal.hostname = get_env_var("VECTOR_SELF_NODE_NAME") ?? ""`,
					`._internal.hostname = "acme.com"`, 1) // replace hostname for testing
			}
			framework.Secrets = append(framework.Secrets, secret)
			Expect(framework.Deploy()).To(BeNil())

			// Wait for splunk to be ready
			splunk.WaitOnSplunk(framework)

			// Write app logs
			timestamp := "2020-11-04T18:13:59.061892+00:00"
			applicationLogLine := functional.NewCRIOLogMessage(timestamp, "This is my test message", false)
			Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 1)).To(BeNil())

			Eventually(func() bool {
				result, err := framework.ReadAppLogsByIndexFromSplunk(functional.SplunkDefaultIndex)
				Expect(err).ToNot(HaveOccurred())
				return strings.Contains(strings.Join(result, ""), `{"message":"This is my test message"}`)
			}, 5*time.Second, 500*time.Millisecond).Should(BeTrue())

			result, err := framework.ReadFieldByIndexFromSplunk(functional.SplunkDefaultIndex, "host", "json")
			Expect(err).To(BeNil(), "Expected no errors getting logs from splunk")
			Expect(len(result)).To(BeEquivalentTo(1))
			Expect(result[0]).To(ContainSubstring(`{"host":"acme.com"}`), "Expected to find match for host")
			colLog, err := framework.ReadCollectorLogs()
			Expect(err).To(BeNil(), "Expected no errors getting collectors log")
			Expect(colLog).ToNot(ContainSubstring("Timestamp was not found. Deferring to Splunk to set the timestamp"))
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
			splunk.WaitOnSplunk(framework)

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
			logs, err := framework.ReadLogsByTypeFromSplunk(string(inputType))
			Expect(err).To(BeNil(), "Expected no errors getting logs from splunk")
			Expect(logs).ToNot(BeEmpty())

			result, err := framework.ReadFieldByIndexFromSplunk(functional.SplunkDefaultIndex, "source", "json")
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
			splunk.WaitOnSplunk(framework)

			// Write app logs
			timestamp := "2020-11-04T18:13:59.061892+00:00"
			applicationLogLine := functional.NewCRIOLogMessage(timestamp, "This is my test message", false)
			Expect(framework.WriteMessagesToApplicationLog(applicationLogLine, 2)).To(BeNil())

			// Read logs
			logs, err := framework.ReadAppLogsByIndexFromSplunk(functional.SplunkDefaultIndex)
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

			result, err := framework.ReadFieldByIndexFromSplunk(functional.SplunkDefaultIndex, "sourcetype", "json")
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

func checkFieldCount(output, field string, count int) bool {
	pattern := `"result":\{"%s":"(.*?)","count":"%d"\}\}$`
	expr := fmt.Sprintf(pattern, field, count)
	re := regexp.MustCompile(expr)
	return re.MatchString(output)
}
