package k8shandler

import (
	"bytes"
	"fmt"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	v1 "k8s.io/api/core/v1"

	elasticsearch "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (cluster *ClusterLogging) getCuratorStatus() ([]logging.CuratorStatus, error) {

	status := []logging.CuratorStatus{}

	curatorCronJobList, err := utils.GetCronJobList(cluster.Namespace, "logging-infra=curator")
	if err != nil {
		return status, err
	}

	for _, cronjob := range curatorCronJobList.Items {

		curatorStatus := logging.CuratorStatus{
			CronJob:   cronjob.Name,
			Schedule:  cronjob.Spec.Schedule,
			Suspended: *cronjob.Spec.Suspend,
		}

		status = append(status, curatorStatus)
	}

	return status, nil
}

func getFluentdCollectorStatus(namespace string) (logging.FluentdCollectorStatus, error) {

	fluentdStatus := logging.FluentdCollectorStatus{}

	fluentdDaemonsetList, err := utils.GetDaemonSetList(namespace, "logging-infra=fluentd")
	if err != nil {
		return fluentdStatus, err
	}

	if len(fluentdDaemonsetList.Items) != 0 {
		daemonset := fluentdDaemonsetList.Items[0]

		fluentdStatus.DaemonSet = daemonset.Name

		// use map to represent {pod: node}
		podList, _ := utils.GetPodList(namespace, "logging-infra=fluentd")
		podNodeMap := make(map[string]string)
		for _, pod := range podList.Items {
			podNodeMap[pod.Name] = pod.Spec.NodeName
		}
		fluentdStatus.Pods = podStateMap(podList.Items)
		fluentdStatus.Nodes = podNodeMap
	}

	return fluentdStatus, nil
}

func getRsyslogCollectorStatus(namespace string) (logging.RsyslogCollectorStatus, error) {

	rsyslogStatus := logging.RsyslogCollectorStatus{}

	rsyslogDaemonsetList, err := utils.GetDaemonSetList(namespace, "logging-infra=rsyslog")
	if err != nil {
		return rsyslogStatus, err
	}

	if len(rsyslogDaemonsetList.Items) != 0 {
		daemonset := rsyslogDaemonsetList.Items[0]

		rsyslogStatus.DaemonSet = daemonset.Name

		// use map to represent {pod: node}
		podList, _ := utils.GetPodList(namespace, "logging-infra=rsyslog")
		podNodeMap := make(map[string]string)
		for _, pod := range podList.Items {
			podNodeMap[pod.Name] = pod.Spec.NodeName
		}
		rsyslogStatus.Pods = podStateMap(podList.Items)
		rsyslogStatus.Nodes = podNodeMap
	}

	return rsyslogStatus, nil
}

func (cluster *ClusterLogging) getKibanaStatus() ([]logging.KibanaStatus, error) {

	status := []logging.KibanaStatus{}

	kibanaDeploymentList, err := utils.GetDeploymentList(cluster.Namespace, "logging-infra=kibana")
	if err != nil {
		return status, err
	}

	for _, deployment := range kibanaDeploymentList.Items {

		var selectorValue bytes.Buffer
		selectorValue.WriteString("component=")
		selectorValue.WriteString(deployment.Name)

		kibanaStatus := logging.KibanaStatus{
			Deployment: deployment.Name,
			Replicas:   *deployment.Spec.Replicas,
		}

		replicaSetList, _ := utils.GetReplicaSetList(cluster.Namespace, selectorValue.String())
		replicaNames := []string{}
		for _, replicaSet := range replicaSetList.Items {
			replicaNames = append(replicaNames, replicaSet.Name)
		}
		kibanaStatus.ReplicaSets = replicaNames

		podList, _ := utils.GetPodList(cluster.Namespace, selectorValue.String())
		kibanaStatus.Pods = podStateMap(podList.Items)

		status = append(status, kibanaStatus)
	}

	return status, nil
}

func (cluster *ClusterLogging) getElasticsearchStatus() ([]logging.ElasticsearchStatus, error) {

	// we can scrape the status provided by the elasticsearch-operator
	// get list of elasticsearch objects
	esList := &elasticsearch.ElasticsearchList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Elasticsearch",
			APIVersion: elasticsearch.SchemeGroupVersion.String(),
		},
	}

	err := sdk.List(cluster.Namespace, esList)
	status := []logging.ElasticsearchStatus{}

	if err != nil {
		return status, fmt.Errorf("Unable to get Elasticsearches: %v", err)
	}

	if len(esList.Items) != 0 {
		for _, node := range esList.Items {

			nodeStatus := logging.ElasticsearchStatus{
				ClusterName:   node.Name,
				NodeCount:     node.Spec.Nodes[0].NodeCount,
				Deployments:   getDeploymentNames(node.Status),
				ReplicaSets:   getReplicaSetNames(node.Status),
				StatefulSets:  getStatefulSetNames(node.Status),
				Pods:          getPodMap(node.Status),
				ClusterHealth: node.Status.ClusterHealth,
			}

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
		logging.PodStateTypeReady:    podStateMap[elasticsearch.PodStateTypeReady],
		logging.PodStateTypeNotReady: podStateMap[elasticsearch.PodStateTypeNotReady],
		logging.PodStateTypeFailed:   podStateMap[elasticsearch.PodStateTypeFailed],
	}
}

func getDeploymentNames(node elasticsearch.ElasticsearchStatus) []string {

	deploymentNames := []string{}

	for _, nodeStatus := range node.Nodes {
		deploymentNames = append(deploymentNames, nodeStatus.DeploymentName)
	}

	return deploymentNames
}

func getReplicaSetNames(node elasticsearch.ElasticsearchStatus) []string {

	replicasetNames := []string{}

	for _, nodeStatus := range node.Nodes {
		replicasetNames = append(replicasetNames, nodeStatus.ReplicaSetName)
	}

	return replicasetNames
}

func getStatefulSetNames(node elasticsearch.ElasticsearchStatus) []string {

	statefulsetNames := []string{}

	for _, nodeStatus := range node.Nodes {
		statefulsetNames = append(statefulsetNames, nodeStatus.StatefulSetName)
	}

	return statefulsetNames
}

func podStateMap(podList []v1.Pod) logging.PodStateMap {
	stateMap := map[logging.PodStateType][]string{
		logging.PodStateTypeReady:    []string{},
		logging.PodStateTypeNotReady: []string{},
		logging.PodStateTypeFailed:   []string{},
	}

	for _, pod := range podList {
		switch pod.Status.Phase {
		case v1.PodPending:
			stateMap[logging.PodStateTypeNotReady] = append(stateMap[logging.PodStateTypeNotReady], pod.Name)
		case v1.PodRunning:
			if isPodReady(pod) {
				stateMap[logging.PodStateTypeReady] = append(stateMap[logging.PodStateTypeReady], pod.Name)
			} else {
				stateMap[logging.PodStateTypeNotReady] = append(stateMap[logging.PodStateTypeNotReady], pod.Name)
			}
		case v1.PodFailed:
			stateMap[logging.PodStateTypeFailed] = append(stateMap[logging.PodStateTypeFailed], pod.Name)
		}
	}

	return stateMap
}

func isPodReady(pod v1.Pod) bool {

	for _, container := range pod.Status.ContainerStatuses {
		if !container.Ready {
			return false
		}
	}

	return true
}
