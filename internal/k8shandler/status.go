package k8shandler

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"sort"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	elasticsearch "github.com/openshift/elasticsearch-operator/apis/logging/v1"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (clusterRequest *ClusterLoggingRequest) getFluentdCollectorStatus() (logging.FluentdCollectorStatus, error) {

	fluentdStatus := logging.FluentdCollectorStatus{}
	selector := map[string]string{
		"logging-infra": constants.CollectorName,
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

		fluentdStatus.Conditions, err = clusterRequest.getPodConditions(constants.CollectorName)
		if err != nil {
			return fluentdStatus, fmt.Errorf("unable to get pod conditions. E: %s", err.Error())
		}
	}

	return fluentdStatus, nil
}

func (clusterRequest *ClusterLoggingRequest) getKibanaStatus() ([]elasticsearch.KibanaStatus, error) {
	cr, err := clusterRequest.getKibanaCR()
	if err != nil {
		return nil, err
	}
	return cr.Status, nil
}

func (clusterRequest *ClusterLoggingRequest) getElasticsearchStatus() ([]logging.ElasticsearchStatus, error) {

	// we can scrape the status provided by the elasticsearch-operator
	// get list of elasticsearch objects
	esList := &elasticsearch.ElasticsearchList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Elasticsearch",
			APIVersion: elasticsearch.GroupVersion.String(),
		},
	}

	err := clusterRequest.List(map[string]string{}, esList)
	status := []logging.ElasticsearchStatus{}

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

func podStateMap(podList []v1.Pod) logging.PodStateMap {
	stateMap := map[logging.PodStateType][]string{
		logging.PodStateTypeReady:    {},
		logging.PodStateTypeNotReady: {},
		logging.PodStateTypeFailed:   {},
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
	for _, pods := range stateMap {
		sort.Strings(pods)
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

func (clusterRequest *ClusterLoggingRequest) getPodConditions(component string) (map[string]logging.ClusterConditions, error) {
	// Get all pods based on status.Nodes[] and check their conditions
	// get pod with label 'node-name=node.getName()'
	podConditions := make(map[string]logging.ClusterConditions)

	nodePodList := &core.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: core.SchemeGroupVersion.String(),
		},
	}

	if err := clusterRequest.List(
		map[string]string{
			"component": component,
		},
		nodePodList,
	); err != nil {
		return nil, fmt.Errorf("unable to list pods. E: %s", err.Error())
	}

	for _, nodePod := range nodePodList.Items {

		var conditions []logging.Condition

		isUnschedulable := false
		for _, podCondition := range nodePod.Status.Conditions {
			if podCondition.Type == v1.PodScheduled && podCondition.Status == v1.ConditionFalse {
				conditions = append(conditions, logging.Condition{
					Type:               logging.Unschedulable,
					Status:             v1.ConditionTrue,
					Reason:             logging.ConditionReason(podCondition.Reason),
					Message:            podCondition.Message,
					LastTransitionTime: podCondition.LastTransitionTime,
				})
				isUnschedulable = true
			}
		}

		if !isUnschedulable {
			for _, containerStatus := range nodePod.Status.ContainerStatuses {
				if containerStatus.State.Waiting != nil {
					conditions = append(conditions, logging.Condition{
						Type:               logging.ContainerWaiting,
						Status:             v1.ConditionTrue,
						Reason:             logging.ConditionReason(containerStatus.State.Waiting.Reason),
						Message:            containerStatus.State.Waiting.Message,
						LastTransitionTime: metav1.Now(),
					})
				}
				if containerStatus.State.Terminated != nil {
					conditions = append(conditions, logging.Condition{
						Type:               logging.ContainerTerminated,
						Status:             v1.ConditionTrue,
						Reason:             logging.ConditionReason(containerStatus.State.Terminated.Reason),
						Message:            containerStatus.State.Terminated.Message,
						LastTransitionTime: containerStatus.State.Terminated.FinishedAt,
					})
				}
			}
		}

		if len(conditions) > 0 {
			podConditions[nodePod.Name] = conditions
		}
	}

	return podConditions, nil
}
