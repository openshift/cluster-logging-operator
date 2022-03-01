package cloudwatch

import (
	"context"
	"crypto/tls"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/framework/functional"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	v1 "k8s.io/api/core/v1"
	"net/http"
	"strings"
	"time"

	"github.com/ViaQ/logerr/log"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/aws/aws-sdk-go-v2/aws"
	cwl "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"

	openshiftv1 "github.com/openshift/api/route/v1"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"

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
		framework *functional.CollectorFunctionalFramework

		addMotoContainerVisitor = func(b *runtime.PodBuilder) error {
			log.V(2).Info("Adding AWS CloudWatch Logs mock container")
			b.AddContainer(logging.OutputTypeCloudwatch, cloudwatchMotoImage).
				WithCmdArgs([]string{"-s"}).
				End()
			return nil
		}

		mountCloudwatchSecretVisitor = func(b *runtime.PodBuilder) error {
			log.V(2).Info("Mounting cloudwatch secret to the collector container")
			b.AddSecretVolume("cloudwatch", "cloudwatch").
				GetContainer(constants.CollectorName).
				AddVolumeMount("cloudwatch", "/var/run/ocp-collector/secrets/cloudwatch", "", true)
			return nil
		}

		cwlClient               *cwl.Client
		service                 *v1.Service
		route                   *openshiftv1.Route
		defaultRetryInterval, _ = time.ParseDuration("10s")

		getRawCloudwatchApplicationLogs = func(f *functional.CollectorFunctionalFramework, cwlClient *cwl.Client) ([]string, error) {
			log.V(2).Info("Retrieving cloudwatch LogGroups")
			var logGroupsOutput *cwl.DescribeLogGroupsOutput
			err := wait.PollImmediate(defaultRetryInterval, f.GetMaxReadDuration(), func() (done bool, err error) {
				logGroupsOutput, err = cwlClient.DescribeLogGroups(context.TODO(), nil)
				if err != nil || len(logGroupsOutput.LogGroups) == 0 {
					return false, err
				}
				return true, nil
			})

			if err != nil {
				return nil, err
			}
			log.V(2).Info("Results", "logGroups", logGroupsOutput.LogGroups)

			log.V(2).Info("Retrieving cloudwatch LogStreams")
			var logStreamsOutput *cwl.DescribeLogStreamsOutput
			err = wait.PollImmediate(defaultRetryInterval, f.GetMaxReadDuration(), func() (done bool, err error) {
				logStreamsOutput, err = cwlClient.DescribeLogStreams(
					context.TODO(),
					&cwl.DescribeLogStreamsInput{LogGroupName: logGroupsOutput.LogGroups[0].LogGroupName},
				)
				if err != nil || len(logStreamsOutput.LogStreams) == 0 {
					return false, err
				}
				return true, nil
			})

			if err != nil {
				return nil, err
			}
			log.V(2).Info("Results", "logStreams", logStreamsOutput.LogStreams)

			log.V(2).Info("Retrieving cloudwatch LogEvents")
			var logEventsOutput *cwl.GetLogEventsOutput
			err = wait.PollImmediate(defaultRetryInterval, f.GetMaxReadDuration(), func() (done bool, err error) {
				logEventsOutput, err = cwlClient.GetLogEvents(
					context.TODO(),
					&cwl.GetLogEventsInput{
						LogGroupName:  logGroupsOutput.LogGroups[0].LogGroupName,
						LogStreamName: logStreamsOutput.LogStreams[0].LogStreamName,
					})
				if err != nil || len(logEventsOutput.Events) == 0 {
					return false, err
				}
				return true, nil
			})
			if err != nil {
				return nil, err
			}
			log.V(3).Info("Results", "LogEvents", logEventsOutput.Events)
			buffer := []string{}
			for i := 0; i < len(logEventsOutput.Events); i++ {
				buffer = append(buffer, *logEventsOutput.Events[i].Message)
			}
			return buffer, nil
		}
	)

	BeforeEach(func() {
		framework = functional.NewCollectorFunctionalFramework()

		log.V(2).Info("Creating service moto")
		service = runtime.NewService(framework.Namespace, "moto")
		runtime.NewServiceBuilder(service).
			AddServicePort(5000, 5000).
			WithSelector(map[string]string{"testname": "functional"})
		if err := framework.Test.Client.Create(service); err != nil {
			panic(err)
		}
		route = runtime.NewRoute(framework.Namespace, "moto", "moto", "5000")
		route.Spec.TLS = &openshiftv1.TLSConfig{
			Termination:                   openshiftv1.TLSTerminationPassthrough,
			InsecureEdgeTerminationPolicy: openshiftv1.InsecureEdgeTerminationPolicyNone,
		}
		if err := framework.Test.Client.Create(route); err != nil {
			panic(err)
		}

		cwlClient = cwl.NewFromConfig(
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

		log.V(0).Info("Creating secret cloudwatch with AWS example credentials")
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

	Context("When sending a sequence of log messages to CloudWatch", func() {
		It("should get there", func() {
			functional.NewClusterLogForwarderBuilder(framework.Forwarder).
				FromInput(logging.InputNameApplication).
				ToCloudwatchOutput()

			Expect(framework.DeployWithVisitors([]runtime.PodBuilderVisitor{
				addMotoContainerVisitor,
				mountCloudwatchSecretVisitor,
			})).To(BeNil())

			Expect(framework.WritesNApplicationLogsOfSize(numOfLogs, logSize)).To(BeNil())

			logs, err := getRawCloudwatchApplicationLogs(framework, cwlClient)
			Expect(err).To(BeNil())
			Expect(logs).To(HaveLen(numOfLogs))
			Expect(len(logs)).To(Equal(numOfLogs))

		})
	})

	It("should reassemble multi-line stacktraces (e.g. LOG-2275)", func() {
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

		Expect(framework.WriteMessagesToApplicationLogInNamespace(strings.Join(buffer, "\n"), appNamespace, 1)).To(Succeed())

		raw, err := getRawCloudwatchApplicationLogs(framework, cwlClient)
		Expect(err).To(BeNil(), "Expected no errors reading the logs")
		logs, err := types.ParseLogs(utils.ToJsonLogs(raw))
		Expect(err).To(BeNil(), "Expected no errors parsing the logs: %s", raw)
		Expect(logs[0].Message).To(Equal(exception))
	})

})
