package kafka

import (
	"fmt"

	loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
)

const (
	DefaultTopic           = "clo-topic"
	AppLogsTopic           = "clo-app-topic"
	AuditLogsTopic         = "clo-audit-topic"
	InfraLogsTopic         = "clo-infra-topic"
	DeploymentName         = "kafka"
	ConsumerDeploymentName = "kafka-consumer"
	kafkaImageRepoOrg      = "quay.io/openshift-logging/"
	kafkaImageTag          = "2.7.0"
	KafkaImage             = kafkaImageRepoOrg + "kafka:" + kafkaImageTag
	KafkaInitUtilsImage    = kafkaImageRepoOrg + "kafka-initutils:" + kafkaImageTag
)

var (
	inputTypeToTopic map[string]string
)

func init() {
	inputTypeToTopic = map[string]string{
		loggingv1.InputNameApplication:    AppLogsTopic,
		loggingv1.InputNameAudit:          AuditLogsTopic,
		loggingv1.InputNameInfrastructure: InfraLogsTopic,
	}
}

func ClusterLocalEndpoint(namespace string) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local:%d", DeploymentName, namespace, kafkaInsidePort)
}

func ConsumerNameForTopic(topic string) string {
	return fmt.Sprintf("%s-%s", ConsumerDeploymentName, topic)
}

func TopicForInputName(topics []string, inputName string) string {
	if len(topics) == 1 {
		return DefaultTopic
	}
	return inputTypeToTopic[inputName]
}
