package k8shandler

import (
	"encoding/json"
	"fmt"

	v1alpha1 "github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
)

const healthUnknown = "cluster health unknown"

// UpdateStatus updates the status of Elasticsearch CRD
func (cState *ClusterState) UpdateStatus(dpl *v1alpha1.Elasticsearch) error {
	dpl.Status.ClusterHealth = clusterHealth(dpl)
	dpl.Status.Nodes = []v1alpha1.ElasticsearchNodeStatus{}
	for _, node := range cState.Nodes {
		updateNodeStatus(node, &dpl.Status)
	}

	dpl.Status.Pods = rolePodStateMap(dpl.Namespace, dpl.Name)
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

	if node.Desired.Roles != nil {
		nodeStatus.Roles = node.Desired.Roles
	}
	dpl.Nodes = append(dpl.Nodes, nodeStatus)
}

func clusterHealth(dpl *v1alpha1.Elasticsearch) string {
	pods, err := listRunningPods(dpl.Name, dpl.Namespace)
	if err != nil {
		return healthUnknown
	}

	// no running elasticsearch pods were found
	if len(pods.Items) == 0 {
		return ""
	}

	// use arbitrary pod
	pod := pods.Items[0]

	config := &ExecConfig{
		pod:            &pod,
		containerName:  "elasticsearch",
		command:        []string{"es_util", "--query=_cluster/health?pretty=true"},
		kubeConfigPath: lookupEnvWithDefault("KUBERNETES_CONFIG", "/etc/origin/master/admin.kubeconfig"),
		masterURL:      "https://kubernetes.default.svc",
		stdOut:         true,
		stdErr:         true,
		tty:            false,
	}

	execOut, _, err := PodExec(config)
	if err != nil {
		logrus.Debug(err)
		return healthUnknown
	}

	var result map[string]interface{}

	err = json.Unmarshal(execOut.Bytes(), &result)
	if err != nil {
		logrus.Debug("could not unmarshal: %v", err)
		return healthUnknown
	}
	if _, present := result["status"]; !present {
		logrus.Debug("response from elasticsearch health API did not contain 'status' field")
		return healthUnknown
	}

	return result["status"].(string)
}

func rolePodStateMap(namespace string, clusterName string) map[v1alpha1.ElasticsearchNodeRole]v1alpha1.PodStateMap {

	baseSelector := fmt.Sprintf("component=%s", clusterName)
	clientList, _ := GetPodList(namespace, fmt.Sprintf("%s,%s", baseSelector, "es-node-client=true"))
	dataList, _ := GetPodList(namespace, fmt.Sprintf("%s,%s", baseSelector, "es-node-data=true"))
	masterList, _ := GetPodList(namespace, fmt.Sprintf("%s,%s", baseSelector, "es-node-master=true"))

	return map[v1alpha1.ElasticsearchNodeRole]v1alpha1.PodStateMap{
		v1alpha1.ElasticsearchRoleClient: podStateMap(clientList.Items),
		v1alpha1.ElasticsearchRoleData:   podStateMap(dataList.Items),
		v1alpha1.ElasticsearchRoleMaster: podStateMap(masterList.Items),
	}
}

func podStateMap(podList []v1.Pod) v1alpha1.PodStateMap {
	stateMap := map[v1alpha1.PodStateType][]string{
		v1alpha1.PodStateTypeReady:    []string{},
		v1alpha1.PodStateTypeNotReady: []string{},
		v1alpha1.PodStateTypeFailed:   []string{},
	}

	for _, pod := range podList {
		switch pod.Status.Phase {
		case v1.PodPending:
			stateMap[v1alpha1.PodStateTypeNotReady] = append(stateMap[v1alpha1.PodStateTypeNotReady], pod.Name)
		case v1.PodRunning:
			if isPodReady(pod) {
				stateMap[v1alpha1.PodStateTypeReady] = append(stateMap[v1alpha1.PodStateTypeReady], pod.Name)
			} else {
				stateMap[v1alpha1.PodStateTypeNotReady] = append(stateMap[v1alpha1.PodStateTypeNotReady], pod.Name)
			}
		case v1.PodFailed:
			stateMap[v1alpha1.PodStateTypeFailed] = append(stateMap[v1alpha1.PodStateTypeFailed], pod.Name)
		}
	}

	return stateMap
}

func isPodReady(pod v1.Pod) bool {

	for _, container := range pod.Status.ContainerStatuses {
		if !container.Ready {
			return false
		}
	}

	return true
}
