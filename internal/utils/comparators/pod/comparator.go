package pod

import (
	"reflect"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
)

// AreSame compares pods for equality and return true equal otherwise false
func AreSame(current *corev1.PodSpec, desired *corev1.PodSpec, name string) (bool, string) {

	if !utils.AreMapsSame(current.NodeSelector, desired.NodeSelector) {
		log.V(3).Info("nodeSelector change", "name", name)
		return false, "nodeSelector"
	}

	if !utils.AreTolerationsSame(current.Tolerations, desired.Tolerations) {
		log.V(3).Info("tolerations change", "name", name)
		return false, "tolerations"
	}

	if !utils.PodVolumeEquivalent(current.Volumes, current.Volumes) {
		log.V(3).Info("volumes changed", "name", name)
		return false, "volumes"
	}

	if len(current.Containers) != len(desired.Containers) {
		log.V(3).Info("number of containers changed", "name", name)
		return false, "numberOfContainers"
	}

	if utils.AreResourcesDifferent(current, desired) {
		log.V(3).Info("resource(s) change", "name", name)
		return false, "resources"
	}

	for i, container := range current.Containers {
		if !utils.EnvValueEqual(container.Env, desired.Containers[i].Env) {
			log.V(3).Info("container EnvVar change found, updating ", "name", name)
			log.V(3).Info("envvars -", "current", container.Env, "desired", desired.Containers[i].Env)
			container.Env = desired.Containers[i].Env
			return false, "environment"
		}

		if !reflect.DeepEqual(container.VolumeMounts, desired.Containers[i].VolumeMounts) {
			log.V(3).Info("%q container volumeMounts change", "name", name)
			return false, "volumeMounts"
		}

		// Check pod image
		for _, des := range desired.Containers {
			// Only compare the images of containers with the same name
			if container.Name == des.Name {
				if container.Image != des.Image {
					return false, "podImage"
				}
			}
		}

	}

	if len(current.InitContainers) != len(desired.InitContainers) {
		log.V(3).Info("number of init containers changed", "name", name)
		return false, "initContainers"
	}
	for i, container := range current.InitContainers {
		if container.Image != desired.InitContainers[i].Image {
			log.V(3).Info("init container image is different from desired", "name", name, "CurrentContainerName", container.Name, "DesiredContainerName", desired.InitContainers[i].Name)
			return false, "initContainerImage"
		}
	}

	return true, ""
}
