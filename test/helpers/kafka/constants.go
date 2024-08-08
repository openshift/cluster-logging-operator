package kafka

import (
	"fmt"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
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
		string(obs.InputTypeApplication):    AppLogsTopic,
		string(obs.InputTypeAudit):          AuditLogsTopic,
		string(obs.InputTypeInfrastructure): InfraLogsTopic,
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
