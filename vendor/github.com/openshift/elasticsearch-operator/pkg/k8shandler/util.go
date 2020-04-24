package k8shandler

import (
	"context"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	api "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
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

func areTolerationsSame(lhs, rhs []v1.Toleration) bool {
	if len(lhs) != len(rhs) {
		return false
	}

	for _, lhsToleration := range lhs {
		if !containsToleration(lhsToleration, rhs) {
			return false
		}
	}

	return true
}

func containsToleration(toleration v1.Toleration, tolerations []v1.Toleration) bool {
	for _, t := range tolerations {
		if isTolerationSame(t, toleration) {
			return true
		}
	}

	return false
}

func isTolerationSame(lhs, rhs v1.Toleration) bool {

	tolerationSecondsBool := false
	// check that both are either null or not null
	if (lhs.TolerationSeconds == nil) == (rhs.TolerationSeconds == nil) {
		if lhs.TolerationSeconds != nil {
			// only compare values (attempt to dereference) if pointers aren't nil
			tolerationSecondsBool = (*lhs.TolerationSeconds == *rhs.TolerationSeconds)
		} else {
			tolerationSecondsBool = true
		}
	}

	return (lhs.Key == rhs.Key) &&
		(lhs.Operator == rhs.Operator) &&
		(lhs.Value == rhs.Value) &&
		(lhs.Effect == rhs.Effect) &&
		tolerationSecondsBool
}

func appendTolerations(nodeTolerations, commonTolerations []v1.Toleration) []v1.Toleration {
	if commonTolerations == nil {
		commonTolerations = []v1.Toleration{}
	}

	return append(commonTolerations, nodeTolerations...)
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

	if len(dpl.Spec.Nodes) == 0 {
		return true
	}

	masterCount := int(getMasterCount(dpl))
	return (masterCount <= maxMasterCount && masterCount > 0)
}

func isValidDataCount(dpl *api.Elasticsearch) bool {

	if len(dpl.Spec.Nodes) == 0 {
		return true
	}

	dataCount := int(getDataCount(dpl))
	return dataCount > 0
}

func isValidRedundancyPolicy(dpl *api.Elasticsearch) bool {
	dataCount := int(getDataCount(dpl))

	switch dpl.Spec.RedundancyPolicy {
	case "":
	case api.ZeroRedundancy:
	case api.SingleRedundancy:
	case api.MultipleRedundancy:
	case api.FullRedundancy:
	default:
		return false
	}

	return !(dataCount == 1 && dpl.Spec.RedundancyPolicy == api.SingleRedundancy)
}

func (elasticsearchRequest *ElasticsearchRequest) isValidConf() error {

	dpl := elasticsearchRequest.cluster
	client := elasticsearchRequest.client

	if !isValidMasterCount(dpl) {
		if err := updateConditionWithRetry(dpl, v1.ConditionTrue, updateInvalidMasterCountCondition, client); err != nil {
			return err
		}
		return fmt.Errorf("Invalid master nodes count. Please ensure there are no more than %v total nodes with master roles", maxMasterCount)
	} else {
		if err := updateConditionWithRetry(dpl, v1.ConditionFalse, updateInvalidMasterCountCondition, client); err != nil {
			return err
		}
	}

	if !isValidDataCount(dpl) {
		if err := updateConditionWithRetry(dpl, v1.ConditionTrue, updateInvalidDataCountCondition, client); err != nil {
			return err
		}
		return fmt.Errorf("No data nodes requested. Please ensure there is at least 1 node with data roles")
	} else {
		if err := updateConditionWithRetry(dpl, v1.ConditionFalse, updateInvalidDataCountCondition, client); err != nil {
			return err
		}
	}

	if !isValidRedundancyPolicy(dpl) {
		if err := updateConditionWithRetry(dpl, v1.ConditionTrue, updateInvalidReplicationCondition, client); err != nil {
			return err
		}
		return fmt.Errorf("Wrong RedundancyPolicy selected '%s'. Choose different RedundancyPolicy or add more nodes with data roles", dpl.Spec.RedundancyPolicy)
	} else {
		if err := updateConditionWithRetry(dpl, v1.ConditionFalse, updateInvalidReplicationCondition, client); err != nil {
			return err
		}
	}

	if ok, msg := hasValidUUIDs(dpl); !ok {
		if err := updateInvalidUUIDChangeCondition(dpl, v1.ConditionTrue, msg, client); err != nil {
			return err
		}
		return fmt.Errorf("Unsupported change to UUIDs made: %v", msg)
	} else {
		if err := updateInvalidUUIDChangeCondition(dpl, v1.ConditionFalse, "", client); err != nil {
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

func DeletePod(podName, namespace string, client client.Client) error {
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

	err := client.Delete(context.TODO(), pod)

	return err
}

func GetPodList(namespace string, selector map[string]string, sdkClient client.Client) (*v1.PodList, error) {
	list := &v1.PodList{}

	labelSelector := labels.SelectorFromSet(selector)

	err := sdkClient.List(
		context.TODO(),
		&client.ListOptions{Namespace: namespace, LabelSelector: labelSelector},
		list,
	)

	return list, err
}
