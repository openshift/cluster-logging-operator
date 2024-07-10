package functional

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/aws/aws-sdk-go-v2/aws"
	cwl "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	openshiftv1 "github.com/openshift/api/route/v1"
	logging "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	AwsAccessKeyID      = "AKIAIOSFODNN7EXAMPLE"
	AwsSecretAccessKey  = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" //nolint:gosec
	cloudwatchMotoImage = "quay.io/openshift-logging/moto:2.2.3.dev0"
	CloudwatchSecret    = "cloudwatch-secret"
)

var (
	cwlClient *cwl.Client
	service   *v1.Service
	route     *openshiftv1.Route
)

func (f *CollectorFunctionalFramework) AddCloudWatchOutput(b *runtime.PodBuilder, obs obs.OutputSpec) error {
	if err := f.createCloudWatchService(); err != nil {
		return err
	}

	if err := f.createServiceRoute(); err != nil {
		return err
	}

	cwlClient = cwl.NewFromConfig(
		aws.Config{
			Region: "us-east-1",
			Credentials: aws.CredentialsProviderFunc(func(context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     AwsAccessKeyID,
					SecretAccessKey: AwsSecretAccessKey,
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

	b.AddContainer(logging.OutputTypeCloudwatch, cloudwatchMotoImage).
		WithCmdArgs([]string{"-s"}).
		End()

	return nil
}

func (f *CollectorFunctionalFramework) createCloudWatchService() error {
	service = runtime.NewService(f.Namespace, "moto-server")
	runtime.NewServiceBuilder(service).
		AddServicePort(5000, 5000).
		WithSelector(map[string]string{"testname": "functional"})

	if err := f.Test.Client.Create(service); err != nil {
		return fmt.Errorf("unable to create service: %v", err)
	}
	return nil
}

func (f *CollectorFunctionalFramework) createServiceRoute() error {
	log.V(2).Info("Creating route moto-server")
	route = runtime.NewRoute(f.Namespace, "moto-server", "moto-server", "5000")
	route.Spec.TLS = &openshiftv1.TLSConfig{
		Termination:                   openshiftv1.TLSTerminationPassthrough,
		InsecureEdgeTerminationPolicy: openshiftv1.InsecureEdgeTerminationPolicyNone,
	}
	if err := f.Test.Client.Create(route); err != nil {
		return fmt.Errorf("unable to create route: %v", err)
	}
	return nil
}

func (f *CollectorFunctionalFramework) pollForCloudwatchGroups() (*cwl.DescribeLogGroupsOutput, error) {
	var (
		logGroupsOutput *cwl.DescribeLogGroupsOutput
	)
	err := wait.PollUntilContextTimeout(context.TODO(), defaultRetryInterval, f.GetMaxReadDuration(), true, func(cxt context.Context) (done bool, err error) {
		logGroupsOutput, err = cwlClient.DescribeLogGroups(cxt, &cwl.DescribeLogGroupsInput{})
		if err != nil || len(logGroupsOutput.LogGroups) == 0 {
			return false, err
		}
		return true, nil
	})
	return logGroupsOutput, err
}

func (f *CollectorFunctionalFramework) GetAllCloudwatchGroups() ([]string, error) {
	var (
		allGroups []string
	)
	logGroupsOutput, err := f.pollForCloudwatchGroups()

	if err != nil {
		return nil, err
	}
	log.V(3).Info("Results", "logGroups", logGroupsOutput.LogGroups)

	for _, l := range logGroupsOutput.LogGroups {
		allGroups = append(allGroups, *l.LogGroupName)
	}
	return allGroups, nil
}

func (f *CollectorFunctionalFramework) GetLogGroupByType(inputName string) ([]string, error) {
	var (
		myGroups []string
	)
	logGroupsOutput, err := f.pollForCloudwatchGroups()

	if err != nil {
		return nil, err
	}
	log.V(3).Info("Results", "logGroups", logGroupsOutput.LogGroups)

	found := false
	for _, l := range logGroupsOutput.LogGroups {
		// Filter by type and get all
		if *l.LogGroupName == "group-prefix."+inputName {
			found = true
			myGroups = append(myGroups, *l.LogGroupName)
		}
	}
	if !found {
		return nil, fmt.Errorf("%s log group not found", inputName)
	}
	return myGroups, nil
}

func (f *CollectorFunctionalFramework) GetLogGroupByName(groupName string) ([]string, error) {
	var (
		myGroups []string
	)
	logGroupsOutput, err := f.pollForCloudwatchGroups()

	if err != nil {
		return nil, err
	}
	log.V(3).Info("Results", "logGroups", logGroupsOutput.LogGroups)

	found := false
	for _, l := range logGroupsOutput.LogGroups {
		// Filter by type and get all
		if *l.LogGroupName == groupName {
			found = true
			myGroups = append(myGroups, *l.LogGroupName)
		}
	}
	if !found {
		return nil, fmt.Errorf("%s log group not found", groupName)
	}
	return myGroups, nil
}

func (f *CollectorFunctionalFramework) GetLogStreamsByGroup(groupName string) ([]string, error) {
	var (
		myStreams        []string
		logStreamsOutput *cwl.DescribeLogStreamsOutput
	)
	err := wait.PollUntilContextTimeout(context.TODO(), defaultRetryInterval, f.GetMaxReadDuration(), true, func(cxt context.Context) (done bool, err error) {
		// TODO: need to query for log group or get more log group info above
		logStreamsOutput, err = cwlClient.DescribeLogStreams(
			cxt,
			&cwl.DescribeLogStreamsInput{LogGroupName: &groupName},
		)
		if err != nil || len(logStreamsOutput.LogStreams) == 0 {
			return false, err
		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	for _, stream := range logStreamsOutput.LogStreams {
		myStreams = append(myStreams, *stream.LogStreamName)
	}
	return myStreams, nil
}

func (f *CollectorFunctionalFramework) GetLogMessagesByGroupAndStream(groupName string, streamName string) ([]string, error) {
	var (
		myMessages      []string
		logEventsOutput *cwl.GetLogEventsOutput
	)
	err := wait.PollUntilContextTimeout(context.TODO(), defaultRetryInterval, f.GetMaxReadDuration(), true, func(cxt context.Context) (done bool, err error) {
		// NOTE:  This is assuming only one stream is created for this group
		logEventsOutput, err = cwlClient.GetLogEvents(
			cxt,
			&cwl.GetLogEventsInput{
				LogGroupName:  &groupName,
				LogStreamName: &streamName,
			})
		if err != nil || len(logEventsOutput.Events) == 0 {
			return false, err
		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	for _, event := range logEventsOutput.Events {
		myMessages = append(myMessages, *event.Message)
	}
	return myMessages, nil
}

func (f *CollectorFunctionalFramework) ReadLogsFromCloudwatch(inputName string) ([]string, error) {
	log.V(3).Info("Reading cloudwatch log groups by type")
	logGroupName, err := f.GetLogGroupByType(inputName)
	if err != nil {
		return nil, err
	}
	log.V(3).Info("GetLogGroupByType", "logGroupName", logGroupName)

	log.V(3).Info("Reading cloudwatch log streams")
	logStreams, e := f.GetLogStreamsByGroup(logGroupName[0])
	if e != nil {
		return nil, e
	}
	log.V(3).Info("GetLogStreamsByGroup", "logStreams", logStreams)

	log.V(3).Info("Reading cloudwatch messages")
	messages, er := f.GetLogMessagesByGroupAndStream(logGroupName[0], logStreams[0])
	if er != nil {
		return nil, er
	}
	log.V(3).Info("GetLogMessagesByGroupAndStream", "messages", messages)

	return messages, nil
}

func (f *CollectorFunctionalFramework) ReadLogsFromCloudwatchByGroupName(groupName string) ([]string, error) {
	log.V(3).Info(fmt.Sprintf("Reading cloudwatch log groups by name: %q", groupName))
	logGroupName, err := f.GetLogGroupByName(groupName)
	if err != nil {
		return nil, err
	}

	log.V(3).Info("Reading cloudwatch log streams")
	logStreams, e := f.GetLogStreamsByGroup(logGroupName[0])
	if e != nil {
		return nil, e
	}
	log.V(3).Info("GetLogStreamsByGroup", "logStreams", logStreams)

	log.V(3).Info("Reading cloudwatch messages")
	messages, er := f.GetLogMessagesByGroupAndStream(logGroupName[0], logStreams[0])
	if er != nil {
		return nil, er
	}
	log.V(3).Info("GetLogMessagesByGroupAndStream", "messages", messages)

	return messages, nil
}
