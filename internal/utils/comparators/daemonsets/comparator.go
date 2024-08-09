package daemonsets

import (
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"reflect"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators/pod"
	apps "k8s.io/api/apps/v1"
)

// AreSame compares daemonset for equality and return true equal otherwise false
func AreSame(current *apps.DaemonSet, desired *apps.DaemonSet) (bool, string) {

	// Check pod spec
	if same, resource := pod.AreSame(&current.Spec.Template.Spec, &desired.Spec.Template.Spec, current.Name); !same {
		log.V(3).Info("Daemonset pod spec change", "name", current.Name)
		return false, resource
	}

	// Check labels
	if !reflect.DeepEqual(current.Labels, desired.Labels) {
		log.V(3).Info("Daemonset labels change", "name", current.Name)
		return false, "labels"
	}

	// Check ownership
	if !reflect.DeepEqual(current.GetOwnerReferences(), desired.GetOwnerReferences()) {
		log.V(3).Info("Daemonset ownership change", "name", current.Name)
		return false, "ownerReference"
	}

	currentHash := current.Spec.Template.Annotations[constants.AnnotationSecretHash]
	desiredHash := desired.Spec.Template.Annotations[constants.AnnotationSecretHash]
	if currentHash != desiredHash {
		return false, "secretHash"
	}

	return true, ""
}
