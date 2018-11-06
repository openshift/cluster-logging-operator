package k8shandler

import (
	"encoding/json"
	"fmt"

	v1alpha1 "github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
)

const healthUnknown = "cluster health unknown"

// UpdateStatus updates the status of Elasticsearch CRD
func (cState *ClusterState) UpdateStatus(dpl *v1alpha1.Elasticsearch) error {
	dpl.Status.ClusterHealth = clusterHealth(dpl)
	dpl.Status.Nodes = []v1alpha1.ElasticsearchNodeStatus{}
	for _, node := range cState.Nodes {
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
