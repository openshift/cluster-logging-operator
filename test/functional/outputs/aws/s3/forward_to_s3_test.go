package s3

/*
var _ = Describe("[Functional][Outputs][S3] Forward Output to s3 (minIO)", func() {

	const (
		logSize        = 128
		numOfLogs      = 4
		TestBucketName = "functional-test-s3-bucket"
		TestRegion     = "us-east-1"
		minioPort      = 9000
	)

	var (
		framework *functional.CollectorFunctionalFramework
		secret    *v1.Secret
		obsAuth   *obs.AwsAuthentication
	)

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

		// create the bucket in minIO first
		Expect(framework.SetupS3Bucket(TestBucketName)).
			To(Succeed(), "should set up the minIO test bucket")
	})

	AfterEach(func() {
		framework.Cleanup()
	})

	Context("When sending application log messages to s3 (minIO)", func() {
		It("should be able to read logs from the bucket and key-prefix", func() {
			const keyPrefix = "application/"

			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToS3Output(*obsAuth, TestBucketName, keyPrefix, minioPort, func(output *obs.OutputSpec) {
					output.S3.Region = TestRegion
				})
			framework.Secrets = append(framework.Secrets, secret)
			Expect(framework.Deploy()).To(BeNil())

			// write logs and give the service a few seconds to process
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

			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeAudit).
				ToS3Output(*obsAuth, TestBucketName, keyPrefix, minioPort, func(output *obs.OutputSpec) {
					output.S3.Region = TestRegion
				})

			framework.Secrets = append(framework.Secrets, secret)
			Expect(framework.Deploy()).To(BeNil())

			// write audit logs, then wait 15s
			ts := functional.CRIOTime(time.Now().Add(-12 * time.Hour))
			tstamp, _ := time.Parse(time.RFC3339Nano, ts)
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

	Context("When setting tuning parameters", func() {
		DescribeTable("should be able to read logs from the specified bucket and key-prefix", func(compression string) {
			const keyPrefix = "application/"
			obstestruntime.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(obs.InputTypeApplication).
				ToS3Output(*obsAuth, TestBucketName, keyPrefix, minioPort, func(output *obs.OutputSpec) {
					output.S3.Region = TestRegion
					output.S3.Tuning = &obs.S3TuningSpec{
						Compression: compression,
					}
				})
			framework.Secrets = append(framework.Secrets, secret)
			Expect(framework.Deploy()).To(BeNil())

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
})
*/
