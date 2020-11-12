package k8shandler

import (
	"strings"

	"github.com/openshift/elasticsearch-operator/pkg/utils"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	loggingv1 "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
)

// recoverOrphanedCluster is used to look for an existing cluster
// that matches the ElasticsearchRequest name but the UUID may be different
// NOTE: we only try to recover if our nodes do not have a UUID defined
func (er *ElasticsearchRequest) recoverOrphanedCluster() error {

	nodesToMatch := make(map[int]loggingv1.ElasticsearchNode)
	for nodeIndex, node := range er.cluster.Spec.Nodes {
		if node.GenUUID == nil {
			// keep track of the index it was at so we can correctly update it later
			nodesToMatch[nodeIndex] = node
		}
	}

	if len(nodesToMatch) > 0 {
		// first get a list of known UUIDs
		knownUUIDs := []string{}
		for _, node := range er.cluster.Spec.Nodes {
			if node.GenUUID != nil {
				knownUUIDs = append(knownUUIDs, *node.GenUUID)
			}
		}

		// collect uuid counts
		selector := map[string]string{}
		pvcList, err := GetPVCList(er.cluster.Namespace, selector, er.client)
		if err != nil {
			logrus.Errorf("Unable to retrieve PVC list while recovering cluster %q in namespace %q: %v", er.cluster.Name, er.cluster.Namespace, err)
			return err
		}

		if len(pvcList.Items) > 0 {
			return er.recoverFromPVCs(knownUUIDs, nodesToMatch)
		} else {
			return er.recoverFromDeployments(knownUUIDs, nodesToMatch)
		}
	}

	return nil
}

func (er *ElasticsearchRequest) deploymentSpecMatchNode(node loggingv1.ElasticsearchNode, deployment appsv1.DeploymentSpec, count int32) bool {

	// we know roles will match, first check node count
	if node.NodeCount != count {
		return false
	}

	return er.podSpecMatchNode(node, deployment.Template.Spec)
}

func (er *ElasticsearchRequest) statefulSetSpecMatchNode(node loggingv1.ElasticsearchNode, statefulset appsv1.StatefulSetSpec) bool {

	// we know roles will match, first check node count
	if node.NodeCount != *statefulset.Replicas {
		return false
	}

	return er.podSpecMatchNode(node, statefulset.Template.Spec)
}

func (er *ElasticsearchRequest) podSpecMatchNode(node loggingv1.ElasticsearchNode, podSpec corev1.PodSpec) bool {

	selectors := mergeSelectors(node.NodeSelector, er.cluster.Spec.Spec.NodeSelector)
	selectors = utils.EnsureLinuxNodeSelector(selectors)

	if !areSelectorsSame(selectors, podSpec.NodeSelector) {
		return false
	}

	tolerations := appendTolerations(node.Tolerations, er.cluster.Spec.Spec.Tolerations)

	if !containsSameTolerations(podSpec.Tolerations, tolerations) {
		return false
	}

	nodeResources := newESResourceRequirements(node.Resources, er.cluster.Spec.Spec.Resources)

	var deploymentNodeResources corev1.ResourceRequirements

	for _, container := range podSpec.Containers {
		if container.Name == "elasticsearch" {
			deploymentNodeResources = container.Resources
		}
	}

	// make sure resources are same
	if different, _ := utils.CompareResources(nodeResources, deploymentNodeResources); different {
		return false
	}

	return true
}

func parseNodeName(name string) (clusterName, roles, uuid string) {

	splitName := strings.Split(name, "-")

	// deployment/statefulset names
	if len(splitName) == 4 {
		clusterName = splitName[0]
		roles = splitName[1]
		uuid = splitName[2]

		return
	}

	// the case of the old PVC name
	if len(splitName) == 5 {
		clusterName = splitName[1]
		roles = splitName[2]
		uuid = splitName[3]

		return
	}

	return
}

func (er *ElasticsearchRequest) recoverFromDeployments(knownUUIDs []string, nodesToMatch map[int]loggingv1.ElasticsearchNode) error {

	uuidCounts := make(map[string]int32)
	if len(nodesToMatch) > 0 {
		// collect uuid counts
		selector := map[string]string{
			"cluster-name": er.cluster.Name,
		}

		deploymentList, err := GetDeploymentList(er.cluster.Namespace, selector, er.client)
		if err != nil {
			logrus.Errorf("Unable to retrieve Deployment list while recovering cluster %q in namespace %q: %v", er.cluster.Name, er.cluster.Namespace, err)
			return err
		}

		for _, deployment := range deploymentList.Items {
			clusterName, _, uuid := parseNodeName(deployment.Name)

			if clusterName != er.cluster.Name {
				continue
			}

			if value, ok := uuidCounts[uuid]; ok {
				uuidCounts[uuid] = value + 1
			} else {
				uuidCounts[uuid] = 1
			}
		}
	}

	// slight misnomer -- the key is the index which refers back to its index for er.cluster.Spec.Nodes
	for nodeIndex, node := range nodesToMatch {

		selector := map[string]string{
			"cluster-name":   er.cluster.Name,
			"es-node-client": "false",
			"es-node-data":   "false",
			"es-node-master": "false",
		}

		for _, role := range node.Roles {
			switch role {
			case loggingv1.ElasticsearchRoleClient:
				selector["es-node-client"] = "true"
				break
			case loggingv1.ElasticsearchRoleData:
				selector["es-node-data"] = "true"
				break
			case loggingv1.ElasticsearchRoleMaster:
				selector["es-node-master"] = "true"
			}
		}

		if isDataNode(node) {
			var deploymentList *appsv1.DeploymentList
			deploymentList, err := GetDeploymentList(er.cluster.Namespace, selector, er.client)
			if err != nil {
				logrus.Errorf("Unable to retrieve Deployment list while recovering cluster %q in namespace %q: %v", er.cluster.Name, er.cluster.Namespace, err)
				return err
			}

			for _, deployment := range deploymentList.Items {
				clusterName, _, uuid := parseNodeName(deployment.Name)

				if clusterName != er.cluster.Name {
					continue
				}

				if sliceContainsString(knownUUIDs, uuid) {
					logrus.Infof("already found %q in %v while adopting", uuid, knownUUIDs)
					continue
				}

				if er.cluster.Spec.Nodes[nodeIndex].GenUUID == nil {
					if er.deploymentSpecMatchNode(node, deployment.Spec, uuidCounts[uuid]) {
						er.cluster.Spec.Nodes[nodeIndex].GenUUID = &uuid
						knownUUIDs = append(knownUUIDs, uuid)

						er.setUUID(nodeIndex, uuid)
						break
					}
				}
			}
		} else {
			var statefulsetList *appsv1.StatefulSetList
			statefulsetList, err := GetStatefulSetList(er.cluster.Namespace, selector, er.client)
			if err != nil {
				logrus.Errorf("Unable to retrieve Statefulset list while recovering cluster %q in namespace %q: %v", er.cluster.Name, er.cluster.Namespace, err)
				return err
			}

			for _, statefulSet := range statefulsetList.Items {
				clusterName, _, uuid := parseNodeName(statefulSet.Name)

				if clusterName != er.cluster.Name {
					continue
				}

				if sliceContainsString(knownUUIDs, uuid) {
					logrus.Infof("already found %q in %v while adopting", uuid, knownUUIDs)
					continue
				}

				if er.cluster.Spec.Nodes[nodeIndex].GenUUID == nil {
					if er.statefulSetSpecMatchNode(node, statefulSet.Spec) {
						er.cluster.Spec.Nodes[nodeIndex].GenUUID = &uuid
						knownUUIDs = append(knownUUIDs, uuid)

						er.setUUID(nodeIndex, uuid)
						break
					}
				}
			}
		}
	}

	return nil
}

// for PVCs we only need to match on roles and replicas for the sake of naming
func (er *ElasticsearchRequest) recoverFromPVCs(knownUUIDs []string, nodesToMatch map[int]loggingv1.ElasticsearchNode) error {

	selector := map[string]string{
		"logging-cluster": er.cluster.Name,
	}

	pvcList, err := GetPVCList(er.cluster.Namespace, selector, er.client)

	if err != nil {
		logrus.Errorf("Unable to retrieve PVC list while recovering cluster %q in namespace %q: %v", er.cluster.Name, er.cluster.Namespace, err)
		return err
	}

	uuidCounts := make(map[string]int32)
	for _, pvc := range pvcList.Items {

		clusterName, _, uuid := parseNodeName(pvc.Name)

		if clusterName != er.cluster.Name {
			continue
		}

		if value, ok := uuidCounts[uuid]; ok {
			uuidCounts[uuid] = value + 1
		} else {
			uuidCounts[uuid] = 1
		}
	}

	// go through the nodesToMatch and match it based on the roles for the pvc
	for nodeIndex, node := range nodesToMatch {

		// if the node doesn't have storage defined, skip it
		if node.Storage.StorageClassName == nil {
			continue
		}

		isClientNode := false
		isDataNode := false
		isMasterNode := false

		for _, role := range node.Roles {
			switch role {
			case loggingv1.ElasticsearchRoleClient:
				isClientNode = true
				break
			case loggingv1.ElasticsearchRoleData:
				isDataNode = true
				break
			case loggingv1.ElasticsearchRoleMaster:
				isMasterNode = true
			}
		}

		for _, pvc := range pvcList.Items {

			clusterName, role, uuid := parseNodeName(pvc.Name)

			if clusterName != er.cluster.Name {
				continue
			}

			if sliceContainsString(knownUUIDs, uuid) {
				logrus.Infof("already found %q in %v while adopting", uuid, knownUUIDs)
				continue
			}

			// check that roles are same, if not then continue
			if isClientNode != strings.Contains(role, "c") {
				continue
			}

			if isDataNode != strings.Contains(role, "d") {
				continue
			}

			if isMasterNode != strings.Contains(role, "m") {
				continue
			}

			if node.NodeCount != uuidCounts[uuid] {
				continue
			}

			// roles and node count match, reuse the UUID
			er.cluster.Spec.Nodes[nodeIndex].GenUUID = &uuid
			knownUUIDs = append(knownUUIDs, uuid)

			er.setUUID(nodeIndex, uuid)
		}
	}

	return nil
}
