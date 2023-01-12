package elasticsearch

import (
	"context"
	"fmt"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	elasticsearch "github.com/openshift/elasticsearch-operator/apis/logging/v1"
	"k8s.io/client-go/util/retry"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
)

func UpdateStatus(k8sClient client.Client, namespace string, fetchClusterLogging func() (*logging.ClusterLogging, error)) error {

	elasticsearchStatus, err := getElasticsearchStatus(k8sClient, namespace)
	if err != nil {
		return fmt.Errorf("Failed to get Elasticsearch status: %v", err)
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		instance, err := fetchClusterLogging()
		if err != nil {
			return err
		}

		if !StatusAreSame(elasticsearchStatus, instance.Status.LogStore.ElasticsearchStatus) {
			instance.Status.LogStore.ElasticsearchStatus = elasticsearchStatus
			return k8sClient.Status().Update(context.TODO(), instance)
		}
		return nil
	})
	if retryErr != nil {
		return fmt.Errorf("Failed to update Cluster Logging Elasticsearch status: %v", retryErr)
	}
	return nil
}

func getElasticsearchStatus(k8sClient client.Client, namespace string) ([]logging.ElasticsearchStatus, error) {

	// we can scrape the status provided by the elasticsearch-operator
	// get list of elasticsearch objects
	esList := &elasticsearch.ElasticsearchList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Elasticsearch",
			APIVersion: elasticsearch.GroupVersion.String(),
		},
	}
	err := k8sClient.List(context.TODO(), esList, client.InNamespace(namespace))
	var status []logging.ElasticsearchStatus

	if err != nil {
		return status, fmt.Errorf("Unable to get Elasticsearches: %v", err)
	}

	if len(esList.Items) != 0 {
		for _, cluster := range esList.Items {
			nodeConditions := make(map[string]logging.ElasticsearchClusterConditions)

			nodeStatus := logging.ElasticsearchStatus{
				ClusterName:            cluster.Name,
				NodeCount:              cluster.Status.Cluster.NumNodes,
				ClusterHealth:          cluster.Status.ClusterHealth,
				Cluster:                cluster.Status.Cluster,
				Pods:                   getPodMap(cluster.Status),
				ClusterConditions:      logging.ElasticsearchClusterConditions(cluster.Status.Conditions),
				ShardAllocationEnabled: cluster.Status.ShardAllocationEnabled,
			}

			for _, node := range cluster.Status.Nodes {
				nodeName := ""

				if node.DeploymentName != "" {
					nodeName = node.DeploymentName
				}

				if node.StatefulSetName != "" {
					nodeName = node.StatefulSetName
				}

				if node.Conditions != nil {
					nodeConditions[nodeName] = logging.ElasticsearchClusterConditions(node.Conditions)
				} else {
					nodeConditions[nodeName] = []elasticsearch.ClusterCondition{}
				}
			}

			nodeStatus.NodeConditions = nodeConditions

			status = append(status, nodeStatus)
		}
	}

	return status, nil
}

func getPodMap(node elasticsearch.ElasticsearchStatus) map[logging.ElasticsearchRoleType]logging.PodStateMap {

	return map[logging.ElasticsearchRoleType]logging.PodStateMap{
		logging.ElasticsearchRoleTypeClient: translatePodMap(node.Pods[elasticsearch.ElasticsearchRoleClient]),
		logging.ElasticsearchRoleTypeData:   translatePodMap(node.Pods[elasticsearch.ElasticsearchRoleData]),
		logging.ElasticsearchRoleTypeMaster: translatePodMap(node.Pods[elasticsearch.ElasticsearchRoleMaster]),
	}
}

func translatePodMap(podStateMap elasticsearch.PodStateMap) logging.PodStateMap {

	return logging.PodStateMap{
		logging.PodStateTypeReady:    getPodState(podStateMap, elasticsearch.PodStateTypeReady),
		logging.PodStateTypeNotReady: getPodState(podStateMap, elasticsearch.PodStateTypeNotReady),
		logging.PodStateTypeFailed:   getPodState(podStateMap, elasticsearch.PodStateTypeFailed),
	}
}

func getPodState(podStateMap elasticsearch.PodStateMap, podStateType elasticsearch.PodStateType) []string {
	if v, ok := podStateMap[podStateType]; ok {
		sort.Strings(v)
		return v
	} else {
		return []string{}
	}
}
