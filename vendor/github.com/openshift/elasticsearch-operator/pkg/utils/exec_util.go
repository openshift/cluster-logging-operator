package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/openshift/elasticsearch-operator/pkg/apis/elasticsearch/v1alpha1"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
)

func ElasticsearchExec(pod *v1.Pod, command []string) (*bytes.Buffer, *bytes.Buffer, error) {
	// when running in a pod, use the values provided for the sa
	// this is primarily used when testing
	kubeConfigPath := LookupEnvWithDefault("KUBERNETES_CONFIG", "")
	masterURL := "https://kubernetes.default.svc"
	if kubeConfigPath == "" {
		// ExecConfig requires both are "", or both have a real value
		masterURL = ""
	}
	config := &ExecConfig{
		Pod:            pod,
		ContainerName:  "elasticsearch",
		Command:        command,
		KubeConfigPath: kubeConfigPath,
		MasterURL:      masterURL,
		StdOut:         true,
		StdErr:         true,
		Tty:            false,
	}
	return PodExec(config)
}

func UpdateClusterSettings(pod *v1.Pod, quorum int) error {
	command := []string{"sh", "-c",
		fmt.Sprintf("es_util --query=_cluster/settings -H 'Content-Type: application/json' -X PUT -d '{\"persistent\":{%s}}'",
			minimumMasterNodesCommand(quorum))}

	_, _, err := ElasticsearchExec(pod, command)

	return err
}

func ClusterHealth(pod *v1.Pod) (map[string]interface{}, error) {
	command := []string{"es_util", "--query=_cluster/health?pretty=true"}
	execOut, _, err := ElasticsearchExec(pod, command)
	if err != nil {
		logrus.Debug(err)
		return nil, err
	}

	var result map[string]interface{}

	err = json.Unmarshal(execOut.Bytes(), &result)
	if err != nil {
		logrus.Debug("could not unmarshal: %v", err)
		return nil, err
	}
	return result, nil
}

func NumberOfNodes(pod *v1.Pod) int {
	healthResponse, err := ClusterHealth(pod)
	if err != nil {
		// logrus.Debugf("failed to get _cluster/health: %v", err)
		return -1
	}

	// is it present?
	value, present := healthResponse["number_of_nodes"]
	if !present {
		return -1
	}

	// json numbers are represented as floats
	// so let's convert from type interface{} to float
	numberofNodes, ok := value.(float64)
	if !ok {
		return -1
	}

	// wow that's a lot of boilerplate...
	return int(numberofNodes)
}

func PerformSyncedFlush(pod *v1.Pod) error {
	command := []string{"sh", "-c", "es_util --query=_flush/synced -X POST"}

	_, _, err := ElasticsearchExec(pod, command)

	return err
}

func SetShardAllocation(pod *v1.Pod, enabled v1alpha1.ShardAllocationState) error {
	command := []string{"sh", "-c",
		fmt.Sprintf("es_util --query=_cluster/settings -H 'Content-Type: application/json' -X PUT -d '{\"transient\":{%s}}'",
			shardAllocationCommand(enabled))}

	_, _, err := ElasticsearchExec(pod, command)

	return err
}

func shardAllocationCommand(shardAllocation v1alpha1.ShardAllocationState) string {
	return fmt.Sprintf("%s:%s", strconv.Quote("cluster.routing.allocation.enable"), strconv.Quote(string(shardAllocation)))
}

func minimumMasterNodesCommand(nodes int) string {
	return fmt.Sprintf("%s:%d", strconv.Quote("discovery.zen.minimum_master_nodes"), nodes)
}
