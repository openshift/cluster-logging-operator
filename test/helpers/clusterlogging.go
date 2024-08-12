package helpers

import (
	receiverses "github.com/openshift/cluster-logging-operator/test/framework/e2e/receivers/elasticsearch"
)

type LogComponentType string

const (
	ComponentTypeStore                          LogComponentType = "LogStore"
	ComponentTypeVisualization                  LogComponentType = "Visualization"
	ComponentTypeCollector                      LogComponentType = "collector"
	ComponentTypeCollectorVector                LogComponentType = "collector-vector"
	ComponentTypeCollectorDeployment            LogComponentType = "collector-deployment"
	ComponentTypeReceiverElasticsearchRHManaged LogComponentType = receiverses.ManagedLogStore
)
