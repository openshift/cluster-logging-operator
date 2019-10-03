package k8shandler

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	api "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const healthUnknown = "cluster health unknown"
const NOT_FOUND_INDEX = -1

var DISK_WATERMARK_LOW_PCT *float64
var DISK_WATERMARK_HIGH_PCT *float64
var DISK_WATERMARK_LOW_ABS *resource.Quantity
var DISK_WATERMARK_HIGH_ABS *resource.Quantity

func (elasticsearchRequest *ElasticsearchRequest) UpdateClusterStatus() error {

	cluster := elasticsearchRequest.cluster

	clusterStatus := cluster.Status.DeepCopy()

	health, err := GetClusterHealth(cluster.Name, cluster.Namespace, elasticsearchRequest.client)
	if err != nil {
		health.Status = healthUnknown
	}
	clusterStatus.Cluster = health

	allocation, err := GetShardAllocation(cluster.Name, cluster.Namespace, elasticsearchRequest.client)
	switch {
	case allocation == "none":
		clusterStatus.ShardAllocationEnabled = api.ShardAllocationNone
	case err != nil:
		clusterStatus.ShardAllocationEnabled = api.ShardAllocationUnknown
	default:
		clusterStatus.ShardAllocationEnabled = api.ShardAllocationAll
	}

	clusterStatus.Pods = rolePodStateMap(cluster.Namespace, cluster.Name, elasticsearchRequest.client)
	updateStatusConditions(clusterStatus)
	updateNodeConditions(cluster.Name, cluster.Namespace, clusterStatus, elasticsearchRequest.client)

	if !reflect.DeepEqual(clusterStatus, cluster.Status) {
		nretries := -1
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			nretries++
			if getErr := elasticsearchRequest.client.Get(context.TODO(), types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}, cluster); getErr != nil {
				logrus.Debugf("Could not get Elasticsearch %v: %v", cluster.Name, getErr)
				return getErr
			}

			cluster.Status.Cluster = clusterStatus.Cluster
			cluster.Status.Conditions = clusterStatus.Conditions
			cluster.Status.Pods = clusterStatus.Pods
			cluster.Status.ShardAllocationEnabled = clusterStatus.ShardAllocationEnabled
			cluster.Status.Nodes = clusterStatus.Nodes

			if updateErr := elasticsearchRequest.client.Update(context.TODO(), cluster); updateErr != nil {
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

func (elasticsearchRequest *ElasticsearchRequest) GetCurrentPodStateMap() map[api.ElasticsearchNodeRole]api.PodStateMap {
	return rolePodStateMap(elasticsearchRequest.cluster.Namespace, elasticsearchRequest.cluster.Name, elasticsearchRequest.client)
}

func containsClusterCondition(condition api.ClusterConditionType, status v1.ConditionStatus, elasticsearchStatus *api.ElasticsearchStatus) bool {
	// if we're looking for a status of v1.ConditionTrue then we want to see if the
	// condition is present and the status is the same
	//
	// if we're looking for a status of v1.ConditionFalse then we want the condition
	// to either be present with status of false or to not find the condition
	defaultValue := (status != v1.ConditionTrue)

	for _, clusterCondition := range elasticsearchStatus.Conditions {
		if clusterCondition.Type == condition {
			return clusterCondition.Status == status
		}
	}

	return defaultValue
}

// if a status doesn't exist, provide a new one
func getNodeStatus(name string, status *api.ElasticsearchStatus) (int, *api.ElasticsearchNodeStatus) {
	for index, status := range status.Nodes {
		if status.DeploymentName == name || status.StatefulSetName == name {
			return index, &status
		}
	}

	return NOT_FOUND_INDEX, &api.ElasticsearchNodeStatus{}
}

func rolePodStateMap(namespace, clusterName string, client client.Client) map[api.ElasticsearchNodeRole]api.PodStateMap {

	clientList, _ := GetPodList(
		namespace,
		map[string]string{
			"component":      "elasticsearch",
			"cluster-name":   clusterName,
			"es-node-client": "true",
		},
		client,
	)
	dataList, _ := GetPodList(
		namespace,
		map[string]string{
			"component":    "elasticsearch",
			"cluster-name": clusterName,
			"es-node-data": "true",
		},
		client,
	)
	masterList, _ := GetPodList(
		namespace,
		map[string]string{
			"component":      "elasticsearch",
			"cluster-name":   clusterName,
			"es-node-master": "true",
		},
		client,
	)

	return map[api.ElasticsearchNodeRole]api.PodStateMap{
		api.ElasticsearchRoleClient: podStateMap(clientList.Items),
		api.ElasticsearchRoleData:   podStateMap(dataList.Items),
		api.ElasticsearchRoleMaster: podStateMap(masterList.Items),
	}
}

func podStateMap(podList []v1.Pod) api.PodStateMap {
	stateMap := map[api.PodStateType][]string{
		api.PodStateTypeReady:    []string{},
		api.PodStateTypeNotReady: []string{},
		api.PodStateTypeFailed:   []string{},
	}

	for _, pod := range podList {
		switch pod.Status.Phase {
		case v1.PodPending:
			stateMap[api.PodStateTypeNotReady] = append(stateMap[api.PodStateTypeNotReady], pod.Name)
		case v1.PodRunning:
			if isPodReady(pod) {
				stateMap[api.PodStateTypeReady] = append(stateMap[api.PodStateTypeReady], pod.Name)
			} else {
				stateMap[api.PodStateTypeNotReady] = append(stateMap[api.PodStateTypeNotReady], pod.Name)
			}
		case v1.PodFailed:
			stateMap[api.PodStateTypeFailed] = append(stateMap[api.PodStateTypeFailed], pod.Name)
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

func updateNodeConditions(clusterName, namespace string, status *api.ElasticsearchStatus, client client.Client) {
	// Get all pods based on status.Nodes[] and check their conditions
	// get pod with label 'node-name=node.getName()'
	thresholdEnabled, err := GetThresholdEnabled(clusterName, namespace, client)
	if err != nil {
		logrus.Debugf("Unable to check if threshold is enabled for %v", clusterName)
	}

	if thresholdEnabled {
		// refresh value of thresholds in case they changed...
		refreshDiskWatermarkThresholds(clusterName, namespace, client)
	}

	for nodeIndex, _ := range status.Nodes {
		node := &status.Nodes[nodeIndex]

		nodeName := "unknown name"
		if node.DeploymentName != "" {
			nodeName = node.DeploymentName
		} else {
			if node.StatefulSetName != "" {
				nodeName = node.StatefulSetName
			}
		}

		nodePodList, _ := GetPodList(
			namespace,
			map[string]string{
				"component":    "elasticsearch",
				"cluster-name": clusterName,
				"node-name":    nodeName,
			},
			client,
		)
		for _, nodePod := range nodePodList.Items {

			isUnschedulable := false
			for _, podCondition := range nodePod.Status.Conditions {
				if podCondition.Type == v1.PodScheduled && podCondition.Status == v1.ConditionFalse {
					podCondition.Type = v1.PodReasonUnschedulable
					podCondition.Status = v1.ConditionTrue
					updatePodUnschedulableCondition(node, podCondition)
					isUnschedulable = true
				}
			}

			if isUnschedulable {
				continue
			}
			updatePodUnschedulableCondition(node, v1.PodCondition{
				Status: v1.ConditionFalse,
			})

			// if the pod can't be scheduled we shouldn't enter here
			for _, containerStatus := range nodePod.Status.ContainerStatuses {
				if containerStatus.Name == "elasticsearch" {
					if containerStatus.State.Waiting != nil {
						updatePodNotReadyCondition(
							node,
							api.ESContainerWaiting,
							containerStatus.State.Waiting.Reason,
							containerStatus.State.Waiting.Message,
						)
					} else {
						updatePodNotReadyCondition(
							node,
							api.ESContainerWaiting,
							"",
							"",
						)
					}
					if containerStatus.State.Terminated != nil {
						updatePodNotReadyCondition(
							node,
							api.ESContainerTerminated,
							containerStatus.State.Terminated.Reason,
							containerStatus.State.Terminated.Message,
						)
					} else {
						updatePodNotReadyCondition(
							node,
							api.ESContainerTerminated,
							"",
							"",
						)
					}
				}
				if containerStatus.Name == "proxy" {
					if containerStatus.State.Waiting != nil {
						updatePodNotReadyCondition(
							node,
							api.ProxyContainerWaiting,
							containerStatus.State.Waiting.Reason,
							containerStatus.State.Waiting.Message,
						)
					} else {
						updatePodNotReadyCondition(
							node,
							api.ProxyContainerWaiting,
							"",
							"",
						)
					}
					if containerStatus.State.Terminated != nil {
						updatePodNotReadyCondition(
							node,
							api.ProxyContainerTerminated,
							containerStatus.State.Terminated.Reason,
							containerStatus.State.Terminated.Message,
						)
					} else {
						updatePodNotReadyCondition(
							node,
							api.ProxyContainerTerminated,
							"",
							"",
						)
					}
				}
			}

			if !thresholdEnabled {
				// disk threshold is not enabled, continue to next node
				continue
			}

			usage, percent, err := GetNodeDiskUsage(clusterName, namespace, nodeName, client)
			if err != nil {
				logrus.Debugf("Unable to get disk usage for %v", nodeName)
				continue
			}

			if exceedsLowWatermark(usage, percent) {
				if exceedsHighWatermark(usage, percent) {
					updatePodNodeStorageCondition(
						node,
						"Disk Watermark High",
						fmt.Sprintf("Disk storage usage for node is %vb (%v%%). Shards will be relocated from this node.", usage, percent),
					)
				} else {
					updatePodNodeStorageCondition(
						node,
						"Disk Watermark Low",
						fmt.Sprintf("Disk storage usage for node is %vb (%v%%). Shards will be not be allocated on this node.", usage, percent),
					)
				}
			} else {
				if percent > float64(0.0) {
					// if we were able to pull the usage but it isn't above the thresholds -- clear the status message
					updatePodNodeStorageCondition(node, "", "")
				}
			}

		}
	}
}

func refreshDiskWatermarkThresholds(clusterName, namespace string, client client.Client) {
	//quantity, err := resource.ParseQuantity(string)
	low, high, err := GetDiskWatermarks(clusterName, namespace, client)
	if err != nil {
		logrus.Debugf("Unable to refresh disk watermarks from cluster, using defaults")
	}

	switch low.(type) {
	case float64:
		value := low.(float64)
		DISK_WATERMARK_LOW_PCT = &value
		DISK_WATERMARK_LOW_ABS = nil
	case string:
		value, err := resource.ParseQuantity(strings.ToUpper(low.(string)))
		if err != nil {
			logrus.Warnf("Unable to parse %v: %v", low.(string), err)
		}
		DISK_WATERMARK_LOW_ABS = &value
		DISK_WATERMARK_LOW_PCT = nil
	default:
		// error
		logrus.Warnf("Unknown type for low: %T", low)
	}

	switch high.(type) {
	case float64:
		value := high.(float64)
		DISK_WATERMARK_HIGH_PCT = &value
		DISK_WATERMARK_HIGH_ABS = nil
	case string:
		value, err := resource.ParseQuantity(strings.ToUpper(high.(string)))
		if err != nil {
			logrus.Warnf("Unable to parse %v: %v", high.(string), err)
		}
		DISK_WATERMARK_HIGH_ABS = &value
		DISK_WATERMARK_HIGH_PCT = nil
	default:
		// error
		logrus.Warnf("Unknown type for high: %T", high)
	}

}

func exceedsLowWatermark(usage string, percent float64) bool {

	return exceedsWatermarks(usage, percent, DISK_WATERMARK_LOW_ABS, DISK_WATERMARK_LOW_PCT)
}

func exceedsHighWatermark(usage string, percent float64) bool {

	return exceedsWatermarks(usage, percent, DISK_WATERMARK_HIGH_ABS, DISK_WATERMARK_HIGH_PCT)
}

func exceedsWatermarks(usage string, percent float64, watermarkUsage *resource.Quantity, watermarkPercent *float64) bool {

	if usage == "" || percent < float64(0) {
		return false
	}

	quantity, err := resource.ParseQuantity(usage)
	if err != nil {
		logrus.Warnf("Unable to parse usage quantity %v: %v", usage, err)
		return false
	}

	// if quantity is > watermarkUsage and is used
	if watermarkUsage != nil && quantity.Cmp(*watermarkUsage) == 1 {
		return true
	}

	if watermarkPercent != nil && percent > *watermarkPercent {
		return true
	}

	return false
}

func updatePodCondition(node *api.ElasticsearchNodeStatus, condition *api.ClusterCondition) bool {
	if node.Conditions == nil {
		node.Conditions = make([]api.ClusterCondition, 0, 4)
	}

	// Try to find this node condition.
	conditionIndex, oldCondition := getPodCondition(node, condition.Type)

	if condition.Status == v1.ConditionFalse {
		if oldCondition != nil {
			node.Conditions = append(node.Conditions[:conditionIndex], node.Conditions[conditionIndex+1:]...)
			return true
		}

		return false
	}

	if oldCondition == nil {
		// We are adding new node condition.
		node.Conditions = append(node.Conditions, *condition)
		return true
	}

	isEqual := condition.Status == oldCondition.Status &&
		condition.Reason == oldCondition.Reason &&
		condition.Message == oldCondition.Message

	node.Conditions[conditionIndex] = *condition
	return !isEqual
}

func getPodCondition(node *api.ElasticsearchNodeStatus, conditionType api.ClusterConditionType) (int, *api.ClusterCondition) {
	if node == nil {
		return -1, nil
	}
	for i := range node.Conditions {
		if node.Conditions[i].Type == conditionType {
			return i, &node.Conditions[i]
		}
	}
	return -1, nil
}

func updatePodUnschedulableCondition(node *api.ElasticsearchNodeStatus, podCondition v1.PodCondition) bool {
	return updatePodCondition(node, &api.ClusterCondition{
		Type:               api.Unschedulable,
		Status:             podCondition.Status,
		Reason:             podCondition.Reason,
		Message:            podCondition.Message,
		LastTransitionTime: podCondition.LastTransitionTime,
	})
}

func updatePodNotReadyCondition(node *api.ElasticsearchNodeStatus, conditionType api.ClusterConditionType, reason, message string) bool {

	var status v1.ConditionStatus
	if message == "" && reason == "" {
		status = v1.ConditionFalse
	} else {
		status = v1.ConditionTrue
	}

	return updatePodCondition(node, &api.ClusterCondition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	})
}

func updatePodNodeStorageCondition(node *api.ElasticsearchNodeStatus, reason, message string) bool {

	var status v1.ConditionStatus
	if message == "" && reason == "" {
		status = v1.ConditionFalse
	} else {
		status = v1.ConditionTrue
	}

	return updatePodCondition(node, &api.ClusterCondition{
		Type:               api.NodeStorage,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	})
}

func updateStatusConditions(status *api.ElasticsearchStatus) {
	if status.Conditions == nil {
		status.Conditions = make([]api.ClusterCondition, 0, 4)
	}
	if _, condition := getESNodeCondition(status.Conditions, api.UpdatingSettings); condition == nil {
		updateUpdatingSettingsCondition(status, v1.ConditionFalse)
	}
	if _, condition := getESNodeCondition(status.Conditions, api.ScalingUp); condition == nil {
		updateScalingUpCondition(status, v1.ConditionFalse)
	}
	if _, condition := getESNodeCondition(status.Conditions, api.ScalingDown); condition == nil {
		updateScalingDownCondition(status, v1.ConditionFalse)
	}
	if _, condition := getESNodeCondition(status.Conditions, api.Restarting); condition == nil {
		updateRestartingCondition(status, v1.ConditionFalse)
	}
}

func isPodUnschedulableConditionTrue(conditions []api.ClusterCondition) bool {
	_, condition := getESNodeCondition(conditions, api.Unschedulable)
	return condition != nil && condition.Status == v1.ConditionTrue
}

func getESNodeCondition(conditions []api.ClusterCondition, conditionType api.ClusterConditionType) (int, *api.ClusterCondition) {
	if conditions == nil {
		return -1, nil
	}
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return i, &conditions[i]
		}
	}
	return -1, nil
}

func updateESNodeCondition(status *api.ElasticsearchStatus, condition *api.ClusterCondition) bool {
	condition.LastTransitionTime = metav1.Now()
	// Try to find this node condition.
	conditionIndex, oldCondition := getESNodeCondition(status.Conditions, condition.Type)

	if condition.Status == v1.ConditionFalse {
		if oldCondition != nil {
			status.Conditions = append(status.Conditions[:conditionIndex], status.Conditions[conditionIndex+1:]...)
			return true
		}

		return false
	}

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

func updateConditionWithRetry(dpl *api.Elasticsearch, value v1.ConditionStatus,
	executeUpdateCondition func(*api.ElasticsearchStatus, v1.ConditionStatus) bool, client client.Client) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		if getErr := client.Get(context.TODO(), types.NamespacedName{Name: dpl.Name, Namespace: dpl.Namespace}, dpl); getErr != nil {
			logrus.Debugf("Could not get Elasticsearch %v: %v", dpl.Name, getErr)
			return getErr
		}

		executeUpdateCondition(&dpl.Status, value)

		if updateErr := client.Update(context.TODO(), dpl); updateErr != nil {
			logrus.Debugf("Failed to update Elasticsearch %v status: %v", dpl.Name, updateErr)
			return updateErr
		}
		return nil
	})
	return retryErr
}

func updateInvalidMasterCountCondition(status *api.ElasticsearchStatus, value v1.ConditionStatus) bool {
	var message string
	var reason string
	if value == v1.ConditionTrue {
		message = fmt.Sprintf("Invalid master nodes count. Please ensure there are no more than %v total nodes with master roles", maxMasterCount)
		reason = "Invalid Settings"
	} else {
		message = ""
		reason = ""
	}
	return updateESNodeCondition(status, &api.ClusterCondition{
		Type:    api.InvalidMasters,
		Status:  value,
		Reason:  reason,
		Message: message,
	})
}

func updateInvalidDataCountCondition(status *api.ElasticsearchStatus, value v1.ConditionStatus) bool {
	var message string
	var reason string
	if value == v1.ConditionTrue {
		message = "No data nodes requested. Please ensure there is at least 1 node with data roles"
		reason = "Invalid Settings"
	} else {
		message = ""
		reason = ""
	}
	return updateESNodeCondition(status, &api.ClusterCondition{
		Type:    api.InvalidData,
		Status:  value,
		Reason:  reason,
		Message: message,
	})
}

func updateInvalidUUIDChangeCondition(cluster *api.Elasticsearch, value v1.ConditionStatus, message string, client client.Client) error {
	var reason string
	if value == v1.ConditionTrue {
		reason = "Invalid Spec"
	} else {
		reason = ""
	}

	return updateConditionWithRetry(
		cluster,
		value,
		func(status *api.ElasticsearchStatus, value v1.ConditionStatus) bool {
			return updateESNodeCondition(&cluster.Status, &api.ClusterCondition{
				Type:    api.InvalidUUID,
				Status:  value,
				Reason:  reason,
				Message: message,
			})
		},
		client,
	)
}

func updateInvalidReplicationCondition(status *api.ElasticsearchStatus, value v1.ConditionStatus) bool {
	var message string
	var reason string
	if value == v1.ConditionTrue {
		message = "Wrong RedundancyPolicy selected. Choose different RedundancyPolicy or add more nodes with data roles"
		reason = "Invalid Settings"
	} else {
		message = ""
		reason = ""
	}
	return updateESNodeCondition(status, &api.ClusterCondition{
		Type:    api.InvalidRedundancy,
		Status:  value,
		Reason:  reason,
		Message: message,
	})
}

func updateUpdatingSettingsCondition(status *api.ElasticsearchStatus, value v1.ConditionStatus) bool {
	return updateESNodeCondition(status, &api.ClusterCondition{
		Type:   api.UpdatingSettings,
		Status: value,
	})
}

func updateScalingUpCondition(status *api.ElasticsearchStatus, value v1.ConditionStatus) bool {
	return updateESNodeCondition(status, &api.ClusterCondition{
		Type:   api.ScalingUp,
		Status: value,
	})
}

func updateScalingDownCondition(status *api.ElasticsearchStatus, value v1.ConditionStatus) bool {
	return updateESNodeCondition(status, &api.ClusterCondition{
		Type:   api.ScalingDown,
		Status: value,
	})
}

func updateRestartingCondition(status *api.ElasticsearchStatus, value v1.ConditionStatus) bool {
	return updateESNodeCondition(status, &api.ClusterCondition{
		Type:   api.Restarting,
		Status: value,
	})
}
