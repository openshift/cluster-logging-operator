package utils

import (
	"errors"
	"reflect"

	"github.com/ViaQ/logerr/v2/log"
	apps "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
)

func CompareResources(current, desired v1.ResourceRequirements) (bool, v1.ResourceRequirements) {

	changed := false

	// Check CPU limits
	if desired.Limits.Cpu().Cmp(*current.Limits.Cpu()) != 0 {
		changed = true
	}
	// Check memory limits
	if desired.Limits.Memory().Cmp(*current.Limits.Memory()) != 0 {
		changed = true
	}
	// Check CPU requests
	if desired.Requests.Cpu().Cmp(*current.Requests.Cpu()) != 0 {
		changed = true
	}
	// Check memory requests
	if desired.Requests.Memory().Cmp(*current.Requests.Memory()) != 0 {
		changed = true
	}

	return changed, desired
}

func AreResourcesDifferent(current, desired interface{}) bool {

	var currentContainers []v1.Container
	var desiredContainers []v1.Container

	currentType := reflect.TypeOf(current)
	desiredType := reflect.TypeOf(desired)

	logger := log.NewLogger("")

	if currentType != desiredType {
		logger.Error(errors.New("Attempting to compare resources for different types"), "",
			"current", currentType, "desired", desiredType)
		return false
	}

	switch currentType {
	case reflect.TypeOf(&apps.Deployment{}):
		currentContainers = current.(*apps.Deployment).Spec.Template.Spec.Containers
		desiredContainers = desired.(*apps.Deployment).Spec.Template.Spec.Containers

	case reflect.TypeOf(&apps.DaemonSet{}):
		currentContainers = current.(*apps.DaemonSet).Spec.Template.Spec.Containers
		desiredContainers = desired.(*apps.DaemonSet).Spec.Template.Spec.Containers

	case reflect.TypeOf(&batch.CronJob{}):
		currentContainers = current.(*batch.CronJob).Spec.JobTemplate.Spec.Template.Spec.Containers
		desiredContainers = desired.(*batch.CronJob).Spec.JobTemplate.Spec.Template.Spec.Containers

	default:
		logger.Info("Attempting to check resources for unmatched type", "current", currentType)
		return false
	}

	containers := currentContainers
	changed := false

	for index, curr := range currentContainers {
		for _, des := range desiredContainers {
			// Only compare the images of containers with the same name
			if curr.Name == des.Name {
				if different, updated := CompareResources(curr.Resources, des.Resources); different {
					containers[index].Resources = updated
					changed = true
				}
			}
		}
	}

	return changed
}
