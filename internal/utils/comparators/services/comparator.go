package services

import (
	"fmt"
	"reflect"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	v1 "k8s.io/api/core/v1"
)

// AreSame compares for equality and return true equal otherwise false
func AreSame(current *v1.Service, desired *v1.Service) (bool, string) {
	log.V(3).Info("Comparing Services current to desired", "current", current, "desired", desired)

	if !utils.AreMapsSame(current.ObjectMeta.Labels, desired.ObjectMeta.Labels) {
		log.V(3).Info("Service label change", "current name", current.Name)
		return false, "meta.labels"
	}
	if !utils.AreMapsSame(current.Spec.Selector, desired.Spec.Selector) {
		log.V(3).Info("Service Selector change", "current name", current.Name)
		return false, "spec.selector"
	}
	if len(current.Spec.Ports) != len(desired.Spec.Ports) {
		return false, "spec.ports"
	}
	for i, port := range current.Spec.Ports {
		dPort := desired.Spec.Ports[i]
		if !reflect.DeepEqual(port, dPort) {
			return false, fmt.Sprintf("spec.ports[%d]", i)
		}
	}

	// Check ownership
	if !reflect.DeepEqual(current.GetOwnerReferences(), desired.GetOwnerReferences()) {
		log.V(3).Info("Service Ownership change", "current name", current.GetOwnerReferences()[0].Name)
		return false, "spec.ownerReference"
	}

	return true, ""
}
