package k8shandler

// the purpose of this file is give an easy means to update/add clusterconditions

import (
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (clusterRequest *ClusterLoggingRequest) UpdateCondition(conditionType logging.ClusterConditionType, message, reason string, status v1.ConditionStatus) error {

	condition := clusterRequest.GetCondition(conditionType)

	condition.Message = message
	condition.Reason = reason
	condition.Status = status

	found := false
	updated := false
	for index, clusterCondition := range clusterRequest.Cluster.Status.Conditions {
		if clusterCondition.Type == condition.Type {
			found = true
			if condition.Status == v1.ConditionFalse {
				clusterRequest.Cluster.Status.Conditions = removeCondition(clusterRequest.Cluster.Status.Conditions, index)
				updated = true
			} else {
				if isConditionDifferent(clusterCondition, condition) {
					condition.LastTransitionTime = metav1.Now()
					clusterRequest.Cluster.Status.Conditions[index] = condition
					updated = true
				}
			}
			break
		}
	}

	if !found {
		if condition.Status == v1.ConditionTrue {
			condition.LastTransitionTime = metav1.Now()
			clusterRequest.Cluster.Status.Conditions = append(clusterRequest.Cluster.Status.Conditions, condition)
			updated = true
		}
	}

	if updated {
		return clusterRequest.UpdateStatus(clusterRequest.Cluster)
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) GetCondition(conditionType logging.ClusterConditionType) logging.ClusterCondition {

	for _, clusterCondition := range clusterRequest.Cluster.Status.Conditions {
		if clusterCondition.Type == conditionType {
			return clusterCondition
		}
	}

	return logging.ClusterCondition{
		Type: conditionType,
	}
}

// we use this to remove a condition from the list of clusterConditions
// typically when the status == v1.ConditionFalse
func removeCondition(conditions []logging.ClusterCondition, index int) []logging.ClusterCondition {
	return append(conditions[:index], conditions[index+1:]...)
}

func isConditionDifferent(lhs, rhs logging.ClusterCondition) bool {

	if lhs.Type != rhs.Type {
		return true
	}

	if lhs.Status != rhs.Status {
		return true
	}

	if lhs.Reason != rhs.Reason {
		return true
	}

	if lhs.Message != rhs.Message {
		return true
	}

	return false
}
