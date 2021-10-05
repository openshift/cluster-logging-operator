package elasticsearch

import (
	"fmt"
	"net/http"

	api "github.com/openshift/elasticsearch-operator/apis/logging/v1"
)

func (ec *esClient) ClearTransientShardAllocation() (bool, error) {
	payload := &EsRequest{
		Method:      http.MethodPut,
		URI:         "_cluster/settings",
		RequestBody: fmt.Sprintf("{%q:{%q:null}}", "transient", "cluster.routing.allocation.enable"),
	}

	ec.fnSendEsRequest(ec.cluster, ec.namespace, payload, ec.k8sClient)

	acknowledged := false
	if acknowledgedBool, ok := payload.ResponseBody["acknowledged"].(bool); ok {
		acknowledged = acknowledgedBool
	}
	return payload.StatusCode == 200 && acknowledged, ec.errorCtx().Wrap(payload.Error, "failed to clear shard allocation",
		"response", payload.RawResponseBody)
}

func (ec *esClient) SetShardAllocation(state api.ShardAllocationState) (bool, error) {
	payload := &EsRequest{
		Method:      http.MethodPut,
		URI:         "_cluster/settings",
		RequestBody: fmt.Sprintf("{%q:{%q:%q}}", "persistent", "cluster.routing.allocation.enable", state),
	}

	ec.fnSendEsRequest(ec.cluster, ec.namespace, payload, ec.k8sClient)

	acknowledged := false
	if acknowledgedBool, ok := payload.ResponseBody["acknowledged"].(bool); ok {
		acknowledged = acknowledgedBool
	}
	return payload.StatusCode == 200 && acknowledged, ec.errorCtx().Wrap(payload.Error, "failed to set shard allocation")
}

func (ec *esClient) GetShardAllocation() (string, error) {
	payload := &EsRequest{
		Method: http.MethodGet,
		URI:    "_cluster/settings?include_defaults=true",
	}

	ec.fnSendEsRequest(ec.cluster, ec.namespace, payload, ec.k8sClient)

	var allocation interface{}

	if value := walkInterfaceMap(
		"defaults.cluster.routing.allocation.enable",
		payload.ResponseBody); value != nil {
		allocation = value
	}

	if value := walkInterfaceMap(
		"persistent.cluster.routing.allocation.enable",
		payload.ResponseBody); value != nil {
		allocation = value
	}

	if value := walkInterfaceMap(
		"transient.cluster.routing.allocation.enable",
		payload.ResponseBody); value != nil {
		allocation = value
	}

	allocationString, ok := allocation.(string)
	if !ok {
		allocationString = ""
	}

	return allocationString, payload.Error
}
