package deployments

import (
	"reflect"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/pod"
	apps "k8s.io/api/apps/v1"
)

// AreSame compares deployments for equality and return true equal otherwise false
func AreSame(current *apps.Deployment, desired *apps.Deployment) (bool, string) {

	// Check pod specs
	if same, resource := pod.AreSame(&current.Spec.Template.Spec, &desired.Spec.Template.Spec, current.Name); !same {
		log.V(3).Info("Deployment pod spec change", "name", current.Name)
		return false, resource
	}

	// Check labels
	if !reflect.DeepEqual(current.Labels, desired.Labels) {
		log.V(3).Info("Deployment labels change", "name", current.Name)
		return false, "labels"
	}

	// Check ownership
	if !reflect.DeepEqual(current.GetOwnerReferences(), desired.GetOwnerReferences()) {
		log.V(3).Info("Deployment ownership change", "name", current.Name)
		return false, "ownerReference"
	}

	return true, ""
}
