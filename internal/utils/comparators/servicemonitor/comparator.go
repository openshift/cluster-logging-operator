package servicemonitor

import (
	"reflect"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

// AreSame compares for equality and return true equal otherwise false
func AreSame(current *monitoringv1.ServiceMonitor, desired *monitoringv1.ServiceMonitor) bool {
	log.V(3).Info("Comparing Services current to desired", "current", current, "desired", desired)

	if !utils.AreMapsSame(current.ObjectMeta.Annotations, desired.ObjectMeta.Annotations) {
		log.V(3).Info("ServiceMonitor  annotation change", "current name", current.Name)
		return false
	}

	if !utils.AreMapsSame(current.ObjectMeta.Labels, desired.ObjectMeta.Labels) {
		log.V(3).Info("ServiceMonitor label change", "current name", current.Name)
		return false
	}

	if !utils.AreMapsSame(current.Spec.Selector.MatchLabels, desired.Spec.Selector.MatchLabels) {
		log.V(3).Info("ServiceMonitor Selector labels change", "current name", current.Name)
		return false
	}

	if current.Spec.JobLabel != desired.Spec.JobLabel {
		log.V(3).Info("Service Selector JobLabel change", "current name", current.Name)
		return false
	}

	if len(current.Spec.Selector.MatchExpressions) != len(desired.Spec.Selector.MatchExpressions) {
		log.V(3).Info("Service Selector MatchExpressions change", "current name", current.Name)
		return false
	}

	for i, matchExpression := range current.Spec.Selector.MatchExpressions {
		m := desired.Spec.Selector.MatchExpressions[i]
		if !reflect.DeepEqual(matchExpression, m) {
			log.V(3).Info("Service Selector MatchExpressions change", "current name", current.Name)
			return false
		}
	}

	if len(current.Spec.PodTargetLabels) != len(desired.Spec.PodTargetLabels) {
		log.V(3).Info("Service Selector PodTargetLabels change", "current name", current.Name)
		return false
	}

	for i, targetLabel := range current.Spec.PodTargetLabels {
		t := desired.Spec.PodTargetLabels[i]
		if targetLabel != t {
			log.V(3).Info("Service Selector PodTargetLabels change", "current name", current.Name)
			return false
		}
	}

	if len(current.Spec.Endpoints) != len(desired.Spec.Endpoints) {
		log.V(3).Info("Service Selector Endpoints change", "current name", current.Name)
		return false
	}

	for i, endpoint := range current.Spec.Endpoints {
		e := desired.Spec.Endpoints[i]
		if !reflect.DeepEqual(endpoint, e) {
			log.V(3).Info("Service Selector Endpoints change", "current name", current.Name)
			return false
		}
	}

	// Check ownership
	if !reflect.DeepEqual(current.GetOwnerReferences(), desired.GetOwnerReferences()) {
		log.V(3).Info("ownership change", "current name", current.GetOwnerReferences()[0].Name)
		return false
	}

	return true
}
