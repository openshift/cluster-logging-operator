package esclient

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
)

// This will idempotently update the index templates and update indices' replica count
func (ec *esClient) UpdateReplicaCount(replicaCount int32) error {
	if ok, _ := ec.updateAllIndexTemplateReplicas(replicaCount); ok {
		if _, err := ec.updateAllIndexReplicas(replicaCount); err != nil {
			return err
		}
	}
	return nil
}

func (ec *esClient) updateAllIndexReplicas(replicaCount int32) (bool, error) {
	indexHealth, _ := ec.GetIndexReplicaCounts()

	// get list of indices and call updateIndexReplicas for each one
	for index, health := range indexHealth {
		if healthMap, ok := health.(map[string]interface{}); ok {
			// only update replicas for indices that don't have same replica count
			if numberOfReplicas := parseString("settings.index.number_of_replicas", healthMap); numberOfReplicas != "" {
				currentReplicas, err := strconv.ParseInt(numberOfReplicas, 10, 32)
				if err != nil {
					return false, err
				}

				if int32(currentReplicas) != replicaCount {
					// best effort initially?
					if ack, err := ec.updateIndexReplicas(index, replicaCount); err != nil {
						return ack, err
					}
				}
			}
		} else {
			return false, ec.errorCtx().New("unable to evaluate the number of replicas for index",
				"index", index,
				"health", health,
			)
		}
	}

	return true, nil
}

func (ec *esClient) GetIndexReplicaCounts() (map[string]interface{}, error) {
	payload := &EsRequest{
		Method: http.MethodGet,
		URI:    "app-*,infra-*,audit-*/_settings/index.number_of_replicas",
	}

	ec.fnSendEsRequest(ec.cluster, ec.namespace, payload, ec.k8sClient)

	return payload.ResponseBody, payload.Error
}

func (ec *esClient) GetLowestReplicaValue() (int32, error) {
	lowestReplica := int32(math.MaxInt32)
	indexHealth, err := ec.GetIndexReplicaCounts()
	if err != nil {
		return lowestReplica, err
	}

	// get list of indices and call updateIndexReplicas for each one
	for index, health := range indexHealth {
		if healthMap, ok := health.(map[string]interface{}); ok {
			if numberOfReplicas := parseString("settings.index.number_of_replicas", healthMap); numberOfReplicas != "" {
				currentReplicas, err := strconv.ParseInt(numberOfReplicas, 10, 32)
				if err != nil {
					return lowestReplica, err
				}

				if int32(currentReplicas) < lowestReplica {
					lowestReplica = int32(currentReplicas)
				}
			}
		} else {
			return lowestReplica, ec.errorCtx().New("unable to evaluate the number of replicas for index",
				"index", index,
				"health", health,
			)
		}
	}

	return lowestReplica, nil
}

func (ec *esClient) updateIndexReplicas(index string, replicaCount int32) (bool, error) {
	payload := &EsRequest{
		Method:      http.MethodPut,
		URI:         fmt.Sprintf("%s/_settings", index),
		RequestBody: fmt.Sprintf("{%q:\"%d\"}}", "index.number_of_replicas", replicaCount),
	}

	ec.fnSendEsRequest(ec.cluster, ec.namespace, payload, ec.k8sClient)

	acknowledged := false
	if acknowledgedBool, ok := payload.ResponseBody["acknowledged"].(bool); ok {
		acknowledged = acknowledgedBool
	}
	return payload.StatusCode == 200 && acknowledged, payload.Error
}
