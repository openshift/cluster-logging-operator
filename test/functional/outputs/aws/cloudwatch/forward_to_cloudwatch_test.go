package cloudwatch

import (
	"fmt"
	"strings"
	"time"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	obstestruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	log "github.com/ViaQ/logerr/v2/log/static"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("[Functional][Outputs][CloudWatch] Forward Output to CloudWatch", func() {

	const (
		logSize   = 128
		numOfLogs = 8
	)

	var (
		framework *functional.CollectorFunctionalFramework
		secret    *v1.Secret
		obsCwAuth *obs.AwsAuthentication
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()
		framework.MaxReadDuration = utils.GetPtr(time.Second * 45)

		log.V(2).Info("Creating secret cloudwatch with AWS example credentials")
		secret = runtime.NewSecret(framework.Namespace, functional.CloudwatchSecret,
			map[string][]byte{
				"aws_access_key_id":     []byte(functional.AwsAccessKeyID),
				"aws_secret_access_key": []byte(functional.AwsSecretAccessKey),
			},
		)

		obsCwAuth = &obs.AwsAuthentication{
			Type: obs.AwsAuthTypeAccessKey,
			AwsAccessKey: &obs.AwsAccessKey{
				KeySecret: obs.SecretReference{
					Key:        constants.AwsSecretAccessKey,
					SecretName: functional.CloudwatchSecret,
				},
				KeyId: obs.SecretReference{
					Key:        constants.AwsAccessKeyID,
					SecretName: functional.CloudwatchSecret,
				},
			},
		}
	})

	AfterEach(func() {
		framework.Cleanup()
	})

	Context("When sending a sequence of app log messages to CloudWatch", func() {
		It("should be able to read messages from application log group", func() {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToCloudwatchOutput(*obsCwAuth)
			framework.Secrets = append(framework.Secrets, secret)

			Expect(framework.Deploy()).To(BeNil())

			Expect(framework.WritesNApplicationLogsOfSize(numOfLogs, logSize, 0)).To(BeNil())
			time.Sleep(10 * time.Second)

			logs, err := framework.ReadLogsFromCloudwatch(string(obs.InputTypeApplication))
			Expect(err).To(BeNil())
			Expect(logs).To(HaveLen(numOfLogs))
		})

		Context("group_name", func() {
			DescribeTable("templated group_name", func(groupName, expGroupName string) {
				obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
					FromInput(obs.InputTypeApplication).
					ToCloudwatchOutput(*obsCwAuth, func(output *obs.OutputSpec) {
						output.Cloudwatch.GroupName = groupName
					})
				framework.Secrets = append(framework.Secrets, secret)

				Expect(framework.Deploy()).To(BeNil())

				Expect(framework.WritesNApplicationLogsOfSize(numOfLogs, logSize, 0)).To(BeNil())
				time.Sleep(10 * time.Second)

				logs, err := framework.ReadLogsFromCloudwatchByGroupName(expGroupName)
				Expect(err).To(BeNil())
				Expect(logs).To(HaveLen(numOfLogs))
			},
				Entry("should write to defined static tenant", "custom-index", "custom-index"),
				Entry("should write to defined dynamic tenant", `{.log_type||"none"}`, "application"),
				Entry("should write to defined static + dynamic tenant", `foo-{.log_type||"none"}`, "foo-application"),
				Entry("should write to defined static + fallback value if field is missing", `foo-{.missing||"none"}`, "foo-none"))
		})

		It("should be able to forward by user-defined pod labels", func() {
			var (
				labelKey   = "env1"
				labelValue = "test1"
			)
			// Add new label
			framework.Labels[labelKey] = labelValue

			// testruntime a pod spec with a selector for labels and namespace
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInputName("myinput",
					func(spec *obs.InputSpec) {
						spec.Type = obs.InputTypeApplication
						spec.Application = &obs.Application{
							Includes: []obs.NamespaceContainerSpec{
								{
									Namespace: framework.Namespace,
								},
							},
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{labelKey: labelValue},
							},
						}
					},
				).
				ToCloudwatchOutput(*obsCwAuth)
			framework.Secrets = append(framework.Secrets, secret)
			Expect(framework.Deploy()).To(BeNil())

			// Write logs
			Expect(framework.WritesApplicationLogs(numOfLogs)).To(BeNil())
			time.Sleep(10 * time.Second)

			logs, err := framework.ReadLogsFromCloudwatch(string(obs.InputTypeApplication))
			Expect(err).To(BeNil())
			Expect(logs).To(HaveLen(numOfLogs))
			Expect(logs[0]).Should(MatchRegexp(fmt.Sprintf(`{.*"%s":"%s".*}`, labelKey, labelValue)))
		})

		It("should reassemble multi-line stacktraces (e.g. LOG-2275)", func() {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInputName("myinput",
					func(spec *obs.InputSpec) {
						spec.Type = obs.InputTypeApplication
						spec.Application = &obs.Application{
							Includes: []obs.NamespaceContainerSpec{
								{
									Namespace: framework.Namespace,
								},
							},
						}
					}).
				WithMultilineErrorDetectionFilter().
				ToCloudwatchOutput(*obsCwAuth)
			framework.Secrets = append(framework.Secrets, secret)
			Expect(framework.Deploy()).To(BeNil())

			exception := `java.lang.NullPointerException: Cannot invoke "String.toString()" because "<parameter1>" is null
        at testjava.Main.printMe(Main.java:19)
        at testjava.Main.main(Main.java:10)`
			timestamp := functional.CRIOTime(time.Now())

			buffer := []string{}
			for _, line := range strings.Split(exception, "\n") {
				crioLine := functional.NewCRIOLogMessage(timestamp, line, false)
				buffer = append(buffer, crioLine)
			}

			Expect(framework.WriteMessagesToNamespace(strings.Join(buffer, "\n"), framework.Namespace, 1)).To(Succeed())
			time.Sleep(10 * time.Second)

			raw, err := framework.ReadLogsFromCloudwatch(string(obs.InputTypeApplication))
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			logs, err := types.ParseLogs(utils.ToJsonLogs(raw))
			Expect(err).To(BeNil(), "Expected no errors parsing the logs: %s", raw)
			Expect(logs[0].Message).Should(Equal(exception))
		})
	})

	Context("When sending audit log messages to CloudWatch", func() {
		var (
			numLogsSent = 2
			readLogType = obs.InputTypeAudit
		)
		It("should appear in the audit log group with audit log_type", func() {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeAudit).
				ToCloudwatchOutput(*obsCwAuth)
			framework.Secrets = append(framework.Secrets, secret)
			Expect(framework.Deploy()).To(BeNil())

			// Write audit logs
			ts := functional.CRIOTime(time.Now().Add(-12 * time.Hour)) // must be less than 14 days old
			tstamp, _ := time.Parse(time.RFC3339Nano, ts)
			auditLogLine := functional.NewAuditHostLog(tstamp)
			writeAuditLogs := framework.WriteMessagesToAuditLog(auditLogLine, numLogsSent)
			Expect(writeAuditLogs).To(BeNil(), "Expect no errors writing logs")
			time.Sleep(10 * time.Second)

			// Get audit logs from Cloudwatch
			logs, err := framework.ReadLogsFromCloudwatch(string(readLogType))
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(logs).To(HaveLen(numLogsSent), "Expected to receive the correct number of audit log messages")
			Expect(logs[0]).Should(MatchRegexp(fmt.Sprintf(`{.*"log_type":"%s".*}`, readLogType)))
		})
	})

	Context("When setting tuning parameters", func() {
		DescribeTable("with compression", func(compression string) {
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToCloudwatchOutput(*obsCwAuth, func(output *obs.OutputSpec) {
					output.Cloudwatch.Tuning = &obs.CloudwatchTuningSpec{
						Compression: compression,
					}
				})
			framework.Secrets = append(framework.Secrets, secret)

			Expect(framework.Deploy()).To(BeNil())

			Expect(framework.WritesNApplicationLogsOfSize(numOfLogs, logSize, 0)).To(BeNil())
			time.Sleep(10 * time.Second)

			logs, err := framework.ReadLogsFromCloudwatch(string(obs.InputTypeApplication))

			Expect(err).To(BeNil())
			Expect(logs).To(HaveLen(numOfLogs))

		},
			Entry("should pass with no compression", "none"),
			Entry("should pass with gzip", "gzip"),
			Entry("should pass with snappy", "snappy"),
			Entry("should pass with zlib", "zlib"),
			Entry("should pass with zstd", "zstd"),
		)
	})
})
