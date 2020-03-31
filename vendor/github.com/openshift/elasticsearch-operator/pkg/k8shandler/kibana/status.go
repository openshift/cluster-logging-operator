package kibana

import (
	kibana "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (clusterRequest *KibanaRequest) getKibanaStatus() ([]kibana.KibanaStatus, error) {
	var status []kibana.KibanaStatus
	selector := map[string]string{
		"logging-infra": "kibana",
	}

	kibanaDeploymentList, err := clusterRequest.GetDeploymentList(selector)
	if err != nil {
		return status, err
	}

	for _, deployment := range kibanaDeploymentList.Items {
		selector["component"] = deployment.Name

		kibanaStatus := kibana.KibanaStatus{
			Deployment: deployment.Name,
			Replicas:   *deployment.Spec.Replicas,
		}

		replicaSetList, _ := clusterRequest.GetReplicaSetList(selector)
		var replicaNames []string
		for _, replicaSet := range replicaSetList.Items {
			replicaNames = append(replicaNames, replicaSet.Name)
		}
		kibanaStatus.ReplicaSets = replicaNames

		podList, _ := clusterRequest.GetPodList(selector)
		kibanaStatus.Pods = podStateMap(podList.Items)

		kibanaStatus.Conditions, err = clusterRequest.getPodConditions("kibana")
		if err != nil {
			return nil, err
		}

		status = append(status, kibanaStatus)
	}

	return status, nil
}

func podStateMap(podList []v1.Pod) kibana.PodStateMap {
	stateMap := map[kibana.PodStateType][]string{
		kibana.PodStateTypeReady:    {},
		kibana.PodStateTypeNotReady: {},
		kibana.PodStateTypeFailed:   {},
	}

	for _, pod := range podList {
		switch pod.Status.Phase {
		case v1.PodPending:
			stateMap[kibana.PodStateTypeNotReady] = append(stateMap[kibana.PodStateTypeNotReady], pod.Name)
		case v1.PodRunning:
			if isPodReady(pod) {
				stateMap[kibana.PodStateTypeReady] = append(stateMap[kibana.PodStateTypeReady], pod.Name)
			} else {
				stateMap[kibana.PodStateTypeNotReady] = append(stateMap[kibana.PodStateTypeNotReady], pod.Name)
			}
		case v1.PodFailed:
			stateMap[kibana.PodStateTypeFailed] = append(stateMap[kibana.PodStateTypeFailed], pod.Name)
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

func (clusterRequest *KibanaRequest) getPodConditions(component string) (map[string]kibana.ClusterConditions, error) {
	// Get all pods based on status.Nodes[] and check their conditions
	// get pod with label 'node-name=node.getName()'
	podConditions := make(map[string]kibana.ClusterConditions)

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
		return nil, err
	}

	for _, nodePod := range nodePodList.Items {

		var conditions []kibana.ClusterCondition

		isUnschedulable := false
		for _, podCondition := range nodePod.Status.Conditions {
			if podCondition.Type == v1.PodScheduled && podCondition.Status == v1.ConditionFalse {
				conditions = append(conditions, kibana.ClusterCondition{
					Type:               kibana.Unschedulable,
					Status:             v1.ConditionTrue,
					Reason:             podCondition.Reason,
					Message:            podCondition.Message,
					LastTransitionTime: podCondition.LastTransitionTime,
				})
				isUnschedulable = true
			}
		}

		if !isUnschedulable {
			for _, containerStatus := range nodePod.Status.ContainerStatuses {
				if containerStatus.State.Waiting != nil {
					conditions = append(conditions, kibana.ClusterCondition{
						Status:             v1.ConditionTrue,
						Reason:             containerStatus.State.Waiting.Reason,
						Message:            containerStatus.State.Waiting.Message,
						LastTransitionTime: metav1.Now(),
					})
				}
				if containerStatus.State.Terminated != nil {
					conditions = append(conditions, kibana.ClusterCondition{
						Status:             v1.ConditionTrue,
						Reason:             containerStatus.State.Terminated.Reason,
						Message:            containerStatus.State.Terminated.Message,
						LastTransitionTime: metav1.Now(),
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
