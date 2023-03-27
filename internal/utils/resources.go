package utils

import (
	"errors"
	"reflect"

	log "github.com/ViaQ/logerr/v2/log/static"
	apps "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
)

func AreResourcesSame(current, desired *v1.ResourceRequirements) bool {
	if current == nil && desired == nil {
		return true
	}
	if (current != nil && desired == nil) || (current == nil && desired != nil) {
		log.V(3).Info("Resources are not the same missing ResoureRequirement", "current", current, "desired", desired)
		return false
	}
	// Check CPU limits
	if desired.Limits.Cpu().Cmp(*current.Limits.Cpu()) != 0 {
		log.V(3).Info("Resources are not the same: limits.cpu", "current", current, "desired", desired)
		return false
	}
	// Check memory limits
	if desired.Limits.Memory().Cmp(*current.Limits.Memory()) != 0 {
		log.V(3).Info("Resources are not the same: limits.memory", "current", current, "desired", desired)
		return false
	}
	// Check CPU requests
	if desired.Requests.Cpu().Cmp(*current.Requests.Cpu()) != 0 {
		log.V(3).Info("Resources are not the same: requests.cpu", "current", current, "desired", desired)
		return false
	}
	// Check memory requests
	if desired.Requests.Memory().Cmp(*current.Requests.Memory()) != 0 {
		log.V(3).Info("Resources are not the same: requests.memory", "current", current, "desired", desired)
		return false
	}

	return true
}

func AreResourcesDifferent(current, desired interface{}) bool {

	var currentContainers []v1.Container
	var desiredContainers []v1.Container

	currentType := reflect.TypeOf(current)
	desiredType := reflect.TypeOf(desired)

	if currentType != desiredType {
		log.Error(errors.New("Attempting to compare resources for different types"), "",
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
		log.Info("Attempting to check resources for unmatched type", "current", currentType)
		return false
	}

	changed := false

	for _, curr := range currentContainers {
		for _, des := range desiredContainers {
			// Only compare the images of containers with the same name
			if curr.Name == des.Name {
				if AreResourcesSame(&curr.Resources, &des.Resources) {
					changed = true
				}
			}
		}
	}

	return changed
}
