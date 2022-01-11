package functional

import (
	"context"
	"fmt"
	"github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/helpers/cmd"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"strings"
)

func (f *CollectorFunctionalFramework) ReadApplicationLogsFrom(outputName string) ([]types.ApplicationLog, error) {
	raw, err := f.ReadLogsFrom(outputName, applicationLog)
	if err != nil {
		return nil, err
	}

	var logs []types.ApplicationLog
	if err = types.StrictlyParseLogs(utils.ToJsonLogs(raw), &logs); err != nil {
		return nil, err
	}
	return logs, nil
}

func (f *CollectorFunctionalFramework) ReadRawApplicationLogsFrom(outputName string) ([]string, error) {
	return f.ReadLogsFrom(outputName, applicationLog)
}

func (f *CollectorFunctionalFramework) ReadInfrastructureLogsFrom(outputName string) ([]string, error) {
	return f.ReadLogsFrom(outputName, infraLog)
}
func (f *CollectorFunctionalFramework) ReadApplicationLogsFromKafka(topic string, brokerlistener string, consumercontainername string) (results []string, err error) {
	//inter broker zookeeper connect is plaintext so use plaintext port to check on sent messages from kafka producer ie. fluent-kafka plugin
	//till this step it must be ensured that a topic is created and published in kafka-consumer-clo-app-topic container within functional pod
	var result string
	outputFilename := "/shared/consumed.logs"
	cmd := fmt.Sprintf("tail -1 %s", outputFilename)
	err = wait.PollImmediate(defaultRetryInterval, maxDuration, func() (done bool, err error) {
		result, err = f.RunCommand(consumercontainername, "/bin/bash", "-c", cmd)
		if result != "" && err == nil {
			return true, nil
		}
		log.V(4).Error(err, "Polling logs from kafka")
		return false, nil
	})
	if err == nil {
		results = strings.Split(strings.TrimSpace(result), "\n")
	}
	log.V(4).Info("Returning", "logs", result)
	return results, err
}

func (f *CollectorFunctionalFramework) ReadAuditLogsFrom(outputName string) ([]string, error) {
	return f.ReadLogsFrom(outputName, auditLog)
}

func (f *CollectorFunctionalFramework) Readk8sAuditLogsFrom(outputName string) ([]string, error) {
	return f.ReadLogsFrom(outputName, k8sAuditLog)
}

func (f *CollectorFunctionalFramework) ReadOvnAuditLogsFrom(outputName string) ([]string, error) {
	return f.ReadLogsFrom(outputName, ovnAuditLog)
}

func (f *CollectorFunctionalFramework) ReadLogsFrom(outputName string, outputLogType string) (results []string, err error) {
	outputSpecs := f.Forwarder.Spec.OutputMap()
	outputLog := outputName
	if output, found := outputSpecs[outputName]; found {
		outputLog = output.Type
	}
	var result string
	outputType, ok := outputLogFile[outputLog]
	if !ok {
		return nil, fmt.Errorf(fmt.Sprintf("cant find output of type %s in outputSpec %v", outputName, outputSpecs))
	}
	file, ok := outputType[outputLogType]
	if !ok {
		return nil, fmt.Errorf(fmt.Sprintf("can't find log of type %s", outputLogType))
	}
	err = wait.PollImmediate(defaultRetryInterval, f.GetMaxReadDuration(), func() (done bool, err error) {
		result, err = f.RunCommand(outputName, "cat", file)
		if result != "" && err == nil {
			return true, nil
		}
		log.V(4).Error(err, "Polling logs")
		return false, nil
	})
	if err == nil {
		results = strings.Split(strings.TrimSpace(result), "\n")
	}
	log.V(4).Info("Returning", "logs", result)
	return results, err
}

func (f *CollectorFunctionalFramework) ReadNApplicationLogsFrom(n uint64, outputName string) ([]string, error) {
	lines := []string{}
	ctx, cancel := context.WithTimeout(context.Background(), test.SuccessTimeout())
	defer cancel()
	reader, err := cmd.TailReaderForContainer(f.Pod, outputName, ApplicationLogFile)
	if err != nil {
		log.V(3).Error(err, "Error creating tail reader")
		return nil, err
	}
	for {
		line, err := reader.ReadLineContext(ctx)
		if err != nil {
			log.V(3).Error(err, "Error readlinecontext")
			return nil, err
		}
		lines = append(lines, line)
		n--
		if n == 0 {
			break
		}
	}
	return lines, err
}
