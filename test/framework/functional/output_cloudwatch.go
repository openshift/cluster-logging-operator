package functional

import (
	"context"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
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

func (f *CollectorFunctionalFramework) GetLogGroupByType(client *cwl.Client, inputName string) ([]string, error) {
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

func (f *CollectorFunctionalFramework) GetLogMessagesByGroupAndStream(client *cwl.Client, groupName string, streamName string, atLeast int) ([]string, error) {
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
		if err != nil || len(logEventsOutput.Events) < atLeast {
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

func (f *CollectorFunctionalFramework) ReadLogsFromCloudwatch(client *cwl.Client, inputName string, atLeast int) ([]string, error) {
	log.V(3).Info("Reading cloudwatch log groups by type")
	logGroupName, err := f.GetLogGroupByType(client, inputName)
	if err != nil {
		return nil, err
	}
	log.V(3).Info("GetLogGroupByType", "logGroupName", logGroupName)

	log.V(3).Info("Reading cloudwatch log streams")
	logStreams, e := f.GetLogStreamsByGroup(client, logGroupName[0])
	if e != nil {
		return nil, e
	}
	log.V(3).Info("GetLogStreamsByGroup", "logStreams", logStreams)

	log.V(3).Info("Reading cloudwatch messages")
	messages, er := f.GetLogMessagesByGroupAndStream(client, logGroupName[0], logStreams[0], atLeast)
	if er != nil {
		return nil, er
	}
	log.V(3).Info("GetLogMessagesByGroupAndStream", "messages", messages)

	return messages, nil
}
