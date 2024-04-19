package cloudwatch

import (
	"fmt"
	testruntime "github.com/openshift/cluster-logging-operator/test/runtime"
	"strings"
	"time"

	log "github.com/ViaQ/logerr/v2/log/static"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testfw "github.com/openshift/cluster-logging-operator/test/functional"
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
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFrameworkUsingCollector(testfw.LogCollectionType)

		log.V(2).Info("Creating secret cloudwatch with AWS example credentials")
		secret = runtime.NewSecret(framework.Namespace, functional.CloudwatchSecret,
			map[string][]byte{
				"aws_access_key_id":     []byte(functional.AwsAccessKeyID),
				"aws_secret_access_key": []byte(functional.AwsSecretAccessKey),
			},
		)
	})

	AfterEach(func() {
		framework.Cleanup()
	})

	Context("When sending a sequence of app log messages to CloudWatch", func() {
		It("should be able to read messages from application log group", func() {
			testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToCloudwatchOutput()
			framework.Secrets = append(framework.Secrets, secret)

			Expect(framework.Deploy()).To(BeNil())

			Expect(framework.WritesNApplicationLogsOfSize(numOfLogs, logSize, 0)).To(BeNil())
			time.Sleep(10 * time.Second)

			logs, err := framework.ReadLogsFromCloudwatch(logging.InputNameApplication)
			Expect(err).To(BeNil())
			Expect(logs).To(HaveLen(numOfLogs))
		})

		It("should be able to forward by user-defined pod labels", func() {
			var (
				labelKey   = "env1"
				labelValue = "test1"
			)
			// Add new label
			framework.Labels[labelKey] = labelValue

			// testruntime a pod spec with a selector for labels and namespace
			testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInputWithVisitor("app-logs",
					func(spec *logging.InputSpec) {
						spec.Application = &logging.Application{
							Namespaces: []string{framework.Namespace},
							Selector: &logging.LabelSelector{
								MatchLabels: map[string]string{labelKey: labelValue},
							},
						}
					},
				).
				ToCloudwatchOutput()
			framework.Secrets = append(framework.Secrets, secret)
			Expect(framework.Deploy()).To(BeNil())

			// Write logs
			Expect(framework.WritesApplicationLogs(numOfLogs)).To(BeNil())
			time.Sleep(10 * time.Second)

			logs, err := framework.ReadLogsFromCloudwatch(logging.InputNameApplication)
			Expect(err).To(BeNil())
			Expect(logs).To(HaveLen(numOfLogs))
			Expect(logs[0]).Should(MatchRegexp(fmt.Sprintf(`{.*"%s":"%s".*}`, labelKey, labelValue)))
		})

		It("should reassemble multi-line stacktraces (e.g. LOG-2275)", func() {
			if testfw.LogCollectionType == logging.LogCollectionTypeVector {
				Skip("not supported for this vector release")
			}
			appNamespace := "multi-line-test"
			testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInputWithVisitor("applogs-one", func(spec *logging.InputSpec) {
					spec.Application = &logging.Application{
						Namespaces: []string{appNamespace},
					}
				}).
				WithMultineErrorDetection().
				ToCloudwatchOutput()
			framework.VisitConfig = functional.TestAPIAdapterConfigVisitor
			framework.Secrets = append(framework.Secrets, secret)
			Expect(framework.Deploy()).To(BeNil())

			exception := `java.lang.NullPointerException: Cannot invoke "String.toString()" because "<parameter1>" is null
        at testjava.Main.printMe(Main.java:19)
        at testjava.Main.main(Main.java:10)`
			timestamp := "2021-03-31T12:59:28.573159188+00:00"

			buffer := []string{}
			for _, line := range strings.Split(exception, "\n") {
				crioLine := functional.NewCRIOLogMessage(timestamp, line, false)
				buffer = append(buffer, crioLine)
			}

			Expect(framework.WriteMessagesToNamespace(strings.Join(buffer, "\n"), appNamespace, 1)).To(Succeed())
			time.Sleep(10 * time.Second)

			raw, err := framework.ReadLogsFromCloudwatch(logging.InputNameApplication)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			logs, err := types.ParseLogs(utils.ToJsonLogs(raw))
			Expect(err).To(BeNil(), "Expected no errors parsing the logs: %s", raw)
			Expect(logs[0].Message).Should(Equal(exception))
		})
	})

	Context("When sending infrastructure log messages to CloudWatch", func() {
		var (
			numLogsSent = 2
			readLogType = logging.InputNameApplication
		)
		It("should not appear in the application log_group (e.g. LOG-2455)", func() {
			// Test method fails for vector since our pod/container namespace will always
			// begin with  "test-", thus cluster infrastructure logs are never found.
			if testfw.LogCollectionType == logging.LogCollectionTypeVector {
				Skip("not a valid test for vector since we route by namespace")
			}
			testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToCloudwatchOutput().
				FromInput(logging.InputNameAudit).
				ToCloudwatchOutput().
				FromInput(logging.InputNameInfrastructure).
				ToCloudwatchOutput()
			framework.Secrets = append(framework.Secrets, secret)
			Expect(framework.Deploy()).To(BeNil())

			// Write audit logs
			tstamp, _ := time.Parse(time.RFC3339Nano, "2021-03-28T14:36:03.243000+00:00")
			auditLogLine := functional.NewAuditHostLog(tstamp)
			writeAuditLogs := framework.WriteMessagesToAuditLog(auditLogLine, 3)
			Expect(writeAuditLogs).To(BeNil(), "Expect no errors writing logs")

			// Use specific namespace from ticket LOG-2455
			infraNamespace := "openshift-authentication-operator"
			payload := functional.NewFullCRIOLogMessage(functional.CRIOTime(time.Now()), `{"index":1,"timestamp":1}`)
			writeTicketLogs := framework.WriteMessagesToNamespace(payload, infraNamespace, 5)
			Expect(writeTicketLogs).To(BeNil(), "Expect no errors writing logs")

			// Write other fake infra messages (namespace: "openshift-fake-infra")
			writeInfraLogs := framework.WriteMessagesToInfraContainerLog(payload, 5)
			Expect(writeInfraLogs).To(BeNil(), "Expect no errors writing logs")

			// Write a single app log just to be sure its picked up ("test-..." namespace)
			writeAppLogs := framework.WritesApplicationLogs(numLogsSent)
			Expect(writeAppLogs).To(BeNil(), "Expect no errors writing logs")
			time.Sleep(10 * time.Second)

			// Get application logs from Cloudwatch
			logs, err := framework.ReadLogsFromCloudwatch(readLogType)
			log.V(2).Info("ReadLogsFromCloudwatch", "logType", readLogType, "logs", logs, "err", err)

			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(logs).To(HaveLen(numLogsSent), "Expected the receiver to receive only the app log messages")
			Expect(logs[0]).Should(MatchRegexp(fmt.Sprintf(`{.*"log_type":"%s".*}`, readLogType)), "Expected log_type to be correct")
		})
	})

	Context("When sending audit log messages to CloudWatch", func() {
		var (
			numLogsSent = 2
			readLogType = logging.InputNameAudit
		)
		It("should appear in the audit log group with audit log_type", func() {
			testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameAudit).
				ToCloudwatchOutput()
			framework.Secrets = append(framework.Secrets, secret)
			Expect(framework.Deploy()).To(BeNil())

			// Write audit logs
			tstamp, _ := time.Parse(time.RFC3339Nano, "2021-03-28T14:36:03.243000+00:00")
			auditLogLine := functional.NewAuditHostLog(tstamp)
			writeAuditLogs := framework.WriteMessagesToAuditLog(auditLogLine, numLogsSent)
			Expect(writeAuditLogs).To(BeNil(), "Expect no errors writing logs")
			time.Sleep(10 * time.Second)

			// Get audit logs from Cloudwatch
			logs, err := framework.ReadLogsFromCloudwatch(readLogType)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(logs).To(HaveLen(numLogsSent), "Expected to receive the correct number of audit log messages")
			Expect(logs[0]).Should(MatchRegexp(fmt.Sprintf(`{.*"log_type":"%s".*}`, readLogType)))
		})
	})

	Context("When setting tuning parameters", func() {
		var (
			compVisitFunc func(spec *logging.OutputSpec)
		)
		DescribeTable("with compression", func(compression string) {
			compVisitFunc = func(spec *logging.OutputSpec) {
				spec.Tuning = &logging.OutputTuningSpec{
					Compression: compression,
				}
			}
			testruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToOutputWithVisitor(compVisitFunc, logging.OutputTypeCloudwatch)
			framework.Secrets = append(framework.Secrets, secret)

			Expect(framework.Deploy()).To(BeNil())

			Expect(framework.WritesNApplicationLogsOfSize(numOfLogs, logSize, 0)).To(BeNil())
			time.Sleep(10 * time.Second)

			logs, err := framework.ReadLogsFromCloudwatch(logging.InputNameApplication)

			Expect(err).To(BeNil())
			Expect(logs).To(HaveLen(numOfLogs))

		},
			Entry("should pass with no compression", ""),
			Entry("should pass with gzip", "gzip"),
			Entry("should pass with snappy", "snappy"),
			Entry("should pass with zlib", "zlib"),
			Entry("should pass with zstd", "zstd"),
		)
	})
})
