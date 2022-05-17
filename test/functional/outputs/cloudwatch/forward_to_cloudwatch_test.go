package cloudwatch

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ViaQ/logerr/v2/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	openshiftv1 "github.com/openshift/api/route/v1"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	testfw "github.com/openshift/cluster-logging-operator/test/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	v1 "k8s.io/api/core/v1"

	"github.com/openshift/cluster-logging-operator/internal/constants"
)

var _ = Describe("[Functional][Outputs][CloudWatch] Forward Output to CloudWatch", func() {

	const (
		logSize   = 128
		numOfLogs = 8

		// examples from AWS docs
		awsAccessKeyID      = "AKIAIOSFODNN7EXAMPLE"
		awsSecretAccessKey  = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" //nolint:gosec
		cloudwatchMotoImage = "quay.io/openshift-logging/moto:2.2.3.dev0"
	)

	var (
		framework               *functional.CollectorFunctionalFramework
		logger                  = log.NewLogger("forward-to-cloudwatch-testing")
		addMotoContainerVisitor = func(b *runtime.PodBuilder) error {
			logger.V(2).Info("Adding AWS CloudWatch Logs mock container")
			b.AddContainer(logging.OutputTypeCloudwatch, cloudwatchMotoImage).
				WithCmdArgs([]string{"-s"}).
				End()
			return nil
		}

		mountCloudwatchSecretVisitor = func(b *runtime.PodBuilder) error {
			logger.V(2).Info("Mounting cloudwatch secret to the collector container")
			b.AddSecretVolume("cloudwatch", "cloudwatch").
				GetContainer(constants.CollectorName).
				AddVolumeMount("cloudwatch", "/var/run/ocp-collector/secrets/cloudwatch", "", true)
			return nil
		}

		cwlClient *cloudwatchlogs.Client
		service   *v1.Service
		route     *openshiftv1.Route
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFrameworkUsingCollector(testfw.LogCollectionType)

		logger.V(2).Info("Creating service moto-server")
		service = runtime.NewService(framework.Namespace, "moto-server")
		runtime.NewServiceBuilder(service).
			AddServicePort(5000, 5000).
			WithSelector(map[string]string{"testname": "functional"})
		if err := framework.Test.Client.Create(service); err != nil {
			panic(err)
		}

		logger.V(2).Info("Creating route moto-server")
		route = runtime.NewRoute(framework.Namespace, "moto-server", "moto-server", "5000")
		route.Spec.TLS = &openshiftv1.TLSConfig{
			Termination:                   openshiftv1.TLSTerminationPassthrough,
			InsecureEdgeTerminationPolicy: openshiftv1.InsecureEdgeTerminationPolicyNone,
		}
		if err := framework.Test.Client.Create(route); err != nil {
			panic(err)
		}

		cwlClient = cloudwatchlogs.NewFromConfig(
			aws.Config{
				Region: "us-east-1",
				Credentials: aws.CredentialsProviderFunc(func(context.Context) (aws.Credentials, error) {
					return aws.Credentials{
						AccessKeyID:     awsAccessKeyID,
						SecretAccessKey: awsSecretAccessKey,
					}, nil
				}),
				HTTPClient: &http.Client{
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{
							InsecureSkipVerify: true, //nolint:gosec
						},
					},
				},
				EndpointResolver: aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
					return aws.Endpoint{
						PartitionID:   "aws",
						URL:           "https://" + route.Spec.Host,
						SigningRegion: "us-east-1",
					}, nil
				}),
			})

		logger.V(2).Info("Creating secret cloudwatch with AWS example credentials")
		secret := runtime.NewSecret(framework.Namespace, "cloudwatch",
			map[string][]byte{
				"aws_access_key_id":     []byte(awsAccessKeyID),
				"aws_secret_access_key": []byte(awsSecretAccessKey),
			},
		)
		if err := framework.Test.Client.Create(secret); err != nil {
			panic(err)
		}

	})

	AfterEach(func() {
		framework.Cleanup()
	})

	Context("When sending a sequence of app log messages to CloudWatch", func() {
		It("should be able to read messages from application log group", func() {
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToCloudwatchOutput()

			Expect(framework.DeployWithVisitors([]runtime.PodBuilderVisitor{
				addMotoContainerVisitor,
				mountCloudwatchSecretVisitor,
			})).To(BeNil())

			Expect(framework.WritesNApplicationLogsOfSize(numOfLogs, logSize)).To(BeNil())

			logs, err := framework.ReadLogsFromCloudwatch(cwlClient, logging.InputNameApplication)
			Expect(err).To(BeNil())
			Expect(logs).To(HaveLen(numOfLogs))
			Expect(len(logs)).To(Equal(numOfLogs))

		})

		It("should reassemble multi-line stacktraces (e.g. LOG-2275)", func() {
			if testfw.LogCollectionType == logging.LogCollectionTypeVector {
				Skip("not supported for this vector release")
			}
			appNamespace := "multi-line-test"
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInputWithVisitor("applogs-one", func(spec *logging.InputSpec) {
					spec.Application = &logging.Application{
						Namespaces: []string{appNamespace},
					}
				}).
				WithMultineErrorDetection().
				ToCloudwatchOutput()
			framework.VisitConfig = functional.TestAPIAdapterConfigVisitor

			Expect(framework.DeployWithVisitors([]runtime.PodBuilderVisitor{
				addMotoContainerVisitor,
				mountCloudwatchSecretVisitor,
			})).To(BeNil())

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

			raw, err := framework.ReadLogsFromCloudwatch(cwlClient, logging.InputNameApplication)
			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			logs, err := types.ParseLogs(utils.ToJsonLogs(raw))
			Expect(err).To(BeNil(), "Expected no errors parsing the logs: %s", raw)
			Expect(logs[0].Message).To(Equal(exception))
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
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToCloudwatchOutput().
				FromInput(logging.InputNameAudit).
				ToCloudwatchOutput().
				FromInput(logging.InputNameInfrastructure).
				ToCloudwatchOutput()

			Expect(framework.DeployWithVisitors([]runtime.PodBuilderVisitor{
				addMotoContainerVisitor,
				mountCloudwatchSecretVisitor,
			})).To(BeNil())

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

			// Get application logs from Cloudwatch
			logs, err := framework.ReadLogsFromCloudwatch(cwlClient, readLogType)
			logger.V(2).Info("ReadLogsFromCloudwatch", "logType", readLogType, "logs", logs, "err", err)

			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(logs).To(HaveLen(numLogsSent), "Expected the receiver to receive only the app log messages")
			expMatch := fmt.Sprintf(`{.*"log_type":"%s".*}`, readLogType)
			Expect(logs[0]).To(MatchRegexp(expMatch), "Expected log_type to be correct")
		})
	})

	Context("When sending audit log messages to CloudWatch", func() {
		var (
			numLogsSent = 2
			readLogType = logging.InputNameAudit
		)
		It("should appear in the audit log group with audit log_type", func() {
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameAudit).
				ToCloudwatchOutput()

			Expect(framework.DeployWithVisitors([]runtime.PodBuilderVisitor{
				addMotoContainerVisitor,
				mountCloudwatchSecretVisitor,
			})).To(BeNil())

			// Write audit logs
			tstamp, _ := time.Parse(time.RFC3339Nano, "2021-03-28T14:36:03.243000+00:00")
			auditLogLine := functional.NewAuditHostLog(tstamp)
			writeAuditLogs := framework.WriteMessagesToAuditLog(auditLogLine, numLogsSent)
			Expect(writeAuditLogs).To(BeNil(), "Expect no errors writing logs")

			// Get audit logs from Cloudwatch
			logs, err := framework.ReadLogsFromCloudwatch(cwlClient, readLogType)
			logger.V(2).Info("GetLogGroupByType", "logType", readLogType, "logs", logs, "err", err)

			Expect(err).To(BeNil(), "Expected no errors reading the logs")
			Expect(logs).To(HaveLen(numLogsSent), "Expected to receive the correct number of audit log messages")
			expMatch := fmt.Sprintf(`{.*"log_type":"%s".*}`, readLogType)
			Expect(logs[0]).To(MatchRegexp(expMatch), "Expected log_type to be correct")
		})
	})
})
