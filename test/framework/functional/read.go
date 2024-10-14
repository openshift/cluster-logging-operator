package functional

import (
	"context"
	"fmt"
	"strings"
	"time"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/url"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test"
	"github.com/openshift/cluster-logging-operator/test/helpers/cmd"
	"github.com/openshift/cluster-logging-operator/test/helpers/kafka"
	"github.com/openshift/cluster-logging-operator/test/helpers/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

type Option struct {
	Name  string
	Value string
}

// OptionsValue returns the value if found or the default
func OptionsValue(options []Option, name string, notFound any) any {
	if found, op := OptionsInclude(name, options); found {
		return op.Value
	}
	return notFound
}

func OptionsInclude(name string, options []Option) (bool, Option) {
	for _, o := range options {
		if o.Name == name {
			return true, o
		}
	}
	return false, Option{}
}

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
	return f.ReadLogsFrom(outputName, string(obs.InputTypeInfrastructure))
}

func (f *CollectorFunctionalFramework) ReadApplicationLogsFromKafka(topic string, brokerlistener string, consumercontainername string) (results []string, err error) {
	// inter broker zookeeper connect is plaintext so use plaintext port to check on sent messages from kafka producer ie. fluent-kafka plugin
	// till this step it must be ensured that a topic is created and published in kafka-consumer-clo-app-topic container within functional pod
	var result string
	outputFilename := "/shared/consumed.logs"
	cmd := fmt.Sprintf("tail -1 %s", outputFilename)
	err = wait.PollUntilContextTimeout(context.TODO(), defaultRetryInterval, maxDuration, true, func(cxt context.Context) (done bool, err error) {
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

func (f *CollectorFunctionalFramework) ReadLogsFrom(outputName, sourceType string) (results []string, err error) {
	outputSpecs := internalobs.Outputs(f.Forwarder.Spec.Outputs).Map()
	var outputSpec obs.OutputSpec
	if output, found := outputSpecs[outputName]; found {
		outputSpec = output
	}
	var readLogs func() ([]string, error)

	switch outputSpec.Type {
	case obs.OutputTypeKafka:
		readLogs = func() ([]string, error) {
			switch sourceType {
			case string(obs.InputTypeAudit):
				sourceType = kafka.AuditLogsTopic
			case string(obs.InputTypeInfrastructure):
				sourceType = kafka.InfraLogsTopic
			default:
				sourceType = kafka.AppLogsTopic
			}
			container := kafka.ConsumerNameForTopic(sourceType)
			return f.ReadApplicationLogsFromKafka(sourceType, "localhost:9092", container)
		}
	case obs.OutputTypeElasticsearch:
		readLogs = func() ([]string, error) {
			option := Option{"port", "9200"}
			if outputSpec.Elasticsearch != nil {
				if esurl, err := url.Parse(outputSpec.Elasticsearch.URL); err == nil {
					option.Value = esurl.Port()
				}
			}
			return f.GetLogsFromElasticSearch(outputName, sourceType, option)
		}
	default:
		readLogs = func() ([]string, error) {
			var result string
			outputFiles, ok := outputLogFile[string(outputSpec.Type)]
			if !ok {
				return nil, fmt.Errorf("cant find output of type %s in outputSpec %v", outputName, outputSpecs)
			}
			file, ok := outputFiles[sourceType]
			if !ok {
				return nil, fmt.Errorf("can't find log of type %s", sourceType)
			}

			result, err = f.ReadFileFrom(outputName, file)
			if err == nil {
				results = strings.Split(strings.TrimSpace(result), "\n")
			}
			log.V(4).Info("Returning", "logs", result)
			return results, err
		}
	}
	return readLogs()
}

func (f *CollectorFunctionalFramework) ReadFileFrom(outputName, filePath string) (result string, err error) {
	return f.ReadFileFromWithRetryInterval(outputName, filePath, defaultRetryInterval)
}

func (f *CollectorFunctionalFramework) ReadFileFromWithRetryInterval(outputName, filePath string, retryInterval time.Duration) (result string, err error) {
	err = wait.PollUntilContextTimeout(context.TODO(), retryInterval, f.GetMaxReadDuration(), true, func(cxt context.Context) (done bool, err error) {
		result, err = f.RunCommand(strings.ToLower(outputName), "cat", filePath)
		if result != "" && err == nil {
			return true, nil
		}
		log.V(4).Error(err, "Polling logs")
		return false, nil
	})
	log.V(3).Info("Returning", "content", result)
	return result, err
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

func (f *CollectorFunctionalFramework) ReadCollectorLogs() (string, error) {
	output, err := oc.Literal().From("oc logs -n %s pod/%s -c %s", f.Test.NS.Name, f.Name, constants.CollectorName).Run()
	return output, err
}
