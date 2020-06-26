package loki

import "fmt"

const (
	DeploymentName = "loki"
	ListenerPort   = 3100

	QuerierName      = "loki-querier"
	QuerierComponent = "test"
)

func ClusterLocalEndpoint(namespace string) string {
	return fmt.Sprintf("%s.%s.svc.cluster.local:%d", DeploymentName, namespace, ListenerPort)
}
