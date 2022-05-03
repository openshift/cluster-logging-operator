package functional

import (
	"context"
	"github.com/ViaQ/logerr/log"
	cwl "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"k8s.io/apimachinery/pkg/util/wait"
)

func (f *CollectorFunctionalFramework) GetAllCloudwatchGroups(svc *cwl.Client) ([]string, error) {
	var (
		allGroups       []string
		logGroupsOutput *cwl.DescribeLogGroupsOutput
	)
	err := wait.PollImmediate(defaultRetryInterval, f.GetMaxReadDuration(), func() (done bool, err error) {
		logGroupsOutput, err = svc.DescribeLogGroups(context.TODO(), &cwl.DescribeLogGroupsInput{})
		if err != nil || len(logGroupsOutput.LogGroups) == 0 {
			return false, err
		}
		return true, nil
	})

	if err != nil {
		return nil, err
	}
	log.V(3).Info("Results", "logGroups", logGroupsOutput.LogGroups)

	for _, l := range logGroupsOutput.LogGroups {
		allGroups = append(allGroups, *l.LogGroupName)
	}
	return allGroups, nil
}

func (f *CollectorFunctionalFramework) GetLogGroupsByType(client *cwl.Client, inputName string) ([]string, error) {
	var (
		myGroups        []string
		logGroupsOutput *cwl.DescribeLogGroupsOutput
	)
	err := wait.PollImmediate(defaultRetryInterval, f.GetMaxReadDuration(), func() (done bool, err error) {
		logGroupsOutput, err = client.DescribeLogGroups(context.TODO(), &cwl.DescribeLogGroupsInput{})
		if err != nil || len(logGroupsOutput.LogGroups) == 0 {
			return false, err
		}
		return true, nil
	})

	if err != nil {
		return nil, err
	}
	log.V(3).Info("Results", "logGroups", logGroupsOutput.LogGroups)

	for _, l := range logGroupsOutput.LogGroups {
		// Filter by type and get all
		if *l.LogGroupName == "group-prefix."+inputName {
			myGroups = append(myGroups, *l.LogGroupName)
		}
	}
	return myGroups, nil
}

func (f *CollectorFunctionalFramework) GetLogStreamsByGroup(client *cwl.Client, groupName string) ([]string, error) {
	var (
		myStreams        []string
		logStreamsOutput *cwl.DescribeLogStreamsOutput
	)
	err := wait.PollImmediate(defaultRetryInterval, f.GetMaxReadDuration(), func() (done bool, err error) {
		// TODO: need to query for log group or get more log group info above
		logStreamsOutput, err = client.DescribeLogStreams(
			context.TODO(),
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

func (f *CollectorFunctionalFramework) GetLogMessagesByGroupAndStream(client *cwl.Client, groupName string, streamName string) ([]string, error) {
	var (
		myMessages      []string
		logEventsOutput *cwl.GetLogEventsOutput
	)
	err := wait.PollImmediate(defaultRetryInterval, f.GetMaxReadDuration(), func() (done bool, err error) {
		// NOTE:  This is assuming only one stream is created for this group
		logEventsOutput, err = client.GetLogEvents(
			context.TODO(),
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

func (f *CollectorFunctionalFramework) ReadLogsFromCloudwatch(client *cwl.Client, inputName string) ([]string, error) {
	log.V(3).Info("Retrieving cloudwatch LogGroups ----------", "LogGroupName:", inputName)
	logGroupNames, err := f.GetLogGroupsByType(client, inputName)
	if err != nil {
		return nil, err
	}
	log.V(3).Info("Results", "logGroups", logGroupNames)

	log.V(3).Info("Retrieving cloudwatch LogStreams ----------")
	logStreams, e := f.GetLogStreamsByGroup(client, logGroupNames[0])
	if e != nil {
		return nil, e
	}
	log.V(3).Info("Results", "logStreams", logStreams)

	log.V(3).Info("Retrieving cloudwatch LogEvents  ----------")
	myMessages, er := f.GetLogMessagesByGroupAndStream(client, logGroupNames[0], logStreams[0])
	if er != nil {
		return nil, er
	}
	log.V(3).Info("Results", "myMessages", myMessages)

	return myMessages, nil
}
