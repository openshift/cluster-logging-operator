package k8shandler

import (
	"fmt"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	elasticsearch "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (clusterRequest *ClusterLoggingRequest) getCuratorStatus() ([]logging.CuratorStatus, error) {

	status := []logging.CuratorStatus{}

	curatorCronJobList, err := clusterRequest.GetCronJobList(
		map[string]string{
			"logging-infra": "curator",
		},
	)
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

func (clusterRequest *ClusterLoggingRequest) getFluentdCollectorStatus() (logging.FluentdCollectorStatus, error) {

	fluentdStatus := logging.FluentdCollectorStatus{}
	selector := map[string]string{
		"logging-infra": "fluentd",
	}

	fluentdDaemonsetList, err := clusterRequest.GetDaemonSetList(selector)

	if err != nil {
		return fluentdStatus, err
	}

	if len(fluentdDaemonsetList.Items) != 0 {
		daemonset := fluentdDaemonsetList.Items[0]

		fluentdStatus.DaemonSet = daemonset.Name

		// use map to represent {pod: node}
		podList, _ := clusterRequest.GetPodList(selector)

		podNodeMap := make(map[string]string)
		for _, pod := range podList.Items {
			podNodeMap[pod.Name] = pod.Spec.NodeName
		}
		fluentdStatus.Pods = podStateMap(podList.Items)
		fluentdStatus.Nodes = podNodeMap
	}

	return fluentdStatus, nil
}

func (clusterRequest *ClusterLoggingRequest) getRsyslogCollectorStatus() (logging.RsyslogCollectorStatus, error) {

	rsyslogStatus := logging.RsyslogCollectorStatus{}
	selector := map[string]string{
		"logging-infra": "rsyslog",
	}

	rsyslogDaemonsetList, err := clusterRequest.GetDaemonSetList(selector)

	if err != nil {
		return rsyslogStatus, err
	}

	if len(rsyslogDaemonsetList.Items) != 0 {
		daemonset := rsyslogDaemonsetList.Items[0]

		rsyslogStatus.DaemonSet = daemonset.Name

		// use map to represent {pod: node}
		podList, _ := clusterRequest.GetPodList(selector)

		podNodeMap := make(map[string]string)
		for _, pod := range podList.Items {
			podNodeMap[pod.Name] = pod.Spec.NodeName
		}
		rsyslogStatus.Pods = podStateMap(podList.Items)
		rsyslogStatus.Nodes = podNodeMap
	}

	return rsyslogStatus, nil
}

func (clusterRequest *ClusterLoggingRequest) getKibanaStatus() ([]logging.KibanaStatus, error) {

	status := []logging.KibanaStatus{}
	selector := map[string]string{
		"logging-infra": "kibana",
	}

	kibanaDeploymentList, err := clusterRequest.GetDeploymentList(selector)
	if err != nil {
		return status, err
	}

	for _, deployment := range kibanaDeploymentList.Items {

		selector["component"] = deployment.Name

		kibanaStatus := logging.KibanaStatus{
			Deployment: deployment.Name,
			Replicas:   *deployment.Spec.Replicas,
		}

		replicaSetList, _ := clusterRequest.GetReplicaSetList(selector)
		replicaNames := []string{}
		for _, replicaSet := range replicaSetList.Items {
			replicaNames = append(replicaNames, replicaSet.Name)
		}
		kibanaStatus.ReplicaSets = replicaNames

		podList, _ := clusterRequest.GetPodList(selector)
		kibanaStatus.Pods = podStateMap(podList.Items)

		status = append(status, kibanaStatus)
	}

	return status, nil
}

func (clusterRequest *ClusterLoggingRequest) getElasticsearchStatus() ([]logging.ElasticsearchStatus, error) {

	// we can scrape the status provided by the elasticsearch-operator
	// get list of elasticsearch objects
	esList := &elasticsearch.ElasticsearchList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Elasticsearch",
			APIVersion: elasticsearch.SchemeGroupVersion.String(),
		},
	}

	err := clusterRequest.List(map[string]string{}, esList)
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

// We are no longer going to populate this field, however since it is in the
// status spec we cannot just remove it.
func getReplicaSetNames(node elasticsearch.ElasticsearchStatus) []string {

	replicasetNames := []string{}
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
