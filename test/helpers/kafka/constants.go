// Licensed to Red Hat, Inc under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Red Hat, Inc licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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
