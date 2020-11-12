package k8shandler

import (
	"context"
	"fmt"
	"reflect"

	"github.com/openshift/elasticsearch-operator/pkg/utils"
	"github.com/openshift/elasticsearch-operator/pkg/utils/comparators"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"

	api "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
)

const expectedMinVersion = "6.0"

var wrongConfig bool
var nodes map[string][]NodeTypeInterface

var aliasNeededMap map[string]bool

func FlushNodes(clusterName, namespace string) {
	nodes[nodeMapKey(clusterName, namespace)] = []NodeTypeInterface{}
}

func nodeMapKey(clusterName, namespace string) string {
	return fmt.Sprintf("%v-%v", clusterName, namespace)
}

// CreateOrUpdateElasticsearchCluster creates an Elasticsearch deployment
func (elasticsearchRequest *ElasticsearchRequest) CreateOrUpdateElasticsearchCluster() error {
	esClient := elasticsearchRequest.esClient

	// Verify that we didn't scale up too many masters
	err := elasticsearchRequest.isValidConf()
	if err != nil {
		// if wrongConfig=true then we've already print out error message
		// don't flood the stderr of the operator with the same message
		if wrongConfig {
			return nil
		}
		wrongConfig = true
		return err
	}
	wrongConfig = false

	// Populate nodes from the custom resources spec.nodes
	if err := elasticsearchRequest.populateNodes(); err != nil {
		return err
	}

	//clearing transient setting because of a bug in earlier releases which
	//may leave the shard allocation in an undesirable state
	if success, err := esClient.ClearTransientShardAllocation(); !success {
		logrus.Warnf("Unable to clear transient shard allocation for %q %q: %v", elasticsearchRequest.cluster.Namespace, elasticsearchRequest.cluster.Namespace, err)
	}

	// Update the cluster status immediately to refresh status.nodes
	// before progressing with any unschedulable nodes.
	// Ensures that deleted nodes are removed from status.nodes.
	if err := elasticsearchRequest.UpdateClusterStatus(); err != nil {
		return err
	}
	progressUnshedulableNodes(elasticsearchRequest.cluster)
	err = elasticsearchRequest.performCertClusterRestart()
	if err != nil {
		return elasticsearchRequest.UpdateClusterStatus()
	}

	// if there is a node currently being upgraded, work on that first
	upgradeInProgressNode := getNodeUpgradeInProgress(elasticsearchRequest.cluster)
	scheduledUpgradeNodes := getScheduledUpgradeNodes(elasticsearchRequest.cluster)
	if upgradeInProgressNode != nil {

		clusterStatus := elasticsearchRequest.cluster.Status.DeepCopy()
		_, nodeStatus := getNodeStatus(upgradeInProgressNode.name(), clusterStatus)

		if _, ok := containsNodeTypeInterface(upgradeInProgressNode, scheduledUpgradeNodes); ok {
			logrus.Debugf("Continuing update for %v", upgradeInProgressNode.name())
			if err := upgradeInProgressNode.update(nodeStatus); err != nil {
				logrus.Errorf("unable to update node. E: %s", err.Error())
			}
		} else {
			logrus.Debugf("Continuing restart for %v", upgradeInProgressNode.name())
			upgradeInProgressNode.rollingRestart(nodeStatus)
		}

		addNodeState(upgradeInProgressNode, nodeStatus)
		if err := elasticsearchRequest.setNodeStatus(upgradeInProgressNode, nodeStatus, clusterStatus); err != nil {
			logrus.Errorf("unable to update node status: E: %s", err)
		}

	} else {

		if len(scheduledUpgradeNodes) > 0 {

			// get the current ES version
			version, err := esClient.GetLowestClusterVersion()
			if err != nil {
				// this can be because we couldn't get a valid response from ES
				logrus.Warnf("when trying to get LowestClusterVersion: %v", err)
			} else {

				logrus.Debugf("Found current cluster version to be %q", version)
				comparison := comparators.CompareVersions(version, expectedMinVersion)

				// if it is < what we expect (6.0) then do full cluster restart:
				if comparison > 0 {
					// perform a full cluster restart
					if err := elasticsearchRequest.fullClusterRestart(scheduledUpgradeNodes); err != nil {
						logrus.Warnf("when trying to perform full cluster restart: %v", err)
					}

				} else {

					for _, node := range scheduledUpgradeNodes {
						logrus.Debugf("Perform a update for %v", node.name())
						clusterStatus := elasticsearchRequest.cluster.Status.DeepCopy()
						_, nodeStatus := getNodeStatus(node.name(), clusterStatus)
						err := node.update(nodeStatus)
						addNodeState(node, nodeStatus)
						if err := elasticsearchRequest.setNodeStatus(node, nodeStatus, clusterStatus); err != nil {
							logrus.Error(err)
						}
						if err != nil {
							logrus.Warnf("Error occurred while updating node %v: %v", node.name(), err)
						}
					}
				}
			}

		} else {

			scheduledRedeployNodes := getScheduledRedeployOnlyNodes(elasticsearchRequest.cluster)
			if len(scheduledRedeployNodes) > 0 {
				// get all nodes that need only a rollout
				// TODO: ready cluster for a pod restart first
				for _, node := range scheduledRedeployNodes {
					logrus.Debugf("Perform a redeploy for %v", node.name())
					clusterStatus := elasticsearchRequest.cluster.Status.DeepCopy()
					_, nodeStatus := getNodeStatus(node.name(), clusterStatus)

					node.rollingRestart(nodeStatus)

					addNodeState(node, nodeStatus)
					if err := elasticsearchRequest.setNodeStatus(node, nodeStatus, clusterStatus); err != nil {
						logrus.Errorf("unable to set node status. E: %s", err.Error())
					}
				}

			} else {

				for _, node := range nodes[nodeMapKey(elasticsearchRequest.cluster.Name, elasticsearchRequest.cluster.Namespace)] {
					clusterStatus := elasticsearchRequest.cluster.Status.DeepCopy()
					_, nodeStatus := getNodeStatus(node.name(), clusterStatus)

					// Verify that we didn't scale up too many masters
					if err := elasticsearchRequest.isValidConf(); err != nil {
						// if wrongConfig=true then we've already print out error message
						// don't flood the stderr of the operator with the same message
						if wrongConfig {
							return nil
						}
						wrongConfig = true
						return err
					}

					if err := node.create(); err != nil {
						return err
					}

					addNodeState(node, nodeStatus)
					if err := elasticsearchRequest.setNodeStatus(node, nodeStatus, clusterStatus); err != nil {
						logrus.Errorf("unable to set node status. E: %s", err.Error())
					}

					elasticsearchRequest.updateMinMasters()
				}

				// ensure we always have shard allocation to All if we aren't doing an upgrade...
				if ok, err := esClient.SetShardAllocation(api.ShardAllocationAll); !ok {
					logrus.Warnf("Unable to enable shard allocation: %v", err)
				}

				// we only want to update our replicas if we aren't in the middle up an upgrade
				if err := esClient.UpdateReplicaCount(int32(calculateReplicaCount(elasticsearchRequest.cluster))); err != nil {
					logrus.Error(err)
				}
				if aliasNeededMap == nil {
					aliasNeededMap = make(map[string]bool)
				}
				if val, ok := aliasNeededMap[nodeMapKey(elasticsearchRequest.cluster.Name, elasticsearchRequest.cluster.Namespace)]; !ok || val {
					// add alias to old indices if they exist and don't have one
					// this should be removed after one release...
					successful := esClient.AddAliasForOldIndices()

					if successful {
						aliasNeededMap[nodeMapKey(elasticsearchRequest.cluster.Name, elasticsearchRequest.cluster.Namespace)] = false
					}

				}
			}
		}
	}

	// Scrape cluster health from elasticsearch every time
	return elasticsearchRequest.UpdateClusterStatus()
}

func (elasticsearchRequest *ElasticsearchRequest) updateMinMasters() {
	// do as best effort -- whenever we create a node update min masters (if required)

	cluster := elasticsearchRequest.cluster
	esClient := elasticsearchRequest.esClient

	currentMasterCount, err := esClient.GetMinMasterNodes()
	if err != nil {
		logrus.Debugf("Unable to get current min master count for cluster %v", cluster.Name)
	}

	desiredMasterCount := getMasterCount(cluster)/2 + 1
	currentNodeCount, err := esClient.GetClusterNodeCount()
	if err != nil {
		logrus.Warnf("unable to get cluster node count. E: %s\r\n", err.Error())
	}

	// check that we have the required number of master nodes in the cluster...
	if currentNodeCount >= desiredMasterCount {
		if currentMasterCount != desiredMasterCount {
			if _, setErr := esClient.SetMinMasterNodes(desiredMasterCount); setErr != nil {
				logrus.Debugf("Unable to set min master count to %d for cluster %v", desiredMasterCount, cluster.Name)
			}
		}
	}
}

func getNodeUpgradeInProgress(cluster *api.Elasticsearch) NodeTypeInterface {
	for _, node := range cluster.Status.Nodes {
		if node.UpgradeStatus.UnderUpgrade == v1.ConditionTrue {
			for _, nodeTypeInterface := range nodes[nodeMapKey(cluster.Name, cluster.Namespace)] {
				if node.DeploymentName == nodeTypeInterface.name() ||
					node.StatefulSetName == nodeTypeInterface.name() {
					return nodeTypeInterface
				}
			}
		}
	}

	return nil
}

func progressUnshedulableNodes(cluster *api.Elasticsearch) {
	clusterNodes := nodes[nodeMapKey(cluster.GetName(), cluster.GetNamespace())]

	for _, nodeStatus := range cluster.Status.Nodes {
		if isPodUnschedulableConditionTrue(nodeStatus.Conditions) ||
			isPodImagePullBackOff(nodeStatus.Conditions) {

			for _, node := range clusterNodes {
				if nodeStatus.DeploymentName == node.name() || nodeStatus.StatefulSetName == node.name() {
					if node.isMissing() {
						logrus.Infof("Unschedulable node does not have k8s resource, skipping node %s", node.name())
						continue
					}

					logrus.Debugf("Node %s is unschedulable, trying to recover...", node.name())
					if err := node.progressUnshedulableNode(&nodeStatus); err != nil {
						logrus.Warnf("Failed to progress update of unschedulable node '%s': %v", node.name(), err)
					}
				}
			}
		}
	}
}

func (elasticsearchRequest *ElasticsearchRequest) setUUIDs() {

	cluster := elasticsearchRequest.cluster

	for index := 0; index < len(cluster.Spec.Nodes); index++ {
		if cluster.Spec.Nodes[index].GenUUID == nil {
			uuid, err := utils.RandStringBytes(8)
			if err != nil {
				continue
			}

			elasticsearchRequest.setUUID(index, uuid)
		}
	}
}

func (er *ElasticsearchRequest) setUUID(index int, uuid string) {

	nretries := -1
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		nretries++
		if err := er.client.Get(context.TODO(), types.NamespacedName{Name: er.cluster.Name, Namespace: er.cluster.Namespace}, er.cluster); err != nil {
			// FIXME: return structured error
			logrus.Infof("Could not get Elasticsearch cluster %q in namespace %q: %v", er.cluster.Name, er.cluster.Namespace, err)
			return err
		}

		if er.cluster.Spec.Nodes[index].GenUUID != nil {
			return nil
		}

		er.cluster.Spec.Nodes[index].GenUUID = &uuid

		if updateErr := er.client.Update(context.TODO(), er.cluster); updateErr != nil {
			// FIXME: return structured error
			logrus.Infof("Failed to update Elasticsearch cluster %q in namespace %q status, trying again.\nError: %v", er.cluster.Name, er.cluster.Namespace, updateErr)
			return updateErr
		}
		return nil
	})

	if err != nil {
		logrus.Errorf("Could not update CR for Elasticsearch cluster %q in namespace %q after %v retries: %v", er.cluster.Name, er.cluster.Namespace, nretries, err)
	} else {
		logrus.Infof("Updated Elasticsearch cluster %q in namespace %q after %v retries", er.cluster.Name, er.cluster.Namespace, nretries)
	}

}

func (elasticsearchRequest *ElasticsearchRequest) populateNodes() error {
	if err := elasticsearchRequest.recoverOrphanedCluster(); err != nil {
		return err
	}
	elasticsearchRequest.setUUIDs()

	if nodes == nil {
		nodes = make(map[string][]NodeTypeInterface)
	}

	cluster := elasticsearchRequest.cluster
	currentNodes := []NodeTypeInterface{}

	// get list of client only nodes, and collapse node info into the node (self field) if needed
	for _, node := range cluster.Spec.Nodes {

		// build the NodeTypeInterface list
		for _, nodeTypeInterface := range elasticsearchRequest.GetNodeTypeInterface(*node.GenUUID, node) {

			nodeIndex, ok := containsNodeTypeInterface(nodeTypeInterface, nodes[nodeMapKey(cluster.Name, cluster.Namespace)])
			if !ok {
				currentNodes = append(currentNodes, nodeTypeInterface)
			} else {
				nodes[nodeMapKey(cluster.Name, cluster.Namespace)][nodeIndex].updateReference(nodeTypeInterface)
				currentNodes = append(currentNodes, nodes[nodeMapKey(cluster.Name, cluster.Namespace)][nodeIndex])
			}

		}
	}

	minMasterUpdated := false

	// we want to only keep nodes that were generated and purge/delete any other ones...
	for _, node := range nodes[nodeMapKey(cluster.Name, cluster.Namespace)] {
		if _, ok := containsNodeTypeInterface(node, currentNodes); !ok {
			if !minMasterUpdated {
				// if we're removing a node make sure we set a lower min masters to keep cluster functional
				elasticsearchRequest.updateMinMasters()
				minMasterUpdated = true
			}
			if err := node.delete(); err != nil {
				logrus.Warnf("unable to delete node. E: %s\r\n", err.Error())
			}

			// remove from status.Nodes
			if index, _ := getNodeStatus(node.name(), &cluster.Status); index != NOT_FOUND_INDEX {
				cluster.Status.Nodes = append(cluster.Status.Nodes[:index], cluster.Status.Nodes[index+1:]...)
			}
		}
	}

	nodes[nodeMapKey(cluster.Name, cluster.Namespace)] = currentNodes

	return nil
}

func getScheduledUpgradeNodes(cluster *api.Elasticsearch) []NodeTypeInterface {
	upgradeNodes := []NodeTypeInterface{}

	for _, node := range cluster.Status.Nodes {
		if node.UpgradeStatus.ScheduledForUpgrade == v1.ConditionTrue {
			for _, nodeTypeInterface := range nodes[nodeMapKey(cluster.Name, cluster.Namespace)] {
				if node.DeploymentName == nodeTypeInterface.name() ||
					node.StatefulSetName == nodeTypeInterface.name() {
					upgradeNodes = append(upgradeNodes, nodeTypeInterface)
				}
			}
		}
	}

	return upgradeNodes
}

func getScheduledRedeployOnlyNodes(cluster *api.Elasticsearch) []NodeTypeInterface {
	redeployNodes := []NodeTypeInterface{}

	for _, node := range cluster.Status.Nodes {
		if node.UpgradeStatus.ScheduledForRedeploy == v1.ConditionTrue &&
			(node.UpgradeStatus.ScheduledForUpgrade == v1.ConditionFalse ||
				node.UpgradeStatus.ScheduledForUpgrade == "") {
			for _, nodeTypeInterface := range nodes[nodeMapKey(cluster.Name, cluster.Namespace)] {
				if node.DeploymentName == nodeTypeInterface.name() ||
					node.StatefulSetName == nodeTypeInterface.name() {
					redeployNodes = append(redeployNodes, nodeTypeInterface)
				}
			}
		}
	}

	return redeployNodes
}

func getScheduledCertRedeployNodes(cluster *api.Elasticsearch) []NodeTypeInterface {
	redeployCertNodes := []NodeTypeInterface{}
	dataNodes := []NodeTypeInterface{}

	for _, node := range cluster.Status.Nodes {
		if node.UpgradeStatus.ScheduledForCertRedeploy == v1.ConditionTrue {
			for _, nodeTypeInterface := range nodes[nodeMapKey(cluster.Name, cluster.Namespace)] {
				if node.DeploymentName == nodeTypeInterface.name() {
					dataNodes = append(dataNodes, nodeTypeInterface)
				}

				if node.StatefulSetName == nodeTypeInterface.name() {
					redeployCertNodes = append(redeployCertNodes, nodeTypeInterface)
				}
			}
		}
	}

	redeployCertNodes = append(redeployCertNodes, dataNodes...)

	return redeployCertNodes
}

func addNodeState(node NodeTypeInterface, nodeStatus *api.ElasticsearchNodeStatus) {

	nodeState := node.state()

	nodeStatus.UpgradeStatus.ScheduledForUpgrade = nodeState.UpgradeStatus.ScheduledForUpgrade
	nodeStatus.UpgradeStatus.ScheduledForRedeploy = nodeState.UpgradeStatus.ScheduledForRedeploy
	nodeStatus.UpgradeStatus.ScheduledForCertRedeploy = nodeState.UpgradeStatus.ScheduledForCertRedeploy
	nodeStatus.DeploymentName = nodeState.DeploymentName
	nodeStatus.StatefulSetName = nodeState.StatefulSetName
}

func (elasticsearchRequest *ElasticsearchRequest) setNodeStatus(node NodeTypeInterface, nodeStatus *api.ElasticsearchNodeStatus, clusterStatus *api.ElasticsearchStatus) error {

	index, _ := getNodeStatus(node.name(), clusterStatus)

	if index == NOT_FOUND_INDEX {
		clusterStatus.Nodes = append(clusterStatus.Nodes, *nodeStatus)
	} else {
		clusterStatus.Nodes[index] = *nodeStatus
	}

	return elasticsearchRequest.updateNodeStatus(*clusterStatus)
}

func (elasticsearchRequest *ElasticsearchRequest) updateNodeStatus(status api.ElasticsearchStatus) error {

	cluster := elasticsearchRequest.cluster
	// if there is nothing to update, don't
	if reflect.DeepEqual(cluster.Status, status) {
		return nil
	}

	nretries := -1
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		nretries++
		if getErr := elasticsearchRequest.client.Get(context.TODO(), types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}, cluster); getErr != nil {
			logrus.Debugf("Could not get Elasticsearch %v: %v", cluster.Name, getErr)
			return getErr
		}

		cluster.Status = status

		if updateErr := elasticsearchRequest.client.Update(context.TODO(), cluster); updateErr != nil {
			logrus.Debugf("Failed to update Elasticsearch %s status. Reason: %v. Trying again...", cluster.Name, updateErr)
			return updateErr
		}

		logrus.Debugf("Updated Elasticsearch %v after %v retries", cluster.Name, nretries)
		return nil
	})

	if retryErr != nil {
		return fmt.Errorf("Error: could not update status for Elasticsearch %v after %v retries: %v", cluster.Name, nretries, retryErr)
	}

	return nil
}

// performCertClusterRestart is required when certs need to be refreshed
// this is a very high priority action since the cluster may be fractured/unusable
// in the case where certs aren't all rolled out correctly or are expired
func (elasticsearchRequest *ElasticsearchRequest) performCertClusterRestart() error {
	esClient := elasticsearchRequest.esClient

	// make sure we have nodes that are scheduled for full cluster restart first
	certRedeployNodes := getScheduledCertRedeployNodes(elasticsearchRequest.cluster)
	clusterStatus := &elasticsearchRequest.cluster.Status

	// 1 -- precheck
	// no restarting conditions set
	if len(certRedeployNodes) > 0 {

		if containsClusterCondition(api.Restarting, v1.ConditionFalse, clusterStatus) &&
			containsClusterCondition(api.UpdatingSettings, v1.ConditionFalse, clusterStatus) {

			// We don't want to gate on cluster health -- we may be in a bad state
			// because pods aren't all available
			logrus.Infof("Beginning full cluster restart for cert redeploy on %v", elasticsearchRequest.cluster.Name)

			// set conditions here for next check
			updateUpdatingSettingsCondition(clusterStatus, v1.ConditionTrue)
		}

		// 2 -- prep for restart
		// condition updatingsettings true
		if containsClusterCondition(api.Restarting, v1.ConditionFalse, clusterStatus) &&
			containsClusterCondition(api.UpdatingSettings, v1.ConditionTrue, clusterStatus) {

			// disable shard allocation
			if ok, err := esClient.SetShardAllocation(api.ShardAllocationPrimaries); !ok {
				logrus.Warnf("Unable to set shard allocation to primaries: %v", err)
			}

			// flush nodes
			if ok, err := esClient.DoSynchronizedFlush(); !ok {
				logrus.Warnf("Unable to perform synchronized flush: %v", err)
			}

			updateRestartingCondition(clusterStatus, v1.ConditionTrue)
			updateUpdatingSettingsCondition(clusterStatus, v1.ConditionFalse)
		}

		// 3 -- restart
		// condition restarting true
		if containsClusterCondition(api.Restarting, v1.ConditionTrue, clusterStatus) &&
			containsClusterCondition(api.UpdatingSettings, v1.ConditionFalse, clusterStatus) {

			// call fullClusterRestart on each node that is scheduled for a full cluster restart
			for _, node := range certRedeployNodes {
				_, nodeStatus := getNodeStatus(node.name(), clusterStatus)
				node.fullClusterRestart(nodeStatus)
				addNodeState(node, nodeStatus)
				if err := elasticsearchRequest.setNodeStatus(node, nodeStatus, clusterStatus); err != nil {
					logrus.Warnf("unable to set node status. %s", err.Error())
				}
			}

			// check that all nodes have been restarted by seeing if they still have the need to cert restart
			if len(getScheduledCertRedeployNodes(elasticsearchRequest.cluster)) > 0 {
				return fmt.Errorf("Not all nodes were able to be restarted yet")
			}

			updateUpdatingSettingsCondition(clusterStatus, v1.ConditionTrue)
		}
	}

	// 4 -- post restart
	// condition restarting true and updatingsettings true
	if containsClusterCondition(api.Restarting, v1.ConditionTrue, clusterStatus) &&
		containsClusterCondition(api.UpdatingSettings, v1.ConditionTrue, clusterStatus) {

		// verify all nodes rejoined
		// check that we have no failed/notReady nodes
		podStatus := elasticsearchRequest.GetCurrentPodStateMap()
		if len(podStatus[api.ElasticsearchRoleData][api.PodStateTypeNotReady]) > 0 ||
			len(podStatus[api.ElasticsearchRoleMaster][api.PodStateTypeNotReady]) > 0 {

			logrus.Warnf("Waiting for all cluster nodes to rejoin after full cluster restart...")
			return fmt.Errorf("Waiting for all cluster nodes to rejoin after full cluster restart")
		}

		// reenable shard allocation
		if ok, err := esClient.SetShardAllocation(api.ShardAllocationAll); !ok {
			logrus.Warnf("Unable to enable shard allocation: %v", err)
			return err
		}

		updateUpdatingSettingsCondition(clusterStatus, v1.ConditionFalse)
	}

	// 5 -- recovery
	// wait for cluster to go green again
	if containsClusterCondition(api.Restarting, v1.ConditionTrue, clusterStatus) {
		status, err := esClient.GetClusterHealthStatus()
		if err != nil {
			return fmt.Errorf("Receieved error while trying to get cluster health for %q: %v", elasticsearchRequest.cluster.Name, err)
		}

		if !utils.Contains(desiredClusterStates, status) {
			return fmt.Errorf("Waiting for cluster to recover: %s / %v", status, desiredClusterStates)
		}

		logrus.Infof("Completed full cluster restart for cert redeploy on %v", elasticsearchRequest.cluster.Name)
		updateRestartingCondition(clusterStatus, v1.ConditionFalse)
	}

	return nil
}

func (elasticsearchRequest *ElasticsearchRequest) fullClusterRestart(scheduledUpgradeNodes []NodeTypeInterface) error {
	esClient := elasticsearchRequest.esClient
	clusterStatus := &elasticsearchRequest.cluster.Status

	// 1 -- precheck
	// no restarting conditions set
	if len(scheduledUpgradeNodes) > 0 {

		if containsClusterCondition(api.Restarting, v1.ConditionFalse, clusterStatus) &&
			containsClusterCondition(api.UpdatingSettings, v1.ConditionFalse, clusterStatus) {

			if status, _ := esClient.GetClusterHealthStatus(); !utils.Contains(desiredClusterStates, status) {
				return fmt.Errorf("Waiting for cluster %q to be recovered before restarting: %s / %v", elasticsearchRequest.cluster.Name, status, desiredClusterStates)
			}

			logrus.Infof("Beginning full cluster restart on %v", elasticsearchRequest.cluster.Name)

			// set conditions here for next check
			updateUpdatingSettingsCondition(clusterStatus, v1.ConditionTrue)
		}

		// 2 -- prep for restart
		// condition updatingsettings true
		if containsClusterCondition(api.Restarting, v1.ConditionFalse, clusterStatus) &&
			containsClusterCondition(api.UpdatingSettings, v1.ConditionTrue, clusterStatus) {

			// disable shard allocation
			if ok, err := esClient.SetShardAllocation(api.ShardAllocationPrimaries); !ok {
				return fmt.Errorf("Unable to set shard allocation to primaries: %v", err)
			}

			// flush nodes
			if ok, err := esClient.DoSynchronizedFlush(); !ok {
				logrus.Warnf("Unable to perform synchronized flush: %v", err)
			}

			updateRestartingCondition(clusterStatus, v1.ConditionTrue)
			updateUpdatingSettingsCondition(clusterStatus, v1.ConditionFalse)
		}

		// 3 -- restart
		// condition restarting true
		if containsClusterCondition(api.Restarting, v1.ConditionTrue, clusterStatus) &&
			containsClusterCondition(api.UpdatingSettings, v1.ConditionFalse, clusterStatus) {

			// call fullClusterRestart on each node that is scheduled for a full cluster restart
			for _, node := range scheduledUpgradeNodes {
				_, nodeStatus := getNodeStatus(node.name(), clusterStatus)

				node.progressNodeChanges(nodeStatus)

				addNodeState(node, nodeStatus)
				if err := elasticsearchRequest.setNodeStatus(node, nodeStatus, clusterStatus); err != nil {
					logrus.Warnf("unable to set node status. %s", err.Error())
				}
			}

			// check that all nodes have been restarted by seeing if they still have the need to cert restart
			if len(getScheduledUpgradeNodes(elasticsearchRequest.cluster)) > 0 {
				return fmt.Errorf("Not all nodes were able to be restarted yet")
			}

			updateUpdatingSettingsCondition(clusterStatus, v1.ConditionTrue)
		}
	}

	// 4 -- post restart
	// condition restarting true and updatingsettings true
	if containsClusterCondition(api.Restarting, v1.ConditionTrue, clusterStatus) &&
		containsClusterCondition(api.UpdatingSettings, v1.ConditionTrue, clusterStatus) {

		// verify all nodes rejoined
		// check that we have no failed/notReady nodes
		logrus.Infof("Waiting for all nodes to rejoin cluster %q in namespace %q", elasticsearchRequest.cluster.Name, elasticsearchRequest.cluster.Namespace)

		for _, node := range scheduledUpgradeNodes {
			if err, _ := node.waitForNodeRejoinCluster(); err != nil {
				return fmt.Errorf("Timed out waiting for %v to rejoin cluster %v", node.name(), elasticsearchRequest.cluster.Name)
			}
		}

		// reenable shard allocation
		if ok, err := esClient.SetShardAllocation(api.ShardAllocationAll); !ok {
			return fmt.Errorf("Unable to enable shard allocation: %v", err)
		}

		updateUpdatingSettingsCondition(clusterStatus, v1.ConditionFalse)
	}

	// 5 -- recovery
	// wait for cluster to go yellow/green again
	if containsClusterCondition(api.Restarting, v1.ConditionTrue, clusterStatus) {

		status, err := esClient.GetClusterHealthStatus()
		if err != nil {
			return fmt.Errorf("Receieved error while trying to get cluster health for %q: %v", elasticsearchRequest.cluster.Name, err)
		}

		if !utils.Contains(desiredClusterStates, status) {
			return fmt.Errorf("Waiting for cluster to recover: %s / %v", status, desiredClusterStates)
		}

		logrus.Infof("Completed full cluster restart on %v", elasticsearchRequest.cluster.Name)
		updateRestartingCondition(clusterStatus, v1.ConditionFalse)
	}

	return nil
}
