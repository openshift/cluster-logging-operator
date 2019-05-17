package k8shandler

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	api "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	apps "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type statefulSetNode struct {
	self apps.StatefulSet
	// prior hash for configmap content
	configmapHash string
	// prior hash for secret content
	secretHash string

	clusterName       string
	clusterSize       int32
	priorReplicaCount int32

	client client.Client
}

func (statefulSetNode *statefulSetNode) populateReference(nodeName string, node api.ElasticsearchNode, cluster *api.Elasticsearch, roleMap map[api.ElasticsearchNodeRole]bool, replicas int32, client client.Client) {

	labels := newLabels(cluster.Name, nodeName, roleMap)

	statefulSet := apps.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: apps.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      nodeName,
			Namespace: cluster.Namespace,
			Labels:    labels,
		},
	}

	partition := int32(0)

	statefulSet.Spec = apps.StatefulSetSpec{
		Replicas: &replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: newLabelSelector(cluster.Name, nodeName, roleMap),
		},
		Template: newPodTemplateSpec(nodeName, cluster.Name, cluster.Namespace, node, cluster.Spec.Spec, labels, roleMap, client),
		UpdateStrategy: apps.StatefulSetUpdateStrategy{
			Type: apps.RollingUpdateStatefulSetStrategyType,
			RollingUpdate: &apps.RollingUpdateStatefulSetStrategy{
				Partition: &partition,
			},
		},
	}
	statefulSet.Spec.Template.Spec.Containers[0].ReadinessProbe = nil

	addOwnerRefToObject(&statefulSet, getOwnerRef(cluster))

	statefulSetNode.self = statefulSet
	statefulSetNode.clusterName = cluster.Name

	statefulSetNode.client = client
}

func (current *statefulSetNode) updateReference(desired NodeTypeInterface) {
	current.self = desired.(*statefulSetNode).self
}

func (node *statefulSetNode) state() api.ElasticsearchNodeStatus {
	var rolloutForReload v1.ConditionStatus
	var rolloutForUpdate v1.ConditionStatus

	// see if we need to update the deployment object
	if node.isChanged() {
		rolloutForUpdate = v1.ConditionTrue
	}

	// check if the configmapHash changed
	/*newConfigmapHash := getConfigmapDataHash(node.clusterName, node.self.Namespace)
	if newConfigmapHash != node.configmapHash {
		rolloutForReload = v1.ConditionTrue
	}*/

	// check if the secretHash changed
	newSecretHash := getSecretDataHash(node.clusterName, node.self.Namespace, node.client)
	if newSecretHash != node.secretHash {
		rolloutForReload = v1.ConditionTrue
	}

	return api.ElasticsearchNodeStatus{
		StatefulSetName: node.self.Name,
		UpgradeStatus: api.ElasticsearchNodeUpgradeStatus{
			ScheduledForUpgrade:  rolloutForUpdate,
			ScheduledForRedeploy: rolloutForReload,
		},
	}
}

func (node *statefulSetNode) name() string {
	return node.self.Name
}

func (node *statefulSetNode) waitForNodeRejoinCluster() (error, bool) {
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

func (node *statefulSetNode) waitForNodeLeaveCluster() (error, bool) {
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

func (node *statefulSetNode) setPartition(partitions int32) error {
	nretries := -1
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		nretries++
		if getErr := node.client.Get(context.TODO(), types.NamespacedName{Name: node.self.Name, Namespace: node.self.Namespace}, &node.self); getErr != nil {
			logrus.Debugf("Could not get Elasticsearch node resource %v: %v", node.self.Name, getErr)
			return getErr
		}

		if *node.self.Spec.UpdateStrategy.RollingUpdate.Partition == partitions {
			return nil
		}

		node.self.Spec.UpdateStrategy.RollingUpdate.Partition = &partitions

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

func (node *statefulSetNode) partition() (int32, error) {

	desired := &apps.StatefulSet{}

	if err := node.client.Get(context.TODO(), types.NamespacedName{Name: node.self.Name, Namespace: node.self.Namespace}, desired); err != nil {
		logrus.Debugf("Could not get Elasticsearch node resource %v: %v", node.self.Name, err)
		return -1, err
	}

	return *desired.Spec.UpdateStrategy.RollingUpdate.Partition, nil
}

func (node *statefulSetNode) setReplicaCount(replicas int32) error {
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

func (node *statefulSetNode) replicaCount() (int32, error) {

	desired := &apps.StatefulSet{}

	if err := node.client.Get(context.TODO(), types.NamespacedName{Name: node.self.Name, Namespace: node.self.Namespace}, desired); err != nil {
		logrus.Debugf("Could not get Elasticsearch node resource %v: %v", node.self.Name, err)
		return -1, err
	}

	return desired.Status.Replicas, nil
}

func (node *statefulSetNode) restart(upgradeStatus *api.ElasticsearchNodeStatus) {

	if upgradeStatus.UpgradeStatus.UnderUpgrade != v1.ConditionTrue {
		if status, _ := GetClusterHealth(node.clusterName, node.self.Namespace, node.client); status != "green" {
			logrus.Infof("Waiting for cluster to be fully recovered before restarting %v: %v / green", node.name(), status)
			return
		}

		size, err := GetClusterNodeCount(node.clusterName, node.self.Namespace, node.client)
		if err != nil {
			logrus.Warnf("Unable to get cluster size prior to restart for %v", node.name())
			return
		}
		node.clusterSize = size

		replicas, err := node.replicaCount()
		if err != nil {
			logrus.Warnf("Unable to get number of replicas prior to restart for %v", node.name())
			return
		}

		node.setPartition(replicas)
		upgradeStatus.UpgradeStatus.UnderUpgrade = v1.ConditionTrue
	}

	if upgradeStatus.UpgradeStatus.UpgradePhase == "" ||
		upgradeStatus.UpgradeStatus.UpgradePhase == api.ControllerUpdated {

		// nothing to do here -- just maintaing a framework structure

		upgradeStatus.UpgradeStatus.UpgradePhase = api.NodeRestarting
	}

	if upgradeStatus.UpgradeStatus.UpgradePhase == api.NodeRestarting {

		ordinal, err := node.partition()
		if err != nil {
			logrus.Infof("Unable to get node ordinal value: %v", err)
			return
		}

		for index := ordinal; index > 0; index-- {
			// get podName based on ordinal index and node.name()
			podName := fmt.Sprintf("%v-%v", node.name(), index-1)

			// make sure we have all nodes in the cluster first -- always
			if err, _ := node.waitForNodeRejoinCluster(); err != nil {
				logrus.Infof("Timed out waiting for %v pods to rejoin cluster", node.name())
				return
			}

			// delete the pod
			if err := DeletePod(podName, node.self.Namespace, node.client); err != nil {
				logrus.Infof("Unable to delete pod %v for restart: %v", podName, err)
				return
			}

			// wait for node to leave cluster
			if err, _ := node.waitForNodeLeaveCluster(); err != nil {
				logrus.Infof("Timed out waiting for %v to leave the cluster", podName)
				return
			}

			// used for tracking in case of timeout
			node.setPartition(index - 1)
		}

		if err, _ := node.waitForNodeRejoinCluster(); err != nil {
			logrus.Infof("Timed out waiting for %v pods to rejoin cluster", node.name())
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

func (node *statefulSetNode) create() error {

	if node.self.ObjectMeta.ResourceVersion == "" {
		err := node.client.Create(context.TODO(), &node.self)
		if err != nil {
			if !errors.IsAlreadyExists(err) {
				return fmt.Errorf("Could not create node resource: %v", err)
			} else {
				node.scale()
				return nil
			}
		}

		// update the hashmaps
		node.configmapHash = getConfigmapDataHash(node.clusterName, node.self.Namespace, node.client)
		node.secretHash = getSecretDataHash(node.clusterName, node.self.Namespace, node.client)
	} else {
		node.scale()
	}

	return nil
}

func (node *statefulSetNode) executeUpdate() error {
	// see if we need to update the deployment object and verify we have latest to update
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// isChanged() will get the latest revision from the apiserver
		// and return false if there is nothing to change and will update the node object if required
		if node.isChanged() {
			if updateErr := node.client.Update(context.TODO(), &node.self); updateErr != nil {
				logrus.Debugf("Failed to update node resource %v: %v", node.self.Name, updateErr)
				return updateErr
			}
		}
		return nil
	})
}

func (node *statefulSetNode) update(upgradeStatus *api.ElasticsearchNodeStatus) error {
	if upgradeStatus.UpgradeStatus.UnderUpgrade != v1.ConditionTrue {
		if status, _ := GetClusterHealth(node.clusterName, node.self.Namespace, node.client); status != "green" {
			logrus.Infof("Waiting for cluster to be fully recovered before restarting %v: %v / green", node.name(), status)
			return fmt.Errorf("Waiting for cluster to be fully recovered before restarting %v: %v / green", node.name(), status)
		}

		size, err := GetClusterNodeCount(node.clusterName, node.self.Namespace, node.client)
		if err != nil {
			logrus.Warnf("Unable to get cluster size prior to restart for %v", node.name())
		}
		node.clusterSize = size

		replicas, err := node.replicaCount()
		if err != nil {
			logrus.Warnf("Unable to get number of replicas prior to restart for %v", node.name())
			return fmt.Errorf("Unable to get number of replicas prior to restart for %v", node.name())
		}

		node.setPartition(replicas)
		upgradeStatus.UpgradeStatus.UnderUpgrade = v1.ConditionTrue
	}

	if upgradeStatus.UpgradeStatus.UpgradePhase == "" ||
		upgradeStatus.UpgradeStatus.UpgradePhase == api.ControllerUpdated {

		if err := node.executeUpdate(); err != nil {
			return err
		}

		upgradeStatus.UpgradeStatus.UpgradePhase = api.NodeRestarting
	}

	if upgradeStatus.UpgradeStatus.UpgradePhase == api.NodeRestarting {

		ordinal, err := node.partition()
		if err != nil {
			logrus.Infof("Unable to get node ordinal value: %v", err)
			return err
		}

		// start partition at replicas and incrementally update it to 0
		// making sure nodes rejoin between each one
		for index := ordinal; index > 0; index-- {

			// make sure we have all nodes in the cluster first -- always
			if err, _ := node.waitForNodeRejoinCluster(); err != nil {
				logrus.Infof("Timed out waiting for %v to rejoin cluster", node.name())
				return fmt.Errorf("Timed out waiting for %v to rejoin cluster", node.name())
			}

			// update partition to cause next pod to be updated
			node.setPartition(index - 1)

			// wait for the node to leave the cluster
			if err, _ := node.waitForNodeLeaveCluster(); err != nil {
				logrus.Infof("Timed out waiting for %v to leave the cluster", node.name())
				return fmt.Errorf("Timed out waiting for %v to leave the cluster", node.name())
			}
		}

		// this is here again because we need to make sure all nodes have rejoined
		// before we move on and say we're done
		if err, _ := node.waitForNodeRejoinCluster(); err != nil {
			logrus.Infof("Timed out waiting for %v to rejoin cluster", node.name())
			return fmt.Errorf("Timed out waiting for %v to rejoin cluster", node.name())
		}

		node.refreshHashes()

		upgradeStatus.UpgradeStatus.UpgradePhase = api.RecoveringData
	}

	if upgradeStatus.UpgradeStatus.UpgradePhase == api.RecoveringData {

		upgradeStatus.UpgradeStatus.UpgradePhase = api.ControllerUpdated
		upgradeStatus.UpgradeStatus.UnderUpgrade = ""
	}

	return nil
}

func (node *statefulSetNode) refreshHashes() {
	newConfigmapHash := getConfigmapDataHash(node.clusterName, node.self.Namespace, node.client)
	if newConfigmapHash != node.configmapHash {
		node.configmapHash = newConfigmapHash
	}

	newSecretHash := getSecretDataHash(node.clusterName, node.self.Namespace, node.client)
	if newSecretHash != node.secretHash {
		node.secretHash = newSecretHash
	}
}

func (node *statefulSetNode) scale() {

	desired := node.self.DeepCopy()
	err := node.client.Get(context.TODO(), types.NamespacedName{Name: node.self.Name, Namespace: node.self.Namespace}, &node.self)
	// error check that it exists, etc
	if err != nil {
		// if it doesn't exist, return true
		return
	}

	if *desired.Spec.Replicas != *node.self.Spec.Replicas {
		node.self.Spec.Replicas = desired.Spec.Replicas
		logrus.Infof("Resource '%s' has different container replicas than desired", node.self.Name)

		node.setReplicaCount(*node.self.Spec.Replicas)
	}
}

func (node *statefulSetNode) isChanged() bool {

	changed := false

	desired := node.self.DeepCopy()
	// we want to blank this out before a get to ensure we get the correct information back (possible sdk issue with maps?)
	node.self.Spec = apps.StatefulSetSpec{}

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

	// Only Image and Resources (CPU & memory) differences trigger rolling restart
	for index := 0; index < len(node.self.Spec.Template.Spec.Containers); index++ {
		nodeContainer := node.self.Spec.Template.Spec.Containers[index]
		desiredContainer := desired.Spec.Template.Spec.Containers[index]

		if nodeContainer.Resources.Requests == nil {
			nodeContainer.Resources.Requests = v1.ResourceList{}
		}

		if nodeContainer.Resources.Limits == nil {
			nodeContainer.Resources.Limits = v1.ResourceList{}
		}

		// check that both exist

		if nodeContainer.Image != desiredContainer.Image {
			logrus.Debugf("Resource '%s' has different container image than desired", node.self.Name)
			nodeContainer.Image = desiredContainer.Image
			changed = true
		}

		if desiredContainer.Resources.Limits.Cpu().Cmp(*nodeContainer.Resources.Limits.Cpu()) != 0 {
			logrus.Debugf("Resource '%s' has different CPU limit than desired", node.self.Name)
			nodeContainer.Resources.Limits[v1.ResourceCPU] = *desiredContainer.Resources.Limits.Cpu()
			changed = true
		}
		// Check memory limits
		if desiredContainer.Resources.Limits.Memory().Cmp(*nodeContainer.Resources.Limits.Memory()) != 0 {
			logrus.Debugf("Resource '%s' has different Memory limit than desired", node.self.Name)
			nodeContainer.Resources.Limits[v1.ResourceMemory] = *desiredContainer.Resources.Limits.Memory()
			changed = true
		}
		// Check CPU requests
		if desiredContainer.Resources.Requests.Cpu().Cmp(*nodeContainer.Resources.Requests.Cpu()) != 0 {
			logrus.Debugf("Resource '%s' has different CPU Request than desired", node.self.Name)
			nodeContainer.Resources.Requests[v1.ResourceCPU] = *desiredContainer.Resources.Requests.Cpu()
			changed = true
		}
		// Check memory requests
		if desiredContainer.Resources.Requests.Memory().Cmp(*nodeContainer.Resources.Requests.Memory()) != 0 {
			logrus.Debugf("Resource '%s' has different Memory Request than desired", node.self.Name)
			nodeContainer.Resources.Requests[v1.ResourceMemory] = *desiredContainer.Resources.Requests.Memory()
			changed = true
		}

		node.self.Spec.Template.Spec.Containers[index] = nodeContainer
	}
	return changed
}

func (node *statefulSetNode) progressUnshedulableNode(upgradeStatus *api.ElasticsearchNodeStatus) error {
	if node.isChanged() {
		if err := node.executeUpdate(); err != nil {
			return err
		}

		partition, err := node.partition()
		if err != nil {
			return err
		}

		podName := fmt.Sprintf("%v-%v", node.name(), partition)

		logrus.Debugf("Updated statefulset %s, manually applying changes on pod: %s", node.name(), podName)

		if err := DeletePod(podName, node.self.Namespace, node.client); err != nil {
			return err
		}

	}
	return nil
}
