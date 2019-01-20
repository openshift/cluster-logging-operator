package k8shandler

import (
	"fmt"

	"k8s.io/client-go/util/retry"

	v1alpha1 "github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	"github.com/openshift/elasticsearch-operator/pkg/utils"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
)

const healthUnknown = "cluster health unknown"

// UpdateStatus updates the status of Elasticsearch CRD
func (cState *ClusterState) UpdateStatus(dpl *v1alpha1.Elasticsearch) error {
	// TODO: only update this when is different from current...
	nretries := -1
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		nretries++
		if getErr := sdk.Get(dpl); getErr != nil {
			logrus.Debugf("Could not get Elasticsearch %v: %v", dpl.Name, getErr)
			return getErr
		}
		dpl.Status.ClusterHealth = clusterHealth(dpl)
		if dpl.Status.ShardAllocationEnabled == "" {
			dpl.Status.ShardAllocationEnabled = v1alpha1.ShardAllocationTrue
		}

		nodes := []v1alpha1.ElasticsearchNodeStatus{}
		for _, node := range cState.Nodes {
			nodes = append(nodes, *updateNodeStatus(node, &dpl.Status))
		}
		dpl.Status.Nodes = nodes
		updateStatusConditions(&dpl.Status)
		dpl.Status.Pods = rolePodStateMap(dpl.Namespace, dpl.Name)
		if updateErr := sdk.Update(dpl); updateErr != nil {
			logrus.Debugf("Failed to update Elasticsearch %s status. Reason: %v. Trying again...", dpl.Name, updateErr)
			return updateErr
		}
		return nil
	})

	if retryErr != nil {
		return fmt.Errorf("Error: could not update status for Elasticsearch %v after %v retries: %v", dpl.Name, nretries, retryErr)
	}
	logrus.Debugf("Updated Elasticsearch %v after %v retries", dpl.Name, nretries)
	return nil
}

func updateNodeStatus(node *nodeState, status *v1alpha1.ElasticsearchStatus) *v1alpha1.ElasticsearchNodeStatus {
	if status.Nodes == nil {
		status.Nodes = []v1alpha1.ElasticsearchNodeStatus{}
	}

	_, nodeStatus := statusExists(node, status)
	if nodeStatus == nil {
		nodeStatus = &v1alpha1.ElasticsearchNodeStatus{}
		nodeStatus.UpgradeStatus = *utils.NodeNormalOperation()
	}
	if node.Actual.Deployment != nil {
		nodeStatus.DeploymentName = node.Actual.Deployment.Name
	}

	if node.Actual.ReplicaSet != nil {
		nodeStatus.ReplicaSetName = node.Actual.ReplicaSet.Name
	}

	if node.Actual.Pod != nil {
		nodeStatus.PodName = node.Actual.Pod.Name
		nodeStatus.Status = string(node.Actual.Pod.Status.Phase)
	}

	if node.Actual.StatefulSet != nil {
		nodeStatus.StatefulSetName = node.Actual.StatefulSet.Name
	}

	if node.Desired.Roles != nil {
		nodeStatus.Roles = node.Desired.Roles
	}
	return nodeStatus
}

func statusExists(node *nodeState, status *v1alpha1.ElasticsearchStatus) (int, *v1alpha1.ElasticsearchNodeStatus) {
	var deploymentName string
	if node.Actual.Deployment != nil {
		deploymentName = node.Actual.Deployment.Name
	}
	if node.Actual.StatefulSet != nil {
		deploymentName = node.Actual.StatefulSet.Name
	}
	if deploymentName == "" {
		return -1, nil
	}

	for index, nodeStatus := range status.Nodes {
		if deploymentName == nodeStatus.DeploymentName ||
			deploymentName == nodeStatus.StatefulSetName {
			return index, &nodeStatus
		}
	}
	return -1, nil
}

func updateStatusConditions(status *v1alpha1.ElasticsearchStatus) {
	if status.Conditions == nil {
		status.Conditions = make([]v1alpha1.ClusterCondition, 0, 4)
	}
	if _, condition := utils.GetESNodeCondition(status, v1alpha1.UpdatingSettings); condition == nil {
		utils.UpdateUpdatingSettingsCondition(status, v1alpha1.ConditionFalse)
	}
	if _, condition := utils.GetESNodeCondition(status, v1alpha1.ScalingUp); condition == nil {
		utils.UpdateScalingUpCondition(status, v1alpha1.ConditionFalse)
	}
	if _, condition := utils.GetESNodeCondition(status, v1alpha1.ScalingDown); condition == nil {
		utils.UpdateScalingDownCondition(status, v1alpha1.ConditionFalse)
	}
	if _, condition := utils.GetESNodeCondition(status, v1alpha1.Restarting); condition == nil {
		utils.UpdateRestartingCondition(status, v1alpha1.ConditionFalse)
	}
}

func clusterHealth(dpl *v1alpha1.Elasticsearch) string {
	pods, err := listRunningPods(dpl.Name, dpl.Namespace)
	if err != nil {
		return healthUnknown
	}

	// no running elasticsearch pods were found
	if len(pods.Items) == 0 {
		return ""
	}

	// use arbitrary pod
	pod := pods.Items[0]

	clusterHealth, err := utils.ClusterHealth(&pod)
	if err != nil {
		return healthUnknown
	}

	status, present := clusterHealth["status"]
	if !present {
		logrus.Debug("response from elasticsearch health API did not contain 'status' field")
		return healthUnknown
	}

	// convert from type interface{} to string
	health, ok := status.(string)
	if !ok {
		return healthUnknown
	}

	return health
}

func rolePodStateMap(namespace string, clusterName string) map[v1alpha1.ElasticsearchNodeRole]v1alpha1.PodStateMap {

	baseSelector := fmt.Sprintf("component=%s", clusterName)
	clientList, _ := GetPodList(namespace, fmt.Sprintf("%s,%s", baseSelector, "es-node-client=true"))
	dataList, _ := GetPodList(namespace, fmt.Sprintf("%s,%s", baseSelector, "es-node-data=true"))
	masterList, _ := GetPodList(namespace, fmt.Sprintf("%s,%s", baseSelector, "es-node-master=true"))

	return map[v1alpha1.ElasticsearchNodeRole]v1alpha1.PodStateMap{
		v1alpha1.ElasticsearchRoleClient: podStateMap(clientList.Items),
		v1alpha1.ElasticsearchRoleData:   podStateMap(dataList.Items),
		v1alpha1.ElasticsearchRoleMaster: podStateMap(masterList.Items),
	}
}

func podStateMap(podList []v1.Pod) v1alpha1.PodStateMap {
	stateMap := map[v1alpha1.PodStateType][]string{
		v1alpha1.PodStateTypeReady:    []string{},
		v1alpha1.PodStateTypeNotReady: []string{},
		v1alpha1.PodStateTypeFailed:   []string{},
	}

	for _, pod := range podList {
		switch pod.Status.Phase {
		case v1.PodPending:
			stateMap[v1alpha1.PodStateTypeNotReady] = append(stateMap[v1alpha1.PodStateTypeNotReady], pod.Name)
		case v1.PodRunning:
			if isPodReady(pod) {
				stateMap[v1alpha1.PodStateTypeReady] = append(stateMap[v1alpha1.PodStateTypeReady], pod.Name)
			} else {
				stateMap[v1alpha1.PodStateTypeNotReady] = append(stateMap[v1alpha1.PodStateTypeNotReady], pod.Name)
			}
		case v1.PodFailed:
			stateMap[v1alpha1.PodStateTypeFailed] = append(stateMap[v1alpha1.PodStateTypeFailed], pod.Name)
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
