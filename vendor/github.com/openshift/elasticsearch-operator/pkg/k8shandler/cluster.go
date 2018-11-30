package k8shandler

import (
	"fmt"

	apps "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
)

// ClusterState struct represents the state of the cluster
type ClusterState struct {
	Nodes                []*nodeState
	DanglingStatefulSets *apps.StatefulSetList
	DanglingDeployments  *apps.DeploymentList
	DanglingReplicaSets  *apps.ReplicaSetList
	DanglingPods         *v1.PodList
}

var wrongConfig bool

// CreateOrUpdateElasticsearchCluster creates an Elasticsearch deployment
func CreateOrUpdateElasticsearchCluster(dpl *v1alpha1.Elasticsearch, configMapName, serviceAccountName string) error {

	cState, err := NewClusterState(dpl, configMapName, serviceAccountName)
	if err != nil {
		return err
	}

	// Verify that we didn't scale up too many masters
	if !isValidMasterCount(dpl) {
		if wrongConfig {
			return nil
		}
		wrongConfig = true
		return fmt.Errorf("Invalid master count. Please ensure there are no more than %v total nodes with master roles", MAX_MASTER_COUNT)
	} else {
		wrongConfig = false
	}

	action, err := cState.getRequiredAction(dpl)
	if err != nil {
		return err
	}

	switch {
	case action == v1alpha1.ElasticsearchActionNewClusterNeeded:
		err = cState.buildNewCluster(asOwner(dpl))
		if err != nil {
			return err
		}
	case action == v1alpha1.ElasticsearchActionScaleDownNeeded:
		err = cState.removeStaleNodes()
		if err != nil {
			return err
		}
	case action == v1alpha1.ElasticsearchActionRollingRestartNeeded:
		// TODO: change this to do the actual rolling restart
		err = cState.buildNewCluster(asOwner(dpl))
		if err != nil {
			return err
		}
	case action == v1alpha1.ElasticsearchActionNone:
		break
	default:
		return fmt.Errorf("Unknown cluster action requested: %v", action)
	}
	// Scrape cluster health from elasticsearch every time
	err = cState.UpdateStatus(dpl)
	if err != nil {
		return err
	}
	return nil
}

// NewClusterState func generates ClusterState for the current cluster
func NewClusterState(dpl *v1alpha1.Elasticsearch, configMapName, serviceAccountName string) (ClusterState, error) {
	nodes := []*nodeState{}
	cState := ClusterState{
		Nodes: nodes,
	}

	numMasters := getMasterCount(dpl)
	numDatas := getDataCount(dpl)

	var i int32
	for nodeNum, node := range dpl.Spec.Nodes {

		for i = 1; i <= node.Replicas; i++ {
			nodeCfg, err := constructNodeSpec(dpl, node, configMapName, serviceAccountName, int32(nodeNum), i, numMasters, numDatas)
			if err != nil {
				return cState, fmt.Errorf("Unable to construct ES node config %v", err)
			}

			node := nodeState{
				Desired: nodeCfg,
			}
			cState.Nodes = append(cState.Nodes, &node)
		}
	}

	err := cState.amendDeployments(dpl)
	if err != nil {
		return cState, fmt.Errorf("Unable to amend Deployments to status: %v", err)
	}

	err = cState.amendStatefulSets(dpl)
	if err != nil {
		return cState, fmt.Errorf("Unable to amend StatefulSets to status: %v", err)
	}

	err = cState.amendReplicaSets(dpl)
	if err != nil {
		return cState, fmt.Errorf("Unable to amend ReplicaSets to status: %v", err)
	}

	err = cState.amendPods(dpl)
	if err != nil {
		return cState, fmt.Errorf("Unable to amend Pods to status: %v", err)
	}

	return cState, nil
}

// getRequiredAction checks the desired state against what's present in current
// deployments/statefulsets/pods
func (cState *ClusterState) getRequiredAction(dpl *v1alpha1.Elasticsearch) (v1alpha1.ElasticsearchRequiredAction, error) {
	// TODO: Add condition that if an operation is currently in progress
	// not to try to queue another action. Instead return ElasticsearchActionInProgress which
	// is noop.

	// TODO: Handle failures. Maybe introduce some ElasticsearchCondition which says
	// what action was attempted last, when, how many tries and what the result is.

	// TODO: implement better logic to understand when new cluster is needed
	// maybe RequiredAction ElasticsearchActionNewClusterNeeded should be renamed to
	// ElasticsearchActionAddNewNodes - will blindly add new nodes to the cluster.

	if dpl.Spec.ManagementState == v1alpha1.ManagementStateManaged {

		for _, node := range cState.Nodes {
			if node.Actual.Deployment == nil && node.Actual.StatefulSet == nil {
				return v1alpha1.ElasticsearchActionNewClusterNeeded, nil
			}
		}

		// TODO: implement rolling restart action if any deployment/configmap actually deployed
		// is different from the desired.
		for _, node := range cState.Nodes {
			if node.Desired.IsUpdateNeeded() {
				return v1alpha1.ElasticsearchActionRollingRestartNeeded, nil
			}
		}

		// If some deployments exist that are not specified in CR, they'll be in DanglingDeployments
		// we need to remove those to comply with the desired cluster structure.
		if cState.DanglingDeployments != nil {
			return v1alpha1.ElasticsearchActionScaleDownNeeded, nil
		}
	}

	return v1alpha1.ElasticsearchActionNone, nil
}

func (cState *ClusterState) buildNewCluster(owner metav1.OwnerReference) error {
	for _, node := range cState.Nodes {
		err := node.Desired.CreateOrUpdateNode(owner)
		if err != nil {
			return fmt.Errorf("Unable to create Elasticsearch node: %v", err)
		}
	}
	return nil
}

// list existing StatefulSets and delete those unmanaged by the operator
func (cState *ClusterState) removeStaleNodes() error {
	for _, node := range cState.DanglingDeployments.Items {
		// the returned deployment doesn't have TypeMeta, so we're adding it.
		node.TypeMeta = metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		}
		err := sdk.Delete(&node)
		if err != nil {
			return fmt.Errorf("Unable to delete resource %v: ", err)
		}
	}
	return nil
}

func (cState *ClusterState) amendStatefulSets(dpl *v1alpha1.Elasticsearch) error {
	statefulSets, err := listStatefulSets(dpl.Name, dpl.Namespace)
	if err != nil {
		return fmt.Errorf("Unable to list Elasticsearch's StatefulSets: %v", err)
	}

	var element apps.StatefulSet
	var ok bool

	for _, node := range cState.Nodes {
		statefulSets, element, ok = popStatefulSet(statefulSets, node.Desired)
		if ok {
			node.setStatefulSet(element)
		}
	}
	if len(statefulSets.Items) != 0 {
		cState.DanglingStatefulSets = statefulSets
	}
	return nil
}

func (cState *ClusterState) amendDeployments(dpl *v1alpha1.Elasticsearch) error {
	deployments, err := listDeployments(dpl.Name, dpl.Namespace)
	if err != nil {
		return fmt.Errorf("Unable to list Elasticsearch's Deployments: %v", err)
	}

	var element apps.Deployment
	var ok bool

	for _, node := range cState.Nodes {
		deployments, element, ok = popDeployment(deployments, node.Desired)
		if ok {
			node.setDeployment(element)
		}
	}
	if len(deployments.Items) != 0 {
		cState.DanglingDeployments = deployments
	}
	return nil
}

func (cState *ClusterState) amendReplicaSets(dpl *v1alpha1.Elasticsearch) error {
	replicaSets, err := listReplicaSets(dpl.Name, dpl.Namespace)
	if err != nil {
		return fmt.Errorf("Unable to list Elasticsearch's ReplicaSets: %v", err)
	}
	var replicaSet apps.ReplicaSet

	for _, node := range cState.Nodes {
		var ok bool
		replicaSets, replicaSet, ok = popReplicaSet(replicaSets, node.Actual)
		if ok {
			node.setReplicaSet(replicaSet)
		}
	}

	if len(replicaSets.Items) != 0 {
		cState.DanglingReplicaSets = replicaSets
	}
	return nil
}

func (cState *ClusterState) amendPods(dpl *v1alpha1.Elasticsearch) error {
	pods, err := listPods(dpl.Name, dpl.Namespace)
	if err != nil {
		return fmt.Errorf("Unable to list Elasticsearch's Pods: %v", err)
	}
	var pod v1.Pod

	for _, node := range cState.Nodes {
		var ok bool
		pods, pod, ok = popPod(pods, node.Actual)
		if ok {
			node.setPod(pod)
		}
	}

	if len(pods.Items) != 0 {
		cState.DanglingPods = pods
	}
	return nil
}
