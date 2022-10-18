package kibana

import (
	"github.com/openshift/cluster-logging-operator/internal/utils"
	es "github.com/openshift/elasticsearch-operator/apis/logging/v1"
	"reflect"
)

// AreSame Checks kibanaSpec for equality and returns true/false and the reason of the difference
func AreSame(current es.Kibana, desired es.Kibana) (bool, string) {
	if current.Spec.ManagementState != desired.Spec.ManagementState {
		return false, "spec.managementState"
	}

	if current.Spec.Replicas != desired.Spec.Replicas {
		return false, "spec.replicas"
	}

	if !utils.AreMapsSame(current.Spec.NodeSelector, desired.Spec.NodeSelector) {
		return false, "spec.nodeSelector"
	}

	if !utils.AreTolerationsSame(current.Spec.Tolerations, desired.Spec.Tolerations) {
		return false, "spec.tolerations"
	}
	if !reflect.DeepEqual(current.Spec.Resources, desired.Spec.Resources) {
		return false, "spec.resources"
	}

	if !reflect.DeepEqual(current.Spec.ProxySpec, desired.Spec.ProxySpec) {
		return false, "spec.proxySpec"
	}

	return true, ""
}

// CompareStatus of the KibanaStatus to see if is different
func CompareStatus(lhs, rhs []es.KibanaStatus) bool {
	// there should only ever be a single kibana status object
	if len(lhs) != len(rhs) {
		return false
	}

	if len(lhs) > 0 {
		for index := range lhs {
			if lhs[index].Deployment != rhs[index].Deployment {
				return false
			}

			if lhs[index].Replicas != rhs[index].Replicas {
				return false
			}

			if len(lhs[index].ReplicaSets) != len(rhs[index].ReplicaSets) {
				return false
			}

			if len(lhs[index].ReplicaSets) > 0 {
				if !reflect.DeepEqual(lhs[index].ReplicaSets, rhs[index].ReplicaSets) {
					return false
				}
			}

			if len(lhs[index].Pods) != len(rhs[index].Pods) {
				return false
			}

			if len(lhs[index].Pods) > 0 {
				if !reflect.DeepEqual(lhs[index].Pods, rhs[index].Pods) {
					return false
				}
			}

			if len(lhs[index].Conditions) != len(rhs[index].Conditions) {
				return false
			}

			if len(lhs[index].Conditions) > 0 {
				if !reflect.DeepEqual(lhs[index].Conditions, rhs[index].Conditions) {
					return false
				}
			}
		}
	}

	return true
}
