package k8shandler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/inhies/go-bytesize"
	api "github.com/openshift/elasticsearch-operator/pkg/apis/logging/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SetShardAllocation(clusterName, namespace string, state api.ShardAllocationState, client client.Client) (bool, error) {

	payload := &esCurlStruct{
		Method:      http.MethodPut,
		URI:         "_cluster/settings",
		RequestBody: fmt.Sprintf("{%q:{%q:%q}}", "transient", "cluster.routing.allocation.enable", state),
	}

	curlESService(clusterName, namespace, payload, client)

	acknowledged := false
	if acknowledgedBool, ok := payload.ResponseBody["acknowledged"].(bool); ok {
		acknowledged = acknowledgedBool
	}
	return (payload.StatusCode == 200 && acknowledged), payload.Error
}

func GetShardAllocation(clusterName, namespace string, client client.Client) (string, error) {

	payload := &esCurlStruct{
		Method: http.MethodGet,
		URI:    "_cluster/settings",
	}

	curlESService(clusterName, namespace, payload, client)

	allocation := parseString("transient.cluster.routing.allocation.enable", payload.ResponseBody)

	return allocation, payload.Error
}

func GetNodeDiskUsage(clusterName, namespace, nodeName string, client client.Client) (string, float64, error) {
	payload := &esCurlStruct{
		Method: http.MethodGet,
		URI:    "_nodes/stats/fs",
	}

	curlESService(clusterName, namespace, payload, client)

	usage := ""
	percentUsage := float64(-1)

	if payload, ok := payload.ResponseBody["nodes"].(map[string]interface{}); ok {
		for _, stats := range payload {

			// ignore the key name here, it is the node UUID
			if parseString("name", stats.(map[string]interface{})) == nodeName {
				total := parseFloat64("fs.total.total_in_bytes", stats.(map[string]interface{}))
				available := parseFloat64("fs.total.available_in_bytes", stats.(map[string]interface{}))

				percentUsage = (total - available) / total * 100.00
				usage = strings.TrimSuffix(fmt.Sprintf("%s", bytesize.New(total)-bytesize.New(available)), "B")

				break
			}
		}
	}

	return usage, percentUsage, payload.Error
}

func GetThresholdEnabled(clusterName, namespace string, client client.Client) (bool, error) {

	payload := &esCurlStruct{
		Method: http.MethodGet,
		URI:    "_cluster/settings?include_defaults=true",
	}

	curlESService(clusterName, namespace, payload, client)

	var enabled interface{}

	if value := walkInterfaceMap(
		"defaults.cluster.routing.allocation.disk.threshold_enabled",
		payload.ResponseBody); value != nil {

		enabled = value
	}

	if value := walkInterfaceMap(
		"persistent.cluster.routing.allocation.disk.threshold_enabled",
		payload.ResponseBody); value != nil {

		enabled = value
	}

	if value := walkInterfaceMap(
		"transient.cluster.routing.allocation.disk.threshold_enabled",
		payload.ResponseBody); value != nil {

		enabled = value
	}

	enabledBool := false
	if enabledString, ok := enabled.(string); ok {
		if enabledTemp, err := strconv.ParseBool(enabledString); err == nil {
			enabledBool = enabledTemp
		}
	}

	return enabledBool, payload.Error
}

func GetDiskWatermarks(clusterName, namespace string, client client.Client) (interface{}, interface{}, error) {

	payload := &esCurlStruct{
		Method: http.MethodGet,
		URI:    "_cluster/settings?include_defaults=true",
	}

	curlESService(clusterName, namespace, payload, client)

	var low interface{}
	var high interface{}

	if value := walkInterfaceMap(
		"defaults.cluster.routing.allocation.disk.watermark.low",
		payload.ResponseBody); value != nil {

		low = value
	}

	if value := walkInterfaceMap(
		"defaults.cluster.routing.allocation.disk.watermark.high",
		payload.ResponseBody); value != nil {

		high = value
	}

	if value := walkInterfaceMap(
		"persistent.cluster.routing.allocation.disk.watermark.low",
		payload.ResponseBody); value != nil {

		low = value
	}

	if value := walkInterfaceMap(
		"persistent.cluster.routing.allocation.disk.watermark.high",
		payload.ResponseBody); value != nil {

		high = value
	}

	if value := walkInterfaceMap(
		"transient.cluster.routing.allocation.disk.watermark.low",
		payload.ResponseBody); value != nil {

		low = value
	}

	if value := walkInterfaceMap(
		"transient.cluster.routing.allocation.disk.watermark.high",
		payload.ResponseBody); value != nil {

		high = value
	}

	if lowString, ok := low.(string); ok {
		if strings.HasSuffix(lowString, "%") {
			low, _ = strconv.ParseFloat(strings.TrimSuffix(lowString, "%"), 64)
		} else {
			if strings.HasSuffix(lowString, "b") {
				low = strings.TrimSuffix(lowString, "b")
			}
		}
	}

	if highString, ok := high.(string); ok {
		if strings.HasSuffix(highString, "%") {
			high, _ = strconv.ParseFloat(strings.TrimSuffix(highString, "%"), 64)
		} else {
			if strings.HasSuffix(highString, "b") {
				high = strings.TrimSuffix(highString, "b")
			}
		}
	}

	return low, high, payload.Error
}

func SetMinMasterNodes(clusterName, namespace string, numberMasters int32, client client.Client) (bool, error) {

	payload := &esCurlStruct{
		Method:      http.MethodPut,
		URI:         "_cluster/settings",
		RequestBody: fmt.Sprintf("{%q:{%q:%d}}", "persistent", "discovery.zen.minimum_master_nodes", numberMasters),
	}

	curlESService(clusterName, namespace, payload, client)

	acknowledged := false
	if acknowledgedBool, ok := payload.ResponseBody["acknowledged"].(bool); ok {
		acknowledged = acknowledgedBool
	}

	return (payload.StatusCode == 200 && acknowledged), payload.Error
}

func GetMinMasterNodes(clusterName, namespace string, client client.Client) (int32, error) {

	payload := &esCurlStruct{
		Method: http.MethodGet,
		URI:    "_cluster/settings",
	}

	curlESService(clusterName, namespace, payload, client)

	masterCount := int32(0)
	if payload.ResponseBody["persistent"] != nil {
		persistentBody := payload.ResponseBody["persistent"].(map[string]interface{})
		if masterCountFloat, ok := persistentBody["discovery.zen.minimum_master_nodes"].(float64); ok {
			masterCount = int32(masterCountFloat)
		}
	}

	return masterCount, payload.Error
}

func GetClusterHealth(clusterName, namespace string, client client.Client) (api.ClusterHealth, error) {

	clusterHealth := api.ClusterHealth{}

	payload := &esCurlStruct{
		Method: http.MethodGet,
		URI:    "_cluster/health",
	}

	curlESService(clusterName, namespace, payload, client)

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

func GetClusterHealthStatus(clusterName, namespace string, client client.Client) (string, error) {

	payload := &esCurlStruct{
		Method: http.MethodGet,
		URI:    "_cluster/health",
	}

	curlESService(clusterName, namespace, payload, client)

	status := ""
	if payload.ResponseBody["status"] != nil {
		if statusString, ok := payload.ResponseBody["status"].(string); ok {
			status = statusString
		}
	}

	return status, payload.Error
}

func GetClusterNodeCount(clusterName, namespace string, client client.Client) (int32, error) {

	payload := &esCurlStruct{
		Method: http.MethodGet,
		URI:    "_cluster/health",
	}

	curlESService(clusterName, namespace, payload, client)

	nodeCount := int32(0)
	if nodeCountFloat, ok := payload.ResponseBody["number_of_nodes"].(float64); ok {
		// we expect at most double digit numbers here, eg cluster with 15 nodes
		nodeCount = int32(nodeCountFloat)
	}

	return nodeCount, payload.Error
}

// TODO: also check that the number of shards in the response > 0?
func DoSynchronizedFlush(clusterName, namespace string, client client.Client) (bool, error) {

	payload := &esCurlStruct{
		Method: http.MethodPost,
		URI:    "_flush/synced",
	}

	curlESService(clusterName, namespace, payload, client)

	failed := 0
	if shards, ok := payload.ResponseBody["_shards"].(map[string]interface{}); ok {
		if failedFload, ok := shards["failed"].(float64); ok {
			failed = int(failedFload)
		}
	}

	if payload.Error == nil && failed != 0 {
		payload.Error = fmt.Errorf("Failed to flush %d shards in preparation for cluster restart", failed)
	}

	return (payload.StatusCode == 200), payload.Error
}

// This will idempotently update the index templates and update indices' replica count
func UpdateReplicaCount(clusterName, namespace string, client client.Client, replicaCount int32) (bool, error) {

	if ok, _ := updateAllIndexTemplateReplicas(clusterName, namespace, client, replicaCount); ok {
		if ok, _ = updateAllIndexReplicas(clusterName, namespace, client, replicaCount); ok {
			return true, nil
		}
	}

	return false, nil
}
