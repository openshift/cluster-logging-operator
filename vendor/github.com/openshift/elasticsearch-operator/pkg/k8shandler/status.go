package k8shandler

import (
	"fmt"
	"reflect"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"

	v1alpha1 "github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const healthUnknown = "cluster health unknown"
const NOT_FOUND_INDEX = -1

func UpdateClusterStatus(cluster *v1alpha1.Elasticsearch) error {

	clusterStatus := cluster.Status.DeepCopy()

	health, err := GetClusterHealth(cluster.Name, cluster.Namespace)
	if err != nil {
		health = healthUnknown
	}
	clusterStatus.ClusterHealth = health

	allocation, err := GetShardAllocation(cluster.Name, cluster.Namespace)
	switch {
	case allocation == "none":
		clusterStatus.ShardAllocationEnabled = v1alpha1.ShardAllocationNone
	case err != nil:
		clusterStatus.ShardAllocationEnabled = v1alpha1.ShardAllocationUnknown
	default:
		clusterStatus.ShardAllocationEnabled = v1alpha1.ShardAllocationAll
	}

	clusterStatus.Pods = rolePodStateMap(cluster.Namespace, cluster.Name)
	updateStatusConditions(clusterStatus)

	if !reflect.DeepEqual(clusterStatus, cluster.Status) {
		nretries := -1
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			nretries++
			if getErr := sdk.Get(cluster); getErr != nil {
				logrus.Debugf("Could not get Elasticsearch %v: %v", cluster.Name, getErr)
				return getErr
			}

			cluster.Status.ClusterHealth = clusterStatus.ClusterHealth
			cluster.Status.Conditions = clusterStatus.Conditions
			cluster.Status.Pods = clusterStatus.Pods
			cluster.Status.ShardAllocationEnabled = clusterStatus.ShardAllocationEnabled

			if updateErr := sdk.Update(cluster); updateErr != nil {
				logrus.Debugf("Failed to update Elasticsearch %s status. Reason: %v. Trying again...", cluster.Name, updateErr)
				return updateErr
			}
			return nil
		})

		if retryErr != nil {
			return fmt.Errorf("Error: could not update status for Elasticsearch %v after %v retries: %v", cluster.Name, nretries, retryErr)
		}
		logrus.Debugf("Updated Elasticsearch %v after %v retries", cluster.Name, nretries)
	}

	return nil
}

// if a status doesn't exist, provide a new one
func getNodeStatus(name string, status *v1alpha1.ElasticsearchStatus) (int, *v1alpha1.ElasticsearchNodeStatus) {
	for index, status := range status.Nodes {
		if status.DeploymentName == name || status.StatefulSetName == name {
			return index, &status
		}
	}

	return NOT_FOUND_INDEX, &v1alpha1.ElasticsearchNodeStatus{}
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

func updateStatusConditions(status *v1alpha1.ElasticsearchStatus) {
	if status.Conditions == nil {
		status.Conditions = make([]v1alpha1.ClusterCondition, 0, 4)
	}
	if _, condition := getESNodeCondition(status, v1alpha1.UpdatingSettings); condition == nil {
		updateUpdatingSettingsCondition(status, v1.ConditionFalse)
	}
	if _, condition := getESNodeCondition(status, v1alpha1.ScalingUp); condition == nil {
		updateScalingUpCondition(status, v1.ConditionFalse)
	}
	if _, condition := getESNodeCondition(status, v1alpha1.ScalingDown); condition == nil {
		updateScalingDownCondition(status, v1.ConditionFalse)
	}
	if _, condition := getESNodeCondition(status, v1alpha1.Restarting); condition == nil {
		updateRestartingCondition(status, v1.ConditionFalse)
	}
}

func getESNodeCondition(status *v1alpha1.ElasticsearchStatus, conditionType v1alpha1.ClusterConditionType) (int, *v1alpha1.ClusterCondition) {
	if status == nil {
		return -1, nil
	}
	for i := range status.Conditions {
		if status.Conditions[i].Type == conditionType {
			return i, &status.Conditions[i]
		}
	}
	return -1, nil
}

func updateESNodeCondition(status *v1alpha1.ElasticsearchStatus, condition *v1alpha1.ClusterCondition) bool {
	condition.LastTransitionTime = metav1.Now()
	// Try to find this node condition.
	conditionIndex, oldCondition := getESNodeCondition(status, condition.Type)

	if oldCondition == nil {
		// We are adding new node condition.
		status.Conditions = append(status.Conditions, *condition)
		return true
	}
	// We are updating an existing condition, so we need to check if it has changed.
	if condition.Status == oldCondition.Status {
		condition.LastTransitionTime = oldCondition.LastTransitionTime
	}

	isEqual := condition.Status == oldCondition.Status &&
		condition.Reason == oldCondition.Reason &&
		condition.Message == oldCondition.Message &&
		condition.LastTransitionTime.Equal(&oldCondition.LastTransitionTime)

	status.Conditions[conditionIndex] = *condition
	// Return true if one of the fields have changed.
	return !isEqual
}

func updateConditionWithRetry(dpl *v1alpha1.Elasticsearch, value v1.ConditionStatus,
	executeUpdateCondition func(*v1alpha1.ElasticsearchStatus, v1.ConditionStatus) bool) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if getErr := sdk.Get(dpl); getErr != nil {
			logrus.Debugf("Could not get Elasticsearch %v: %v", dpl.Name, getErr)
			return getErr
		}

		executeUpdateCondition(&dpl.Status, value)

		if updateErr := sdk.Update(dpl); updateErr != nil {
			logrus.Debugf("Failed to update Elasticsearch %v status: %v", dpl.Name, updateErr)
			return updateErr
		}
		return nil
	})
	return retryErr
}

func updateUpdatingSettingsCondition(status *v1alpha1.ElasticsearchStatus, value v1.ConditionStatus) bool {
	var message string
	if value == v1.ConditionTrue {
		message = "Config Map is different"
	} else {
		message = "Config Map is up to date"
	}
	return updateESNodeCondition(status, &v1alpha1.ClusterCondition{
		Type:    v1alpha1.UpdatingSettings,
		Status:  value,
		Reason:  "ConfigChange",
		Message: message,
	})
}

func updateScalingUpCondition(status *v1alpha1.ElasticsearchStatus, value v1.ConditionStatus) bool {
	return updateESNodeCondition(status, &v1alpha1.ClusterCondition{
		Type:   v1alpha1.ScalingUp,
		Status: value,
	})
}

func updateScalingDownCondition(status *v1alpha1.ElasticsearchStatus, value v1.ConditionStatus) bool {
	return updateESNodeCondition(status, &v1alpha1.ClusterCondition{
		Type:   v1alpha1.ScalingDown,
		Status: value,
	})
}

func updateRestartingCondition(status *v1alpha1.ElasticsearchStatus, value v1.ConditionStatus) bool {
	return updateESNodeCondition(status, &v1alpha1.ClusterCondition{
		Type:   v1alpha1.Restarting,
		Status: value,
	})
}
