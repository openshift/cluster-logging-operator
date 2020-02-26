package k8shandler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	api "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	"github.com/openshift/elasticsearch-operator/pkg/logger"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type deploymentNode struct {
	self apps.Deployment
	// prior hash for configmap content
	configmapHash string
	// prior hash for secret content
	secretHash string

	clusterName string

	currentRevision string

	clusterSize int32

	client client.Client
}

func (deploymentNode *deploymentNode) populateReference(nodeName string, node api.ElasticsearchNode, cluster *api.Elasticsearch, roleMap map[api.ElasticsearchNodeRole]bool, replicas int32, client client.Client) {

	labels := newLabels(cluster.Name, nodeName, roleMap)

	deployment := apps.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: apps.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      nodeName,
			Namespace: cluster.Namespace,
			Labels:    labels,
		},
	}

	progressDeadlineSeconds := int32(1800)

	deployment.Spec = apps.DeploymentSpec{
		Replicas: &replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: newLabelSelector(cluster.Name, nodeName, roleMap),
		},
		Strategy: apps.DeploymentStrategy{
			Type: "Recreate",
		},
		ProgressDeadlineSeconds: &progressDeadlineSeconds,
		Paused:                  false,
		Template:                newPodTemplateSpec(nodeName, cluster.Name, cluster.Namespace, node, cluster.Spec.Spec, labels, roleMap, client),
	}

	addOwnerRefToObject(&deployment, getOwnerRef(cluster))

	deploymentNode.self = deployment
	deploymentNode.clusterName = cluster.Name

	deploymentNode.client = client
}

func (current *deploymentNode) updateReference(node NodeTypeInterface) {
	current.self = node.(*deploymentNode).self
}

func (node *deploymentNode) name() string {
	return node.self.Name
}

func (node *deploymentNode) state() api.ElasticsearchNodeStatus {

	//var rolloutForReload v1.ConditionStatus
	var rolloutForUpdate v1.ConditionStatus
	var rolloutForCertReload v1.ConditionStatus

	// see if we need to update the deployment object
	if node.isChanged() {
		rolloutForUpdate = v1.ConditionTrue
	}

	// check if the configmapHash changed
	/*newConfigmapHash := getConfigmapDataHash(node.clusterName, node.self.Namespace)
	if newConfigmapHash != node.configmapHash {
		rolloutForReload = v1.ConditionTrue
	}*/

	// check for a case where our hash is missing -- operator restarted?
	newSecretHash := getSecretDataHash(node.clusterName, node.self.Namespace, node.client)
	if node.secretHash == "" {
		// if we were already scheduled to restart, don't worry? -- just grab
		// the current hash -- we should have already had our upgradeStatus set if
		// we required a restart...
		node.secretHash = newSecretHash
	} else {
		// check if the secretHash changed
		if newSecretHash != node.secretHash {
			rolloutForCertReload = v1.ConditionTrue
		}
	}

	return api.ElasticsearchNodeStatus{
		DeploymentName: node.self.Name,
		UpgradeStatus: api.ElasticsearchNodeUpgradeStatus{
			ScheduledForUpgrade:      rolloutForUpdate,
			ScheduledForCertRedeploy: rolloutForCertReload,
		},
	}
}

func (node *deploymentNode) delete() {
	node.client.Delete(context.TODO(), &node.self)
}

func (node *deploymentNode) create() error {

	if node.self.ObjectMeta.ResourceVersion == "" {
		err := node.client.Create(context.TODO(), &node.self)
		if err != nil {
			if !errors.IsAlreadyExists(err) {
				return fmt.Errorf("Could not create node resource: %v", err)
			} else {
				return node.pause()
			}
		}

		// created unpaused, pause after deployment...
		// wait until we have a revision annotation...
		if err = node.waitForInitialRollout(); err != nil {
			return err
		}

		// update the hashmaps
		node.configmapHash = getConfigmapDataHash(node.clusterName, node.self.Namespace, node.client)
		node.secretHash = getSecretDataHash(node.clusterName, node.self.Namespace, node.client)
	}

	return node.pause()
}

func (node *deploymentNode) waitForInitialRollout() error {
	err := wait.Poll(time.Second*1, time.Second*30, func() (done bool, err error) {
		if getErr := node.client.Get(context.TODO(), types.NamespacedName{Name: node.self.Name, Namespace: node.self.Namespace}, &node.self); getErr != nil {
			logrus.Debugf("Could not get Elasticsearch node resource %v: %v", node.self.Name, getErr)
			return false, getErr
		}

		_, ok := node.self.ObjectMeta.Annotations["deployment.kubernetes.io/revision"]
		if ok {
			return true, nil
		}

		return false, nil
	})
	return err
}

func (node *deploymentNode) nodeRevision() string {
	val, ok := node.self.ObjectMeta.Annotations["deployment.kubernetes.io/revision"]

	if ok {
		return val
	}

	return ""
}

func (node *deploymentNode) waitForNodeRollout(currentRevision string) error {
	err := wait.Poll(time.Second*1, time.Second*30, func() (done bool, err error) {
		if getErr := node.client.Get(context.TODO(), types.NamespacedName{Name: node.self.Name, Namespace: node.self.Namespace}, &node.self); getErr != nil {
			logrus.Debugf("Could not get Elasticsearch node resource %v: %v", node.self.Name, getErr)
			return false, getErr
		}

		revision := node.nodeRevision()

		return (revision != currentRevision), nil
	})
	return err
}

func (node *deploymentNode) pause() error {
	return node.setPaused(true)
}

func (node *deploymentNode) unpause() error {
	return node.setPaused(false)
}

func (node *deploymentNode) setPaused(paused bool) error {

	// we use pauseNode so that we don't revert any new changes that should be made and
	// noticed in state()
	pauseNode := node.self.DeepCopy()

	nretries := -1
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		nretries++
		if getErr := node.client.Get(context.TODO(), types.NamespacedName{Name: pauseNode.Name, Namespace: pauseNode.Namespace}, pauseNode); getErr != nil {
			logrus.Debugf("Could not get Elasticsearch node resource %v: %v", pauseNode.Name, getErr)
			return getErr
		}

		if pauseNode.Spec.Paused == paused {
			return nil
		}

		pauseNode.Spec.Paused = paused

		if updateErr := node.client.Update(context.TODO(), pauseNode); updateErr != nil {
			logrus.Debugf("Failed to update node resource %v: %v", pauseNode.Name, updateErr)
			return updateErr
		}
		return nil
	})
	if retryErr != nil {
		return fmt.Errorf("Error: could not update Elasticsearch node %v after %v retries: %v", node.self.Name, nretries, retryErr)
	}

	node.self.Spec.Paused = pauseNode.Spec.Paused

	return nil
}

func (node *deploymentNode) setReplicaCount(replicas int32) error {
	nretries := -1
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		nretries++
		if getErr := node.client.Get(context.TODO(), types.NamespacedName{Name: node.self.Name, Namespace: node.self.Namespace}, &node.self); getErr != nil {
			logrus.Debugf("Could not get Elasticsearch node resource %v: %v", node.self.Name, getErr)
			return getErr
		}

		if *node.self.Spec.Replicas == replicas {
			return nil
		}

		node.self.Spec.Replicas = &replicas

		if updateErr := node.client.Update(context.TODO(), &node.self); updateErr != nil {
			logrus.Debugf("Failed to update node resource %v: %v", node.self.Name, updateErr)
			return updateErr
		}
		return nil
	})
	if retryErr != nil {
		return fmt.Errorf("Error: could not update Elasticsearch node %v after %v retries: %v", node.self.Name, nretries, retryErr)
	}

	return nil
}

func (node *deploymentNode) replicaCount() (error, int32) {
	nodeCopy := &apps.Deployment{}

	if err := node.client.Get(context.TODO(), types.NamespacedName{Name: node.self.Name, Namespace: node.self.Namespace}, nodeCopy); err != nil {
		logrus.Debugf("Could not get Elasticsearch node resource %v: %v", node.self.Name, err)
		return err, -1
	}

	return nil, nodeCopy.Status.Replicas
}

func (node *deploymentNode) waitForNodeRejoinCluster() (error, bool) {
	err := wait.Poll(time.Second*1, time.Second*60, func() (done bool, err error) {
		clusterSize, getErr := GetClusterNodeCount(node.clusterName, node.self.Namespace, node.client)
		if err != nil {
			logrus.Warnf("Unable to get cluster size waiting for %v to rejoin cluster", node.name())
			return false, getErr
		}

		return (node.clusterSize <= clusterSize), nil
	})

	return err, (err == nil)
}

func (node *deploymentNode) waitForNodeLeaveCluster() (error, bool) {
	err := wait.Poll(time.Second*1, time.Second*60, func() (done bool, err error) {
		clusterSize, getErr := GetClusterNodeCount(node.clusterName, node.self.Namespace, node.client)
		if err != nil {
			logrus.Warnf("Unable to get cluster size waiting for %v to leave cluster", node.name())
			return false, getErr
		}

		return (node.clusterSize > clusterSize), nil
	})

	return err, (err == nil)
}

func (node *deploymentNode) isMissing() bool {
	getNode := &apps.Deployment{}
	if getErr := node.client.Get(context.TODO(), types.NamespacedName{Name: node.name(), Namespace: node.self.Namespace}, getNode); getErr != nil {
		if errors.IsNotFound(getErr) {
			return true
		}
	}

	return false
}

func (node *deploymentNode) rollingRestart(upgradeStatus *api.ElasticsearchNodeStatus) {

	if upgradeStatus.UpgradeStatus.UnderUpgrade != v1.ConditionTrue {
		if status, _ := GetClusterHealthStatus(node.clusterName, node.self.Namespace, node.client); status != "green" {
			logrus.Infof("Waiting for cluster to be fully recovered before restarting %v: %v / green", node.name(), status)
			return
		}

		size, err := GetClusterNodeCount(node.clusterName, node.self.Namespace, node.client)
		if err != nil {
			logrus.Warnf("Unable to get cluster size prior to restart for %v", node.name())
			return
		}
		node.clusterSize = size
		upgradeStatus.UpgradeStatus.UnderUpgrade = v1.ConditionTrue
	}

	if upgradeStatus.UpgradeStatus.UpgradePhase == "" ||
		upgradeStatus.UpgradeStatus.UpgradePhase == api.ControllerUpdated {

		err, replicas := node.replicaCount()
		if err != nil {
			logrus.Warnf("Unable to get replica count for %v", node.name())
		}

		if replicas > 0 {

			// disable shard allocation
			if ok, err := SetShardAllocation(node.clusterName, node.self.Namespace, api.ShardAllocationNone, node.client); !ok {
				logrus.Warnf("Unable to disable shard allocation: %v", err)
				return
			}

			if ok, err := DoSynchronizedFlush(node.clusterName, node.self.Namespace, node.client); !ok {
				logrus.Warnf("Unable to perform synchronized flush: %v", err)
			}

			// check for available replicas empty
			// node.self.Status.Replicas
			// if we aren't at 0, then we need to scale down to 0
			if err = node.setReplicaCount(0); err != nil {
				logrus.Warnf("Unable to scale down %v", node.name())
				return
			}

			if err, _ = node.waitForNodeLeaveCluster(); err != nil {
				logrus.Infof("Timed out waiting for %v to leave the cluster", node.name())
				return
			}
		}

		upgradeStatus.UpgradeStatus.UpgradePhase = api.NodeRestarting
	}

	if upgradeStatus.UpgradeStatus.UpgradePhase == api.NodeRestarting {

		// if the node doesn't exist -- create it
		if node.isMissing() {
			node.create()
		}

		if err := node.setReplicaCount(1); err != nil {
			logrus.Warnf("Unable to scale up %v", node.name())
			return
		}

		if err, _ := node.waitForNodeRejoinCluster(); err != nil {
			logrus.Infof("Timed out waiting for %v to rejoin cluster", node.name())
			return
		}

		node.refreshHashes()

		// reenable shard allocation
		if ok, err := SetShardAllocation(node.clusterName, node.self.Namespace, api.ShardAllocationAll, node.client); !ok {
			logrus.Warnf("Unable to enable shard allocation: %v", err)
			return
		}

		upgradeStatus.UpgradeStatus.UpgradePhase = api.RecoveringData
	}

	if upgradeStatus.UpgradeStatus.UpgradePhase == api.RecoveringData {

		if status, _ := GetClusterHealthStatus(node.clusterName, node.self.Namespace, node.client); status != "green" {
			logrus.Infof("Waiting for cluster to complete recovery: %v / green", status)
			return
		}

		upgradeStatus.UpgradeStatus.UpgradePhase = api.ControllerUpdated
		upgradeStatus.UpgradeStatus.UnderUpgrade = ""
	}
}

func (node *deploymentNode) fullClusterRestart(upgradeStatus *api.ElasticsearchNodeStatus) {

	if upgradeStatus.UpgradeStatus.UnderUpgrade != v1.ConditionTrue {
		upgradeStatus.UpgradeStatus.UnderUpgrade = v1.ConditionTrue
		size, err := GetClusterNodeCount(node.clusterName, node.self.Namespace, node.client)
		if err != nil {
			logrus.Warnf("Unable to get cluster size prior to restart for %v", node.name())
			return
		}
		node.clusterSize = size
	}

	if upgradeStatus.UpgradeStatus.UpgradePhase == "" ||
		upgradeStatus.UpgradeStatus.UpgradePhase == api.ControllerUpdated {

		err, replicas := node.replicaCount()
		if err != nil {
			logrus.Warnf("Unable to get replica count for %v", node.name())
		}

		if replicas > 0 {
			// check for available replicas empty
			// node.self.Status.Replicas
			// if we aren't at 0, then we need to scale down to 0
			if err = node.setReplicaCount(0); err != nil {
				logrus.Warnf("Unable to scale down %v", node.name())
				return
			}

			if err, _ = node.waitForNodeLeaveCluster(); err != nil {
				logrus.Infof("Timed out waiting for %v to leave the cluster", node.name())
				return
			}
		}

		upgradeStatus.UpgradeStatus.UpgradePhase = api.NodeRestarting
	}

	if upgradeStatus.UpgradeStatus.UpgradePhase == api.NodeRestarting {

		if err := node.setReplicaCount(1); err != nil {
			logrus.Warnf("Unable to scale up %v", node.name())
			return
		}

		node.refreshHashes()

		upgradeStatus.UpgradeStatus.UpgradePhase = api.RecoveringData
	}

	if upgradeStatus.UpgradeStatus.UpgradePhase == api.RecoveringData {

		upgradeStatus.UpgradeStatus.UpgradePhase = api.ControllerUpdated
		upgradeStatus.UpgradeStatus.UnderUpgrade = ""
	}
}

func (node *deploymentNode) update(upgradeStatus *api.ElasticsearchNodeStatus) error {

	// set our state to being under upgrade
	if upgradeStatus.UpgradeStatus.UnderUpgrade != v1.ConditionTrue {
		if status, _ := GetClusterHealthStatus(node.clusterName, node.self.Namespace, node.client); status != "green" {
			logrus.Infof("Waiting for cluster to be fully recovered before upgrading %v: %v / green", node.name(), status)
			return fmt.Errorf("Cluster not in green state before beginning upgrade: %v", status)
		}

		size, err := GetClusterNodeCount(node.clusterName, node.self.Namespace, node.client)
		if err != nil {
			logrus.Warnf("Unable to get cluster size prior to update for %v", node.name())
		}
		node.clusterSize = size
		upgradeStatus.UpgradeStatus.UnderUpgrade = v1.ConditionTrue
	}

	// use UpgradePhase to gate what we work on, update phase when we complete a task
	if upgradeStatus.UpgradeStatus.UpgradePhase == "" ||
		upgradeStatus.UpgradeStatus.UpgradePhase == api.ControllerUpdated {

		// disable shard allocation
		if ok, err := SetShardAllocation(node.clusterName, node.self.Namespace, api.ShardAllocationNone, node.client); !ok {
			logrus.Warnf("Unable to disable shard allocation: %v", err)
			return err
		}

		if ok, err := DoSynchronizedFlush(node.clusterName, node.self.Namespace, node.client); !ok {
			logrus.Warnf("Unable to perform synchronized flush: %v", err)
		}

		if err := node.executeUpdate(); err != nil {
			return err
		}

		upgradeStatus.UpgradeStatus.UpgradePhase = api.NodeRestarting
		node.currentRevision = node.nodeRevision()
	}

	if upgradeStatus.UpgradeStatus.UpgradePhase == api.NodeRestarting {

		// do a unpause, wait, and pause again
		node.unpause()

		// wait for rollout
		if err := node.waitForNodeRollout(node.currentRevision); err != nil {
			logrus.Infof("Timed out waiting for node %v to rollout", node.name())
			return err
		}

		// pause again
		node.pause()

		// once we've restarted this is taken care of
		node.refreshHashes()

		// wait for node to rejoin cluster
		if err, _ := node.waitForNodeRejoinCluster(); err != nil {
			logrus.Infof("Timed out waiting for %v to rejoin cluster", node.name())
			return fmt.Errorf("Node %v has not rejoined cluster %v yet", node.name(), node.clusterName)
		}

		// reenable shard allocation
		if ok, err := SetShardAllocation(node.clusterName, node.self.Namespace, api.ShardAllocationAll, node.client); !ok {
			logrus.Warnf("Unable to enable shard allocation: %v", err)
			return err
		}

		upgradeStatus.UpgradeStatus.UpgradePhase = api.RecoveringData
	}

	if upgradeStatus.UpgradeStatus.UpgradePhase == api.RecoveringData {

		if status, err := GetClusterHealthStatus(node.clusterName, node.self.Namespace, node.client); status != "green" {
			logrus.Infof("Waiting for cluster to complete recovery: %v / green", status)
			return err
		}

		upgradeStatus.UpgradeStatus.UpgradePhase = api.ControllerUpdated
		upgradeStatus.UpgradeStatus.UnderUpgrade = ""
	}

	return nil
}

func (node *deploymentNode) executeUpdate() error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// isChanged() will get the latest revision from the apiserver
		// and return false if there is nothing to change and will update the node object if required
		if node.isChanged() {
			if logger.IsDebugEnabled() {
				pretty, err := json.MarshalIndent(node.self, "", "  ")
				if err != nil {
					logger.Debugf("Error marshaling node for debug log: %v", err)
				}
				logger.Debugf("Attempting to update node deployment: %+v", string(pretty))
			}
			if updateErr := node.client.Update(context.TODO(), &node.self); updateErr != nil {
				logger.Debugf("Failed to update node resource %v: %v", node.self.Name, updateErr)
				return updateErr
			}
		}
		return nil
	})
}

func (node *deploymentNode) progressUnshedulableNode(upgradeStatus *api.ElasticsearchNodeStatus) error {
	if node.isChanged() {
		logrus.Infof("Requested to update node '%s', which is unschedulable. Skipping rolling restart scenario and performing redeploy now", upgradeStatus.DeploymentName)

		if err := node.executeUpdate(); err != nil {
			return err
		}

		if err := node.unpause(); err != nil {
			return err
		}
		// if unpause is succesfull, always try to pause
		defer node.pause()

		logrus.Debugf("Waiting for node '%s' to rollout...", node.name())

		if err := node.waitForNodeRollout(node.currentRevision); err != nil {
			logrus.Infof("Timed out waiting for node %v to rollout", node.name())
			return err
		}

		upgradeStatus.UpgradeStatus.UpgradePhase = api.NodeRestarting
		node.currentRevision = node.nodeRevision()
	}
	return nil
}

func (node *deploymentNode) refreshHashes() {
	newConfigmapHash := getConfigmapDataHash(node.clusterName, node.self.Namespace, node.client)
	if newConfigmapHash != node.configmapHash {
		node.configmapHash = newConfigmapHash
	}

	newSecretHash := getSecretDataHash(node.clusterName, node.self.Namespace, node.client)
	if newSecretHash != node.secretHash {
		node.secretHash = newSecretHash
	}
}

func (node *deploymentNode) isChanged() bool {

	changed := false

	desired := node.self.DeepCopy()
	// we want to blank this out before a get to ensure we get the correct information back (possible sdk issue with maps?)
	node.self.Spec = apps.DeploymentSpec{}

	err := node.client.Get(context.TODO(), types.NamespacedName{Name: node.self.Name, Namespace: node.self.Namespace}, &node.self)
	// error check that it exists, etc
	if err != nil {
		// if it doesn't exist, return true
		return false
	}

	// check the pod's nodeselector
	if !areSelectorsSame(node.self.Spec.Template.Spec.NodeSelector, desired.Spec.Template.Spec.NodeSelector) {
		logrus.Debugf("Resource '%s' has different nodeSelector than desired", node.self.Name)
		node.self.Spec.Template.Spec.NodeSelector = desired.Spec.Template.Spec.NodeSelector
		changed = true
	}

	// check the pod's tolerations
	if !areTolerationsSame(node.self.Spec.Template.Spec.Tolerations, desired.Spec.Template.Spec.Tolerations) {
		logrus.Debugf("Resource '%s' has different tolerations than desired", node.self.Name)
		node.self.Spec.Template.Spec.Tolerations = desired.Spec.Template.Spec.Tolerations
		changed = true
	}

	// Only Image and Resources (CPU & memory) differences trigger rolling restart
	for index := 0; index < len(node.self.Spec.Template.Spec.Containers); index++ {
		nodeContainer := node.self.Spec.Template.Spec.Containers[index]
		desiredContainer := desired.Spec.Template.Spec.Containers[index]

		if nodeContainer.Image != desiredContainer.Image {
			logrus.Debugf("Resource '%s' has different container image than desired", node.self.Name)
			nodeContainer.Image = desiredContainer.Image
			changed = true
		}

		var updatedContainer v1.Container
		var resourceUpdated bool
		if updatedContainer, resourceUpdated = updateResources(node, nodeContainer, desiredContainer); resourceUpdated {
			changed = true
		}

		node.self.Spec.Template.Spec.Containers[index] = updatedContainer
	}

	return changed
}

//updateResources for the node; return true if an actual change is made
func updateResources(node *deploymentNode, nodeContainer, desiredContainer v1.Container) (v1.Container, bool) {
	changed := false
	if nodeContainer.Resources.Requests == nil {
		nodeContainer.Resources.Requests = v1.ResourceList{}
	}

	if nodeContainer.Resources.Limits == nil {
		nodeContainer.Resources.Limits = v1.ResourceList{}
	}

	// Check CPU limits
	if desiredContainer.Resources.Limits.Cpu().Cmp(*nodeContainer.Resources.Limits.Cpu()) != 0 {
		logrus.Debugf("Resource '%s' has different CPU (%+v) limit than desired (%+v)", node.self.Name, *nodeContainer.Resources.Limits.Cpu(), desiredContainer.Resources.Limits.Cpu())
		nodeContainer.Resources.Limits[v1.ResourceCPU] = *desiredContainer.Resources.Limits.Cpu()
		if nodeContainer.Resources.Limits.Cpu().IsZero() {
			delete(nodeContainer.Resources.Limits, v1.ResourceCPU)
		}
		changed = true
	}
	// Check memory limits
	if desiredContainer.Resources.Limits.Memory().Cmp(*nodeContainer.Resources.Limits.Memory()) != 0 {
		logrus.Debugf("Resource '%s' has different Memory limit than desired", node.self.Name)
		nodeContainer.Resources.Limits[v1.ResourceMemory] = *desiredContainer.Resources.Limits.Memory()
		if nodeContainer.Resources.Limits.Memory().IsZero() {
			delete(nodeContainer.Resources.Limits, v1.ResourceMemory)
		}
		changed = true
	}
	// Check CPU requests
	if desiredContainer.Resources.Requests.Cpu().Cmp(*nodeContainer.Resources.Requests.Cpu()) != 0 {
		logrus.Debugf("Resource '%s' has different CPU Request than desired", node.self.Name)
		nodeContainer.Resources.Requests[v1.ResourceCPU] = *desiredContainer.Resources.Requests.Cpu()
		if nodeContainer.Resources.Requests.Cpu().IsZero() {
			delete(nodeContainer.Resources.Requests, v1.ResourceCPU)
		}
		changed = true
	}
	// Check memory requests
	if desiredContainer.Resources.Requests.Memory().Cmp(*nodeContainer.Resources.Requests.Memory()) != 0 {
		logrus.Debugf("Resource '%s' has different Memory Request than desired", node.self.Name)
		nodeContainer.Resources.Requests[v1.ResourceMemory] = *desiredContainer.Resources.Requests.Memory()
		if nodeContainer.Resources.Requests.Memory().IsZero() {
			delete(nodeContainer.Resources.Requests, v1.ResourceMemory)
		}
		changed = true
	}

	return nodeContainer, changed
}
