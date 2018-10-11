package k8shandler

import (
	"fmt"

	v1alpha1 "github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
)

// UpdateStatus updates the status of Elasticsearch CRD
func (cState *ClusterState) UpdateStatus(dpl *v1alpha1.Elasticsearch) error {
	// TODO: add Elasticsearch cluster health
	// TODO: add Elasticsearch nodes list/roles
	// TODO: add configmap hash
	// TODO: add status of the cluster: i.e. is cluster restart in progress?
	// TODO: add secrets hash

	dpl.Status.Nodes = []v1alpha1.ElasticsearchNodeStatus{}
	for _, node := range cState.Nodes {
		//	logrus.Infof("Examining pod %v", pod)
		updateNodeStatus(node, &dpl.Status)
	}
	err := sdk.Update(dpl)
	if err != nil {
		return fmt.Errorf("failed to update Elasticsearch status: %v", err)
	}

	return nil
}

func updateNodeStatus(node *nodeState, dpl *v1alpha1.ElasticsearchStatus) {

	nodeStatus := v1alpha1.ElasticsearchNodeStatus{}
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
	dpl.Nodes = append(dpl.Nodes, nodeStatus)
}
