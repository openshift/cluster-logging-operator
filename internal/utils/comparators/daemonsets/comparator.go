package daemonsets

import (
	"reflect"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	apps "k8s.io/api/apps/v1"
)

// AreSame compares daemonset for equality and return true equal otherwise false
func AreSame(current *apps.DaemonSet, desired *apps.DaemonSet) (bool, string) {
	if !utils.AreMapsSame(current.Spec.Template.Spec.NodeSelector, desired.Spec.Template.Spec.NodeSelector) {
		log.V(3).Info("DaemonSet nodeSelector change", "DaemonSetName", current.Name)
		return false, "nodeSelector"
	}

	if !utils.AreTolerationsSame(current.Spec.Template.Spec.Tolerations, desired.Spec.Template.Spec.Tolerations) {
		log.V(3).Info("DaemonSet tolerations change", "DaemonSetName", current.Name)
		return false, "tolerations"
	}

	if !utils.PodVolumeEquivalent(current.Spec.Template.Spec.Volumes, desired.Spec.Template.Spec.Volumes) {
		log.V(3).Info("DaemonSet volumes change", "DaemonSetName", current.Name)
		return false, "volumes"
	}

	if len(current.Spec.Template.Spec.Containers) != len(desired.Spec.Template.Spec.Containers) {
		log.V(3).Info("DaemonSet number of containers changed", "DaemonSetName", current.Name)
		return false, "numberOfContainers"
	}

	if isDaemonsetImageDifference(current, desired) {
		log.V(3).Info("DaemonSet image change", "DaemonSetName", current.Name)
		return false, "image"
	}

	if utils.AreResourcesDifferent(current, desired) {
		log.V(3).Info("DaemonSet resource(s) change", "DaemonSetName", current.Name)
		return false, "resources"
	}

	if !utils.EnvValueEqual(current.Spec.Template.Spec.Containers[0].Env, desired.Spec.Template.Spec.Containers[0].Env) {
		log.V(3).Info("collector container EnvVar change found, updating ", "DaemonSetName", current.Name)
		log.V(3).Info("collector envvars -", "current", current.Spec.Template.Spec.Containers[0].Env, "desired", desired.Spec.Template.Spec.Containers[0].Env)
		current.Spec.Template.Spec.Containers[0].Env = desired.Spec.Template.Spec.Containers[0].Env
		return false, "environment"
	}

	if !reflect.DeepEqual(current.Spec.Template.Spec.Containers[0].VolumeMounts, desired.Spec.Template.Spec.Containers[0].VolumeMounts) {
		log.V(3).Info("Daemonset %q container volumeMounts change", "DaemonSetName", current.Name)
		return false, "volumeMounts"
	}

	if len(current.Spec.Template.Spec.InitContainers) != len(desired.Spec.Template.Spec.InitContainers) {
		log.V(3).Info("DaemonSet number of init containers changed", "DaemonSetName", current.Name)
		return false, "initContainers"
	}
	for i, container := range current.Spec.Template.Spec.InitContainers {
		if container.Image != desired.Spec.Template.Spec.InitContainers[i].Image {
			log.V(3).Info("Daemonset init container image is different from desired", "DaemonSetName", current.Name, "CurrentContainerName", container.Name, "DesiredContainerName", desired.Spec.Template.Spec.InitContainers[i].Name)
			return false, "initContainerImage"
		}
	}

	// Check ownership
	if !reflect.DeepEqual(current.GetOwnerReferences(), desired.GetOwnerReferences()) {
		log.V(3).Info("Daemonset ownership change", "current name", current.GetOwnerReferences()[0].Name)
		return false, "ownerReference"
	}

	return true, ""
}

func isDaemonsetImageDifference(current *apps.DaemonSet, desired *apps.DaemonSet) bool {

	for _, curr := range current.Spec.Template.Spec.Containers {
		for _, des := range desired.Spec.Template.Spec.Containers {
			// Only compare the images of containers with the same name
			if curr.Name == des.Name {
				if curr.Image != des.Image {
					return true
				}
			}
		}
	}

	return false
}
