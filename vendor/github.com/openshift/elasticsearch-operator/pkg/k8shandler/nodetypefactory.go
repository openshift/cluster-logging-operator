package k8shandler

import (
	"fmt"

	api "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NodeTypeInterface interace represents individual Elasticsearch node
type NodeTypeInterface interface {
	populateReference(nodeName string, node api.ElasticsearchNode, cluster *api.Elasticsearch, roleMap map[api.ElasticsearchNodeRole]bool, replicas int32, client client.Client)
	state() api.ElasticsearchNodeStatus                      // this will get the current -- used for status
	create() error                                           // this will create the node in the case where it is new
	update(upgradeStatus *api.ElasticsearchNodeStatus) error // this will handle updates
	restart(upgradeStatus *api.ElasticsearchNodeStatus)
	progressUnshedulableNode(upgradeStatus *api.ElasticsearchNodeStatus) error
	name() string
	updateReference(node NodeTypeInterface)
	delete()
}

// NodeTypeFactory is a factory to construct either statefulset or deployment
type NodeTypeFactory func(name, namespace string) NodeTypeInterface

// this can potentially return a list if we have replicas > 1 for a data node
func (elasticsearchRequest *ElasticsearchRequest) GetNodeTypeInterface(uuid string, node api.ElasticsearchNode) []NodeTypeInterface {

	nodes := []NodeTypeInterface{}

	roleMap := getNodeRoleMap(node)

	cluster := elasticsearchRequest.cluster

	// common spec => cluster.Spec.Spec
	nodeName := fmt.Sprintf("%s-%s", cluster.Name, getNodeSuffix(uuid, roleMap))

	// if we have a data node then we need to create one deployment per replica
	if isDataNode(node) {
		// for loop from 1 to replica as replicaIndex
		//   it is 1 instead of 0 because of legacy code
		for replicaIndex := int32(1); replicaIndex <= node.NodeCount; replicaIndex++ {
			dataNodeName := addDataNodeSuffix(nodeName, replicaIndex)
			node := newDeploymentNode(dataNodeName, node, cluster, roleMap, elasticsearchRequest.client)
			nodes = append(nodes, node)
		}
	} else {
		node := newStatefulSetNode(nodeName, node, cluster, roleMap, elasticsearchRequest.client)
		nodes = append(nodes, node)
	}

	return nodes
}

func getNodeSuffix(uuid string, roleMap map[api.ElasticsearchNodeRole]bool) string {

	suffix := ""
	if roleMap[api.ElasticsearchRoleClient] {
		suffix = fmt.Sprintf("%s%s", suffix, "c")
	}

	if roleMap[api.ElasticsearchRoleData] {
		suffix = fmt.Sprintf("%s%s", suffix, "d")
	}

	if roleMap[api.ElasticsearchRoleMaster] {
		suffix = fmt.Sprintf("%s%s", suffix, "m")
	}

	return fmt.Sprintf("%s-%s", suffix, uuid)
}

func addDataNodeSuffix(nodeName string, replicaNumber int32) string {
	return fmt.Sprintf("%s-%d", nodeName, replicaNumber)
}

// newDeploymentNode constructs deploymentNode struct for data nodes
func newDeploymentNode(nodeName string, node api.ElasticsearchNode, cluster *api.Elasticsearch, roleMap map[api.ElasticsearchNodeRole]bool, client client.Client) NodeTypeInterface {
	deploymentNode := deploymentNode{}

	deploymentNode.populateReference(nodeName, node, cluster, roleMap, int32(1), client)

	return &deploymentNode
}

// newStatefulSetNode constructs statefulSetNode struct for non-data nodes
func newStatefulSetNode(nodeName string, node api.ElasticsearchNode, cluster *api.Elasticsearch, roleMap map[api.ElasticsearchNodeRole]bool, client client.Client) NodeTypeInterface {
	statefulSetNode := statefulSetNode{}

	statefulSetNode.populateReference(nodeName, node, cluster, roleMap, node.NodeCount, client)

	return &statefulSetNode
}

func containsNodeTypeInterface(node NodeTypeInterface, list []NodeTypeInterface) (int, bool) {
	for index, nodeTypeInterface := range list {
		if nodeTypeInterface.name() == node.name() {
			return index, true
		}
	}

	return -1, false
}
