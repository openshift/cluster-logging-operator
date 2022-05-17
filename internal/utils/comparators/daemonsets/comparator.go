package daemonsets

import (
	"reflect"

	"github.com/ViaQ/logerr/v2/log"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	apps "k8s.io/api/apps/v1"
)

//AreSame compares daemonset for equality and return true equal otherwise false
func AreSame(current *apps.DaemonSet, desired *apps.DaemonSet) bool {
	logger := log.NewLogger("")
	if !utils.AreMapsSame(current.Spec.Template.Spec.NodeSelector, desired.Spec.Template.Spec.NodeSelector) {
		logger.V(3).Info("DaemonSet nodeSelector change", "DaemonSetName", current.Name)
		return false
	}

	if !utils.AreTolerationsSame(current.Spec.Template.Spec.Tolerations, desired.Spec.Template.Spec.Tolerations) {
		logger.V(3).Info("DaemonSet tolerations change", "DaemonSetName", current.Name)
		return false
	}

	if !utils.PodVolumeEquivalent(current.Spec.Template.Spec.Volumes, desired.Spec.Template.Spec.Volumes) {
		logger.V(3).Info("DaemonSet volumes change", "DaemonSetName", current.Name)
		return false
	}

	if len(current.Spec.Template.Spec.Containers) != len(desired.Spec.Template.Spec.Containers) {
		logger.V(3).Info("DaemonSet number of containers changed", "DaemonSetName", current.Name)
		return false
	}

	if isDaemonsetImageDifference(current, desired) {
		logger.V(3).Info("DaemonSet image change", "DaemonSetName", current.Name)
		return false
	}

	if utils.AreResourcesDifferent(current, desired) {
		logger.V(3).Info("DaemonSet resource(s) change", "DaemonSetName", current.Name)
		return false
	}

	if !utils.EnvValueEqual(current.Spec.Template.Spec.Containers[0].Env, desired.Spec.Template.Spec.Containers[0].Env) {
		logger.V(3).Info("collector container EnvVar change found, updating ", "DaemonSetName", current.Name)
		logger.V(3).Info("collector envvars -", "current", current.Spec.Template.Spec.Containers[0].Env, "desired", desired.Spec.Template.Spec.Containers[0].Env)
		current.Spec.Template.Spec.Containers[0].Env = desired.Spec.Template.Spec.Containers[0].Env
		return false
	}

	if !reflect.DeepEqual(current.Spec.Template.Spec.Containers[0].VolumeMounts, desired.Spec.Template.Spec.Containers[0].VolumeMounts) {
		logger.V(3).Info("Daemonset %q container volumemounts change", "DaemonSetName", current.Name)
		return false
	}

	if len(current.Spec.Template.Spec.InitContainers) != len(desired.Spec.Template.Spec.InitContainers) {
		logger.V(3).Info("DaemonSet number of init containers changed", "DaemonSetName", current.Name)
		return false
	}
	for i, container := range current.Spec.Template.Spec.InitContainers {
		if container.Image != desired.Spec.Template.Spec.InitContainers[i].Image {
			logger.V(3).Info("Daemonset init container image is different from desired", "DaemonSetName", current.Name, "CurrentContainerName", container.Name, "DesiredContainerName", desired.Spec.Template.Spec.InitContainers[i].Name)
			return false
		}
	}

	return true
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
