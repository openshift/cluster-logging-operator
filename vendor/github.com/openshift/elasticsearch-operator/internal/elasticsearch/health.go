package elasticsearch

import (
	"net/http"

	api "github.com/openshift/elasticsearch-operator/apis/logging/v1"
)

func (ec *esClient) GetClusterHealth() (api.ClusterHealth, error) {
	clusterHealth := api.ClusterHealth{}

	payload := &EsRequest{
		Method: http.MethodGet,
		URI:    "_cluster/health",
	}

	ec.fnSendEsRequest(ec.cluster, ec.namespace, payload, ec.k8sClient)

	if payload.Error != nil {
		return clusterHealth, payload.Error
	}

	clusterHealth.Status = parseString("status", payload.ResponseBody)
	clusterHealth.NumNodes = parseInt32("number_of_nodes", payload.ResponseBody)
	clusterHealth.NumDataNodes = parseInt32("number_of_data_nodes", payload.ResponseBody)
	clusterHealth.ActivePrimaryShards = parseInt32("active_primary_shards", payload.ResponseBody)
	clusterHealth.ActiveShards = parseInt32("active_shards", payload.ResponseBody)
	clusterHealth.RelocatingShards = parseInt32("relocating_shards", payload.ResponseBody)
	clusterHealth.InitializingShards = parseInt32("initializing_shards", payload.ResponseBody)
	clusterHealth.UnassignedShards = parseInt32("unassigned_shards", payload.ResponseBody)
	clusterHealth.PendingTasks = parseInt32("number_of_pending_tasks", payload.ResponseBody)

	return clusterHealth, nil
}

func (ec *esClient) GetClusterHealthStatus() (string, error) {
	payload := &EsRequest{
		Method: http.MethodGet,
		URI:    "_cluster/health",
	}

	ec.fnSendEsRequest(ec.cluster, ec.namespace, payload, ec.k8sClient)

	status := ""
	if payload.ResponseBody["status"] != nil {
		if statusString, ok := payload.ResponseBody["status"].(string); ok {
			status = statusString
		}
	}

	return status, payload.Error
}

func (ec *esClient) GetClusterNodeCount() (int32, error) {
	payload := &EsRequest{
		Method: http.MethodGet,
		URI:    "_cluster/health",
	}

	ec.fnSendEsRequest(ec.cluster, ec.namespace, payload, ec.k8sClient)

	nodeCount := int32(0)
	if nodeCountFloat, ok := payload.ResponseBody["number_of_nodes"].(float64); ok {
		// we expect at most double digit numbers here, eg cluster with 15 nodes
		nodeCount = int32(nodeCountFloat)
	}

	return nodeCount, payload.Error
}
