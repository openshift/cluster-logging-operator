package services

import (
	"reflect"

	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	v1 "k8s.io/api/core/v1"
)

//AreSame compares for equality and return true equal otherwise false
func AreSame(current *v1.Service, desired *v1.Service) bool {
	logger.Tracef("Comparing Services current %v to desired %v", current, desired)

	if !utils.AreMapsSame(current.ObjectMeta.Labels, desired.ObjectMeta.Labels) {
		logger.Debugf("Service %q label change", current.Name)
		return false
	}
	if !utils.AreMapsSame(current.Spec.Selector, desired.Spec.Selector) {
		logger.Debugf("Service %q Selector change", current.Name)
		return false
	}
	if len(current.Spec.Ports) != len(desired.Spec.Ports) {
		return false
	}
	for i, port := range current.Spec.Ports {
		dPort := desired.Spec.Ports[i]
		if !reflect.DeepEqual(port, dPort) {
			return false
		}
	}

	return true
}
