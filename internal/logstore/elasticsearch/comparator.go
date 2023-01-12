package elasticsearch

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	elasticsearch "github.com/openshift/elasticsearch-operator/apis/logging/v1"
	"reflect"
)

func StatusAreSame(lhs, rhs []logging.ElasticsearchStatus) bool {
	// there should only ever be a single elasticsearch status object
	if len(lhs) != len(rhs) {
		return false
	}

	if len(lhs) > 0 {
		for index := range lhs {
			if lhs[index].ClusterName != rhs[index].ClusterName {
				return false
			}

			if lhs[index].NodeCount != rhs[index].NodeCount {
				return false
			}

			if lhs[index].ClusterHealth != rhs[index].ClusterHealth {
				return false
			}

			if lhs[index].Cluster != rhs[index].Cluster {
				return false
			}

			if lhs[index].ShardAllocationEnabled != rhs[index].ShardAllocationEnabled {
				return false
			}

			if len(lhs[index].Pods) != len(rhs[index].Pods) {
				return false
			}

			if len(lhs[index].Pods) > 0 {
				if !reflect.DeepEqual(lhs[index].Pods, rhs[index].Pods) {
					return false
				}
			}

			if len(lhs[index].ClusterConditions) != len(rhs[index].ClusterConditions) {
				return false
			}

			if len(lhs[index].ClusterConditions) > 0 {
				if !reflect.DeepEqual(lhs[index].ClusterConditions, rhs[index].ClusterConditions) {
					return false
				}
			}

			if len(lhs[index].NodeConditions) != len(rhs[index].NodeConditions) {
				return false
			}

			if len(lhs[index].NodeConditions) > 0 {
				if !reflect.DeepEqual(lhs[index].NodeConditions, rhs[index].NodeConditions) {
					return false
				}
			}
		}
	}

	return true
}

func IsElasticsearchCRDifferent(current *elasticsearch.Elasticsearch, desired *elasticsearch.Elasticsearch) (*elasticsearch.Elasticsearch, bool) {

	different := false

	if !utils.AreMapsSame(current.Spec.Spec.NodeSelector, desired.Spec.Spec.NodeSelector) {
		log.Info("Elasticsearch nodeSelector change found, updating", "currentName", current.Name)
		current.Spec.Spec.NodeSelector = desired.Spec.Spec.NodeSelector
		different = true
	}

	if !utils.AreTolerationsSame(current.Spec.Spec.Tolerations, desired.Spec.Spec.Tolerations) {
		log.Info("Elasticsearch tolerations change found, updating", "currentName", current.Name)
		current.Spec.Spec.Tolerations = desired.Spec.Spec.Tolerations
		different = true
	}

	if current.Spec.Spec.Image != desired.Spec.Spec.Image {
		log.Info("Elasticsearch image change found, updating", "currentName", current.Name)
		current.Spec.Spec.Image = desired.Spec.Spec.Image
		different = true
	}

	if current.Spec.RedundancyPolicy != desired.Spec.RedundancyPolicy {
		log.Info("Elasticsearch redundancy policy change found, updating", "currentName", current.Name)
		current.Spec.RedundancyPolicy = desired.Spec.RedundancyPolicy
		different = true
	}

	if !reflect.DeepEqual(current.ObjectMeta.Annotations, desired.ObjectMeta.Annotations) {
		log.Info("Elasticsearch resources change found in Annotations, updating", "currentName", current.Name)
		current.Annotations = desired.Annotations
		different = true
	}

	if !reflect.DeepEqual(current.Spec.Spec.Resources, desired.Spec.Spec.Resources) {
		log.Info("Elasticsearch resources change found, updating", "currentName", current.Name)
		current.Spec.Spec.Resources = desired.Spec.Spec.Resources
		different = true
	}

	if !reflect.DeepEqual(current.Spec.Spec.ProxyResources, desired.Spec.Spec.ProxyResources) {
		log.Info("Elasticsearch Proxy resources change found, updating", "currentName", current.Name)
		current.Spec.Spec.ProxyResources = desired.Spec.Spec.ProxyResources
		different = true
	}

	if nodes, ok := areNodesDifferent(current.Spec.Nodes, desired.Spec.Nodes); ok {
		log.Info("Elasticsearch node configuration change found, updating", "currentName", current.Name)
		current.Spec.Nodes = nodes
		different = true
	}

	if !reflect.DeepEqual(current.Spec.IndexManagement, desired.Spec.IndexManagement) {
		log.Info("Elasticsearch IndexManagement change found, updating", "currentName", current.Name)
		current.Spec.IndexManagement = desired.Spec.IndexManagement
		different = true
	}

	return current, different
}

func areNodesDifferent(current, desired []elasticsearch.ElasticsearchNode) ([]elasticsearch.ElasticsearchNode, bool) {

	different := false

	// nodes were removed
	if len(current) == 0 {
		return desired, true
	}

	foundRoleMatch := false
	for nodeIndex := 0; nodeIndex < len(desired); nodeIndex++ {
		for _, node := range current {
			if areNodeRolesSame(node, desired[nodeIndex]) {
				updatedNode, isDifferent := isNodeDifferent(node, desired[nodeIndex])
				if isDifferent {
					desired[nodeIndex] = updatedNode
					different = true
				} else if desired[nodeIndex].GenUUID == nil {
					// ensure that we are setting the GenUUID if it existed
					desired[nodeIndex].GenUUID = updatedNode.GenUUID
				}
				foundRoleMatch = true
			}
		}
	}

	// if we didn't find a role match, then that means changes were made
	if !foundRoleMatch {
		different = true
	}

	// we don't use this to shortcut because the above loop will help to preserve
	// any generated UUIDs
	if len(current) != len(desired) {
		return desired, true
	}

	return desired, different
}

func areNodeRolesSame(lhs, rhs elasticsearch.ElasticsearchNode) bool {

	if len(lhs.Roles) != len(rhs.Roles) {
		return false
	}

	lhsClient := false
	lhsData := false
	lhsMaster := false

	rhsClient := false
	rhsData := false
	rhsMaster := false

	for _, role := range lhs.Roles {
		if role == elasticsearch.ElasticsearchRoleClient {
			lhsClient = true
		}

		if role == elasticsearch.ElasticsearchRoleData {
			lhsData = true
		}

		if role == elasticsearch.ElasticsearchRoleMaster {
			lhsMaster = true
		}
	}

	for _, role := range rhs.Roles {
		if role == elasticsearch.ElasticsearchRoleClient {
			rhsClient = true
		}

		if role == elasticsearch.ElasticsearchRoleData {
			rhsData = true
		}

		if role == elasticsearch.ElasticsearchRoleMaster {
			rhsMaster = true
		}
	}

	return (lhsClient == rhsClient) && (lhsData == rhsData) && (lhsMaster == rhsMaster)
}

func isNodeDifferent(current, desired elasticsearch.ElasticsearchNode) (elasticsearch.ElasticsearchNode, bool) {

	different := false

	// check the different components that we normally set instead of using reflect
	// ignore the GenUUID if we aren't setting it.
	if desired.GenUUID == nil {
		desired.GenUUID = current.GenUUID
	}

	if !reflect.DeepEqual(current, desired) {
		current = desired
		different = true
	}

	return current, different
}
