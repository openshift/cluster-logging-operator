package daemonsets

import (
	"reflect"

	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	apps "k8s.io/api/apps/v1"
)

//AreSame compares daemonset for equality and return true equal otherwise false
func AreSame(current *apps.DaemonSet, desired *apps.DaemonSet) bool {
	logger.Tracef("Comparing DaemonSets current %v to desired %v", current, desired)

	if !utils.AreMapsSame(current.Spec.Template.Spec.NodeSelector, desired.Spec.Template.Spec.NodeSelector) {
		logger.Debugf("DaemonSet %q nodeSelector change", current.Name)
		return false
	}

	if !utils.AreTolerationsSame(current.Spec.Template.Spec.Tolerations, desired.Spec.Template.Spec.Tolerations) {
		logger.Debugf("DaemonSet %q tolerations change", current.Name)
		return false
	}

	if !utils.PodVolumeEquivalent(current.Spec.Template.Spec.Volumes, desired.Spec.Template.Spec.Volumes) {
		logger.Debugf("DaemonSet %q volumes change", current.Name)
		return false
	}

	if len(current.Spec.Template.Spec.Containers) != len(desired.Spec.Template.Spec.Containers) {
		logger.Debugf("DaemonSet %q number of containers changed", current.Name)
		return false
	}

	if isDaemonsetImageDifference(current, desired) {
		logger.Debugf("DaemonSet %q image change", current.Name)
		return false
	}

	if utils.AreResourcesDifferent(current, desired) {
		logger.Debugf("DaemonSet %q resource(s) change", current.Name)
		return false
	}

	if !utils.EnvValueEqual(current.Spec.Template.Spec.Containers[0].Env, desired.Spec.Template.Spec.Containers[0].Env) {
		logger.Infof("Collector container EnvVar change found, updating %q", current.Name)
		logger.Debugf("Collector envvars - current: %v, desired: %v", current.Spec.Template.Spec.Containers[0].Env, desired.Spec.Template.Spec.Containers[0].Env)
		current.Spec.Template.Spec.Containers[0].Env = desired.Spec.Template.Spec.Containers[0].Env
		return false
	}

	if !reflect.DeepEqual(current.Spec.Template.Spec.Containers[0].VolumeMounts, desired.Spec.Template.Spec.Containers[0].VolumeMounts) {
		logger.Debugf("Daemonset %q container volumemounts change", current.Name)
		return false
	}

	if len(current.Spec.Template.Spec.InitContainers) != len(desired.Spec.Template.Spec.InitContainers) {
		logger.Debugf("DaemonSet %q number of init containers changed", current.Name)
		return false
	}
	for i, container := range current.Spec.Template.Spec.InitContainers {
		if container.Image != desired.Spec.Template.Spec.InitContainers[i].Image {
			logger.Debugf("Daemonset %q initcontainer %q image is different from desired %q", current.Name, container.Name, desired.Spec.Template.Spec.InitContainers[i].Name)
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
