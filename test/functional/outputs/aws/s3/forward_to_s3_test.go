package s3

import (
	"fmt"
	"strings"
	"time"

	log "github.com/ViaQ/logerr/v2/log/static"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	obstestruntime "github.com/openshift/cluster-logging-operator/test/runtime/observability"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("[Functional][Outputs][S3] Forward Output to s3 (minIO)", func() {

	const (
		logSize        = 128
		numOfLogs      = 4
		TestBucketName = "functional-test-s3-bucket"
		TestRegion     = "us-east-2"
	)

	var (
		framework *functional.CollectorFunctionalFramework
		secret    *v1.Secret
		obsAuth   *obs.AwsAuthentication
	)

	// setupS3Output configures S3 output with batch timeout and secrets
	setupS3Output := func(inputType obs.InputType, compression, keyPrefix string) {
		obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(inputType).
			ToS3Output(*obsAuth, TestBucketName, keyPrefix, fmt.Sprintf("%d", functional.MinioPort), func(output *obs.OutputSpec) {
				output.S3.Region = TestRegion
				output.S3.Tuning = &obs.S3TuningSpec{
					Compression: compression,
				}
			})
		// Override the generated Vector config to set a short batch timeout.
		// Vector's default batch timeout for S3 is 300s which is too long for tests.
		// https://vector.dev/docs/reference/configuration/sinks/aws_s3/#batch.timeout_secs
		framework.VisitConfig = func(conf string) string {
			modifiedConf, err := functional.SetS3BatchTimeout(conf, 10)
			Expect(err).To(BeNil(), "Failed to set S3 batch timeout in Vector config")
			return modifiedConf
		}
		framework.Secrets = append(framework.Secrets, secret)
	}

	BeforeEach(func() {
		// s3 and minIO often require a long timeout
		framework = functional.NewCollectorFunctionalFramework()
		framework.MaxReadDuration = utils.GetPtr(time.Second * 60)

		log.V(2).Info("Creating secret s3-secret with example AWS creds for minIO")
		secret = runtime.NewSecret(framework.Namespace, functional.S3Secret,
			map[string][]byte{
				constants.AwsAccessKeyID:     []byte(functional.AwsAccessKeyID),
				constants.AwsSecretAccessKey: []byte(functional.AwsSecretAccessKey),
			},
		)

		// aws auth spec
		obsAuth = &obs.AwsAuthentication{
			Type: obs.AwsAuthTypeAccessKey,
			AwsAccessKey: &obs.AwsAccessKey{
				KeySecret: obs.SecretReference{
					Key:        constants.AwsSecretAccessKey,
					SecretName: functional.S3Secret,
				},
				KeyId: obs.SecretReference{
					Key:        constants.AwsAccessKeyID,
					SecretName: functional.S3Secret,
				},
			},
		}
	})

	AfterEach(func() {
		// Clean up S3 port-forward before general framework cleanup
		framework.CleanupS3PortForward()
		framework.Cleanup()
	})

	Context("When sending application log messages to s3 (minIO)", func() {
		It("should be able to read logs from the bucket and key-prefix", func() {
			const keyPrefix = "application/"

			setupS3Output(obs.InputTypeApplication, "none", keyPrefix)
			Expect(framework.Deploy()).To(BeNil())

			// create the bucket in minIO after deployment
			Expect(framework.SetupS3Bucket(TestBucketName)).
				To(Succeed(), "should set up the minIO test bucket")

			// write logs and wait for Vector to batch and upload to S3 (batch timeout = 10s)
			Expect(framework.WritesNApplicationLogsOfSize(numOfLogs, logSize, 0)).To(BeNil())
			time.Sleep(15 * time.Second)

			// read logs
			logs, err := framework.ReadLogsFromS3(TestBucketName, keyPrefix)
			Expect(err).To(BeNil(), "Expected no errors reading logs from s3")
			Expect(logs).To(HaveLen(numOfLogs), "Expected to find the correct number of logs in s3")
		})
	})

	Context("When sending audit log messages to s3 (minIO)", func() {
		It("should appear in s3 with the audit log_type key-prefix", func() {
			const keyPrefix = "audit/"
			const numLogsSent = 5

			setupS3Output(obs.InputTypeAudit, "none", keyPrefix)
			Expect(framework.Deploy()).To(BeNil())

			// create the bucket in minIO after deployment
			Expect(framework.SetupS3Bucket(TestBucketName)).
				To(Succeed(), "should set up the minIO test bucket")

			// write audit logs
			ts := functional.CRIOTime(time.Now().Add(-12 * time.Hour))
			tstamp, err := time.Parse(time.RFC3339Nano, ts)
			Expect(err).To(BeNil(), "Failed to parse CRIO timestamp")
			auditLogLine := functional.NewAuditHostLog(tstamp)
			writeAuditLogs := framework.WriteMessagesToAuditLog(auditLogLine, numLogsSent)
			Expect(writeAuditLogs).To(BeNil(), "Expect no errors writing audit logs")
			time.Sleep(15 * time.Second)

			// read logs
			logs, err := framework.ReadLogsFromS3(TestBucketName, keyPrefix)
			Expect(err).To(BeNil(), "Expected no errors reading audit logs from S3")
			Expect(logs).To(HaveLen(numLogsSent), "Expected to receive the correct number of audit log messages in S3")
		})
	})

	Context("When sending infrastructure log messages to s3 (minIO)", func() {
		It("should appear in s3 with the infrastructure log_type key-prefix", func() {
			const keyPrefix = "infrastructure/"
			const numLogsSent = 3

			setupS3Output(obs.InputTypeInfrastructure, "none", keyPrefix)
			Expect(framework.Deploy()).To(BeNil())

			// create the bucket in minIO after deployment
			Expect(framework.SetupS3Bucket(TestBucketName)).
				To(Succeed(), "should set up the minIO test bucket")

			// write infrastructure logs (journal logs)
			logline := functional.NewJournalLog(3, "*", "*")
			Expect(framework.WriteMessagesToInfraJournalLog(logline, numLogsSent)).
				To(BeNil(), "Expect no errors writing infrastructure logs")
			time.Sleep(15 * time.Second)

			// read logs
			logs, err := framework.ReadLogsFromS3(TestBucketName, keyPrefix)
			Expect(err).To(BeNil(), "Expected no errors reading infrastructure logs from S3")
			Expect(logs).To(HaveLen(numLogsSent), "Expected to receive the correct number of infrastructure log messages in S3")
		})
	})

	Context("When setting tuning parameters", func() {
		DescribeTable("should be able to read logs from the specified bucket and key-prefix", func(compression string) {
			const keyPrefix = "application/"
			setupS3Output(obs.InputTypeApplication, compression, keyPrefix)
			Expect(framework.Deploy()).To(BeNil())

			// create the bucket in minIO after deployment
			Expect(framework.SetupS3Bucket(TestBucketName)).
				To(Succeed(), "should set up the minIO test bucket")

			Expect(framework.WritesNApplicationLogsOfSize(numOfLogs, logSize, 0)).To(BeNil())
			time.Sleep(15 * time.Second)

			logs, err := framework.ReadLogsFromS3(TestBucketName, keyPrefix)

			Expect(err).To(BeNil(), "Expected no errors reading logs from s3")
			Expect(logs).To(HaveLen(numOfLogs), "Expected to find the correct number of logs in s3")
		},
			Entry("should pass with no compression", "none"),
			Entry("should pass with gzip", "gzip"),
			Entry("should pass with snappy", "snappy"),
			Entry("should pass with zlib", "zlib"),
			Entry("should pass with zstd", "zstd"),
		)
	})

	// Regression test for https://redhat.atlassian.net/browse/LOG-7893
	//
	// vector_component_sent_bytes_total must carry component_id, component_kind,
	// and component_type labels so dashboards can filter by component_kind="sink".
	// A bug in Vector's AwsBytesSent causes these labels to be lost when the
	// metric is emitted from tower::buffer worker tasks outside the component
	// tracing span.
	Context("When checking collector metrics for S3 output", func() {
		BeforeEach(func() {
			role, binding, tokenBinding, err := framework.SetupMetricsRBAC()
			Expect(err).To(Succeed())
			DeferCleanup(func() {
				_ = framework.Test.Delete(tokenBinding)
				_ = framework.Test.Delete(binding)
				_ = framework.Test.Delete(role)
			})
		})

		It("should emit component_sent_bytes_total with component_id and region labels, and no unlabeled duplicates (LOG-7893)", func() {
			const keyPrefix = "metrics-test/"

			setupS3Output(obs.InputTypeApplication, "none", keyPrefix)
			Expect(framework.Deploy()).To(BeNil())

			Expect(framework.SetupS3Bucket(TestBucketName)).
				To(Succeed(), "should set up the minIO test bucket")

			Expect(framework.WritesNApplicationLogsOfSize(numOfLogs, logSize, 0)).To(BeNil())
			time.Sleep(15 * time.Second)

			lines, err := framework.CollectMetricLines("component_sent_bytes_total", `component_id="output_s3"`, 30*time.Second)
			Expect(err).To(BeNil(), "Timed out waiting for component_sent_bytes_total metric")

			for _, line := range lines {
				log.V(2).Info("component_sent_bytes_total line", "line", line)
				Expect(line).To(ContainSubstring(`component_id=`),
					"component_sent_bytes_total without component_id label (transport-layer duplicate): %s", line)
				if strings.Contains(line, `component_id="output_s3"`) {
					Expect(line).To(ContainSubstring(`region=`),
						"component_sent_bytes_total for S3 output is missing region label: %s", line)
				}
			}
		})
	})
})
