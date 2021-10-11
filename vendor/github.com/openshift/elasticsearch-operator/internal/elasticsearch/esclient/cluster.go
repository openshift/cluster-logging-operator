package esclient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ViaQ/logerr/kverrors"
	estypes "github.com/openshift/elasticsearch-operator/internal/types/elasticsearch"
	"github.com/openshift/elasticsearch-operator/internal/utils/comparators"
)

func (ec *esClient) GetClusterNodeVersions() ([]string, error) {
	payload := &EsRequest{
		Method: http.MethodGet,
		URI:    "_cluster/stats",
	}

	ec.fnSendEsRequest(ec.cluster, ec.namespace, payload, ec.k8sClient)

	var nodeVersions []string
	if versions := walkInterfaceMap("nodes.versions", payload.ResponseBody); versions != nil {
		for _, value := range versions.([]interface{}) {
			version := value.(string)
			nodeVersions = append(nodeVersions, version)
		}
	}

	return nodeVersions, nil
}

func (ec *esClient) GetThresholdEnabled() (bool, error) {
	payload := &EsRequest{
		Method: http.MethodGet,
		URI:    "_cluster/settings?include_defaults=true",
	}

	ec.fnSendEsRequest(ec.cluster, ec.namespace, payload, ec.k8sClient)

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

func (ec *esClient) GetDiskWatermarks() (interface{}, interface{}, interface{}, error) {
	payload := &EsRequest{
		Method: http.MethodGet,
		URI:    "_cluster/settings?include_defaults=true",
	}

	ec.fnSendEsRequest(ec.cluster, ec.namespace, payload, ec.k8sClient)

	var low interface{}
	var high interface{}
	var flood interface{}

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
		"defaults.cluster.routing.allocation.disk.watermark.flood_stage",
		payload.ResponseBody); value != nil {
		flood = value
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
		"persistent.cluster.routing.allocation.disk.watermark.flood_stage",
		payload.ResponseBody); value != nil {
		flood = value
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

	if value := walkInterfaceMap(
		"transient.cluster.routing.allocation.disk.watermark.flood_stage",
		payload.ResponseBody); value != nil {
		flood = value
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

	if floodString, ok := flood.(string); ok {
		if strings.HasSuffix(floodString, "%") {
			flood, _ = strconv.ParseFloat(strings.TrimSuffix(floodString, "%"), 64)
		} else {
			if strings.HasSuffix(floodString, "b") {
				flood = strings.TrimSuffix(floodString, "b")
			}
		}
	}

	return low, high, flood, payload.Error
}

func (ec *esClient) SetMinMasterNodes(numberMasters int32) (bool, error) {
	payload := &EsRequest{
		Method:      http.MethodPut,
		URI:         "_cluster/settings",
		RequestBody: fmt.Sprintf("{%q:{%q:%d}}", "persistent", "discovery.zen.minimum_master_nodes", numberMasters),
	}

	ec.fnSendEsRequest(ec.cluster, ec.namespace, payload, ec.k8sClient)

	acknowledged := false
	if acknowledgedBool, ok := payload.ResponseBody["acknowledged"].(bool); ok {
		acknowledged = acknowledgedBool
	}

	return payload.StatusCode == 200 && acknowledged, payload.Error
}

func (ec *esClient) GetMinMasterNodes() (int32, error) {
	payload := &EsRequest{
		Method: http.MethodGet,
		URI:    "_cluster/settings",
	}

	ec.fnSendEsRequest(ec.cluster, ec.namespace, payload, ec.k8sClient)

	masterCount := int32(0)
	if payload.ResponseBody["persistent"] != nil {
		persistentBody := payload.ResponseBody["persistent"].(map[string]interface{})
		if masterCountFloat, ok := persistentBody["discovery.zen.minimum_master_nodes"].(float64); ok {
			masterCount = int32(masterCountFloat)
		}
	}

	return masterCount, payload.Error
}

// TODO: also check that the number of shards in the response > 0?
func (ec *esClient) DoSynchronizedFlush() (bool, error) {
	payload := &EsRequest{
		Method: http.MethodPost,
		URI:    "_flush/synced",
	}

	ec.fnSendEsRequest(ec.cluster, ec.namespace, payload, ec.k8sClient)

	failed := 0
	if shards, ok := payload.ResponseBody["_shards"].(map[string]interface{}); ok {
		if failedFload, ok := shards["failed"].(float64); ok {
			failed = int(failedFload)
		}
	}

	if payload.Error == nil && failed != 0 {
		payload.Error = kverrors.New("failed to flush shards in preparation for cluster restart",
			"num_failed_shards", failed)
	}

	return payload.StatusCode == 200, payload.Error
}

func (ec *esClient) GetLowestClusterVersion() (string, error) {
	payload := &EsRequest{
		Method: http.MethodGet,
		URI:    "_cluster/stats/nodes/_all",
	}

	ec.fnSendEsRequest(ec.cluster, ec.namespace, payload, ec.k8sClient)
	if payload.Error != nil {
		return "", payload.Error
	}
	if payload.StatusCode != http.StatusOK {
		return "", ec.errorCtx().New("failed to get cluster state",
			"response_code", payload.StatusCode,
			"response_body", payload.ResponseBody)
	}

	res := &estypes.StatsNodesResponse{}
	err := json.Unmarshal([]byte(payload.RawResponseBody), res)
	if err != nil {
		return "", ec.errorCtx().Wrap(err, "failed to decode raw response body into `estypes.StatsNodesResponse`")
	}

	if len(res.Nodes.Versions) == 0 {
		return "", ec.errorCtx().New("received no node versions from cluster")
	}

	if len(res.Nodes.Versions) == 1 {
		return res.Nodes.Versions[0], nil
	}

	lowestVersion := res.Nodes.Versions[0]
	for _, version := range res.Nodes.Versions {
		comparison := comparators.CompareVersions(lowestVersion, version)

		if comparison < 0 {
			lowestVersion = version
		}
	}

	return lowestVersion, nil
}

func (ec *esClient) IsNodeInCluster(nodeName string) (bool, error) {
	payload := &EsRequest{
		Method: http.MethodGet,
		URI:    "_cluster/state/nodes",
	}

	ec.fnSendEsRequest(ec.cluster, ec.namespace, payload, ec.k8sClient)
	if payload.Error != nil {
		return false, payload.Error
	}
	if payload.StatusCode != http.StatusOK {
		return false, ec.errorCtx().New("failed to get cluster state",
			"response_status", payload.StatusCode,
			"response_body", payload.ResponseBody,
		)
	}

	res := &estypes.NodesStateResponse{}
	err := json.Unmarshal([]byte(payload.RawResponseBody), res)
	if err != nil {
		return false, ec.errorCtx().Wrap(err, "failed to decode raw response body into `estypes.NodesStateResponse`")
	}

	for _, node := range res.Nodes {
		if node.Name == nodeName {
			return true, nil
		}
	}

	return false, nil
}
