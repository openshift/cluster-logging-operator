package k8shandler

import (
	"fmt"
	"strings"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	v1 "k8s.io/api/core/v1"

	api "github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func selectorForES(nodeRole string, clusterName string) map[string]string {

	return map[string]string{
		nodeRole:       "true",
		"cluster-name": clusterName,
	}
}

func labelsForESCluster(clusterName string) map[string]string {

	return map[string]string{
		"cluster-name": clusterName,
	}
}

func appendDefaultLabel(clusterName string, labels map[string]string) map[string]string {
	if _, ok := labels["cluster-name"]; ok {
		return labels
	}
	if labels == nil {
		labels = map[string]string{}
	}
	labels["cluster-name"] = clusterName
	return labels
}

func areSelectorsSame(lhs, rhs map[string]string) bool {

	if len(lhs) != len(rhs) {
		return false
	}

	for lhsKey, lhsVal := range lhs {
		rhsVal, ok := rhs[lhsKey]
		if !ok || lhsVal != rhsVal {
			return false
		}
	}

	return true
}

func mergeSelectors(nodeSelectors, commonSelectors map[string]string) map[string]string {

	if commonSelectors == nil {
		commonSelectors = make(map[string]string)
	}

	for k, v := range nodeSelectors {
		commonSelectors[k] = v
	}

	return commonSelectors
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []v1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

func getMasterCount(dpl *api.Elasticsearch) int32 {
	masterCount := int32(0)
	for _, node := range dpl.Spec.Nodes {
		if isMasterNode(node) {
			masterCount += node.NodeCount
		}
	}

	return masterCount
}

func getDataCount(dpl *api.Elasticsearch) int32 {
	dataCount := int32(0)
	for _, node := range dpl.Spec.Nodes {
		if isDataNode(node) {
			dataCount = dataCount + node.NodeCount
		}
	}

	return dataCount
}

func getClientCount(dpl *api.Elasticsearch) int32 {
	clientCount := int32(0)
	for _, node := range dpl.Spec.Nodes {
		if isClientNode(node) {
			clientCount = clientCount + node.NodeCount
		}
	}

	return clientCount
}

func isValidMasterCount(dpl *api.Elasticsearch) bool {
	masterCount := int(getMasterCount(dpl))
	return (masterCount <= maxMasterCount && masterCount > 0)
}

func isValidDataCount(dpl *api.Elasticsearch) bool {
	dataCount := int(getDataCount(dpl))
	return dataCount > 0
}

func isValidRedundancyPolicy(dpl *api.Elasticsearch) bool {
	dataCount := int(getDataCount(dpl))

	return !(dataCount == 1 && dpl.Spec.RedundancyPolicy == api.SingleRedundancy)
}

func isValidConf(dpl *api.Elasticsearch) error {
	if !isValidMasterCount(dpl) {
		if err := updateConditionWithRetry(dpl, v1.ConditionTrue, updateInvalidMasterCountCondition); err != nil {
			return err
		}
		return fmt.Errorf("Invalid master nodes count. Please ensure there are no more than %v total nodes with master roles", maxMasterCount)
	} else {
		if err := updateConditionWithRetry(dpl, v1.ConditionFalse, updateInvalidMasterCountCondition); err != nil {
			return err
		}
	}
	if !isValidDataCount(dpl) {
		if err := updateConditionWithRetry(dpl, v1.ConditionTrue, updateInvalidDataCountCondition); err != nil {
			return err
		}
		return fmt.Errorf("No data nodes requested. Please ensure there is at least 1 node with data roles")
	} else {
		if err := updateConditionWithRetry(dpl, v1.ConditionFalse, updateInvalidDataCountCondition); err != nil {
			return err
		}
	}
	if !isValidRedundancyPolicy(dpl) {
		if err := updateConditionWithRetry(dpl, v1.ConditionTrue, updateInvalidReplicationCondition); err != nil {
			return err
		}
		return fmt.Errorf("Wrong RedundancyPolicy selected. Choose different RedundancyPolicy or add more nodes with data roles")
	} else {
		if err := updateConditionWithRetry(dpl, v1.ConditionFalse, updateInvalidReplicationCondition); err != nil {
			return err
		}
	}

	if ok, msg := hasValidUUIDs(dpl); !ok {
		if err := updateInvalidUUIDChangeCondition(dpl, v1.ConditionTrue, msg); err != nil {
			return err
		}
		return fmt.Errorf("Unsupported change to UUIDs made: %v", msg)
	} else {
		if err := updateInvalidUUIDChangeCondition(dpl, v1.ConditionFalse, ""); err != nil {
			return err
		}
	}

	return nil
}

func hasValidUUIDs(dpl *api.Elasticsearch) (bool, string) {

	// TODO:
	// check that someone didn't update a uuid
	// check status.nodes[*].deploymentName for list of used uuids
	// deploymentName should match pattern {cluster.Name}-{uuid}[-replica]
	// if any in that list aren't found in spec.Nodes[*].GenUUID then someone did something bad...
	// somehow rollback the cluster object change and update message?
	// no way to rollback, but maybe maintain a last known "good state" and update SPEC to that?
	// update status message to be very descriptive of this

	prefix := fmt.Sprintf("%s-", dpl.Name)

	var knownUUIDs []string
	for _, node := range dpl.Status.Nodes {

		var nodeName string
		if node.DeploymentName != "" {
			nodeName = node.DeploymentName
		}

		if node.StatefulSetName != "" {
			nodeName = node.StatefulSetName
		}

		parts := strings.Split(strings.TrimPrefix(nodeName, prefix), "-")

		if len(parts) < 2 {
			return false, fmt.Sprintf("Invalid name found for %q", nodeName)
		}

		uuid := parts[1]

		if !sliceContainsString(knownUUIDs, uuid) {
			knownUUIDs = append(knownUUIDs, uuid)
		}
	}

	// make sure all known UUIDs are found amongst spec.nodes[*].genuuid
	for _, uuid := range knownUUIDs {
		if !isUUIDFound(uuid, dpl.Spec.Nodes) {
			return false, fmt.Sprintf("Previously used GenUUID %q is no longer found in Spec.Nodes", uuid)
		}
	}

	return true, ""
}

func isUUIDFound(uuid string, nodes []api.ElasticsearchNode) bool {

	for _, node := range nodes {

		if node.GenUUID != nil {
			if *node.GenUUID == uuid {
				return true
			}
		}
	}

	return false
}

func sliceContainsString(slice []string, value string) bool {

	for _, s := range slice {
		if value == s {
			return true
		}
	}

	return false
}

func DeletePod(podName, namespace string) error {
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
		},
	}

	err := sdk.Delete(pod)

	return err
}

func GetPodList(namespace string, selector string) (*v1.PodList, error) {
	list := &v1.PodList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: v1.SchemeGroupVersion.String(),
		},
	}

	err := sdk.List(
		namespace,
		list,
		sdk.WithListOptions(&metav1.ListOptions{
			LabelSelector: selector,
		}),
	)

	return list, err
}
