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
	for index, clusterCondition := range clusterRequest.cluster.Status.Conditions {
		if clusterCondition.Type == condition.Type {
			found = true
			if condition.Status == v1.ConditionFalse {
				clusterRequest.cluster.Status.Conditions = removeCondition(clusterRequest.cluster.Status.Conditions, index)
				updated = true
			} else {
				if isConditionDifferent(clusterCondition, condition) {
					condition.LastTransitionTime = metav1.Now()
					clusterRequest.cluster.Status.Conditions[index] = condition
					updated = true
				}
			}
			break
		}
	}

	if !found {
		if condition.Status == v1.ConditionTrue {
			condition.LastTransitionTime = metav1.Now()
			clusterRequest.cluster.Status.Conditions = append(clusterRequest.cluster.Status.Conditions, condition)
			updated = true
		}
	}

	if updated {
		return clusterRequest.UpdateStatus(clusterRequest.cluster)
	}

	return nil
}

func (clusterRequest *ClusterLoggingRequest) GetCondition(conditionType logging.ClusterConditionType) logging.ClusterCondition {

	for _, clusterCondition := range clusterRequest.cluster.Status.Conditions {
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
	return lhs.Type != rhs.Type ||
		lhs.Status != rhs.Status ||
		lhs.Reason != rhs.Reason ||
		lhs.Message != rhs.Message
}
