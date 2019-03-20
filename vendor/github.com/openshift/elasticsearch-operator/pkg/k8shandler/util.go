package k8shandler

import (
	"fmt"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"k8s.io/api/core/v1"

	api "github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
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
		return fmt.Errorf("Invalid master nodes count. Please ensure there are no more than %v total nodes with master roles", maxMasterCount)
	}
	if !isValidDataCount(dpl) {
		return fmt.Errorf("No data nodes requested. Please ensure there is at least 1 node with data roles")
	}
	if !isValidRedundancyPolicy(dpl) {
		return fmt.Errorf("Wrong RedundancyPolicy selected. Choose different RedundancyPolicy or add more nodes with data roles")
	}
	return nil
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
