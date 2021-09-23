package outputs

import (
	"context"
	"crypto/tls"
	"encoding/json"
	runtime2 "github.com/openshift/cluster-logging-operator/internal/runtime"
	"net/http"
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
	"github.com/openshift/cluster-logging-operator/test/functional"
)

var _ = Describe("[Functional][Outputs][CloudWatch] FluentdForward Output to CloudWatch", func() {

	const (
		logSize   = 128
		numOfLogs = 8

		// examples from AWS docs
		awsAccessKeyID     = "AKIAIOSFODNN7EXAMPLE"
		awsSecretAccessKey = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" //nolint:gosec
	)

	var (
		framework *functional.FluentdFunctionalFramework

		addMotoContainerVisitor = func(b *runtime2.PodBuilder) error {
			log.V(2).Info("Adding AWS CloudWatch Logs mock container")
			b.AddContainer("moto", "quay.io/openshift-logging/moto:2.2.3.dev0").
				WithCmdArgs([]string{"-s"}).
				End()
			return nil
		}

		mountCloudwatchSecretVisitor = func(b *runtime2.PodBuilder) error {
			log.V(2).Info("Mounting cloudwatch secret to the fluentd container")
			b.AddSecretVolume("cloudwatch", "cloudwatch").
				GetContainer(constants.CollectorName).
				AddVolumeMount("cloudwatch", "/var/run/ocp-collector/secrets/cloudwatch", "", true)
			return nil
		}

		cwlClient *cwl.Client

		maxDuration, _          = time.ParseDuration("10m")
		defaultRetryInterval, _ = time.ParseDuration("10s")
	)

	BeforeEach(func() {
		framework = functional.NewFluentdFunctionalFramework()

		log.V(2).Info("Creating service moto")
		service := runtime2.NewService(framework.Namespace, "moto")
		runtime2.NewServiceBuilderFor(service).
			AddServicePort(5000, 5000).
			WithSelector(map[string]string{"testname": "functional"})
		if err := framework.Test.Client.Create(service); err != nil {
			panic(err)
		}
		route := runtime2.NewRoute(framework.Namespace, "moto", "moto", "5000")
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

		log.V(2).Info("Creating secret cloudwatch with AWS example credentials")
		secret := runtime2.NewSecret(framework.Namespace, "cloudwatch",
			map[string][]byte{
				"aws_access_key_id":     []byte(awsAccessKeyID),
				"aws_secret_access_key": []byte(awsSecretAccessKey),
			},
		)
		if err := framework.Test.Client.Create(secret); err != nil {
			panic(err)
		}

		functional.NewClusterLogForwarderBuilder(framework.Forwarder).
			FromInput(logging.InputNameApplication).
			ToCloudwatchOutput()

		Expect(framework.DeployWithVisitors([]runtime2.PodBuilderVisitor{
			addMotoContainerVisitor,
			mountCloudwatchSecretVisitor,
		})).To(BeNil())

		Expect(framework.WritesNApplicationLogsOfSize(numOfLogs, logSize)).To(BeNil())
	})

	AfterEach(func() {
		framework.Cleanup()
	})

	Context("When sending a sequence of log messages to CloudWatch", func() {
		It("should get there", func() {

			var logGroupsOutput *cwl.DescribeLogGroupsOutput
			err := wait.PollImmediate(defaultRetryInterval, maxDuration, func() (done bool, err error) {
				logGroupsOutput, err = cwlClient.DescribeLogGroups(context.TODO(), nil)
				if err != nil || len(logGroupsOutput.LogGroups) == 0 {
					return false, nil
				}
				return true, nil
			})
			Expect(err).To(BeNil())
			Expect(logGroupsOutput).To(Not(BeNil()))
			Expect(len(logGroupsOutput.LogGroups)).To(Equal(1))
			Expect(*logGroupsOutput.LogGroups[0].LogGroupName).To(Equal("group-prefix.application"))

			var logStreamsOutput *cwl.DescribeLogStreamsOutput
			err = wait.PollImmediate(defaultRetryInterval, maxDuration, func() (done bool, err error) {
				logStreamsOutput, err = cwlClient.DescribeLogStreams(
					context.TODO(),
					&cwl.DescribeLogStreamsInput{LogGroupName: logGroupsOutput.LogGroups[0].LogGroupName},
				)
				if err != nil || len(logStreamsOutput.LogStreams) == 0 {
					return false, nil
				}
				return true, nil
			})
			Expect(err).To(BeNil())
			Expect(logStreamsOutput).To(Not(BeNil()))
			Expect(len(logStreamsOutput.LogStreams)).To(Equal(1))

			var logEventsOutput *cwl.GetLogEventsOutput
			err = wait.PollImmediate(defaultRetryInterval, maxDuration, func() (done bool, err error) {
				logEventsOutput, err = cwlClient.GetLogEvents(
					context.TODO(),
					&cwl.GetLogEventsInput{
						LogGroupName:  logGroupsOutput.LogGroups[0].LogGroupName,
						LogStreamName: logStreamsOutput.LogStreams[0].LogStreamName,
					})
				if err != nil || len(logEventsOutput.Events) == 0 {
					return false, nil
				}
				return true, nil
			})
			Expect(err).To(BeNil())
			Expect(logEventsOutput).To(Not(BeNil()))
			Expect(len(logEventsOutput.Events)).To(Equal(numOfLogs))
			for i := 0; i < numOfLogs; i++ {
				type messageEntry struct{ Message string }
				var entry messageEntry
				err := json.Unmarshal([]byte(*logEventsOutput.Events[i].Message), &entry)
				Expect(err).To(BeNil())
				Expect(len(entry.Message)).To(Equal(logSize))
			}
		})
	})
})
