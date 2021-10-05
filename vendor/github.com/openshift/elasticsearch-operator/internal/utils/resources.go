package utils

import (
	"reflect"

	"github.com/ViaQ/logerr/log"

	apps "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func CompareResources(current, desired v1.ResourceRequirements) (bool, v1.ResourceRequirements) {
	changed := false
	desiredResources := *current.DeepCopy()
	if desiredResources.Limits == nil {
		desiredResources.Limits = map[v1.ResourceName]resource.Quantity{}
	}
	if desiredResources.Requests == nil {
		desiredResources.Requests = map[v1.ResourceName]resource.Quantity{}
	}

	if desired.Limits.Cpu().Cmp(*current.Limits.Cpu()) != 0 {
		desiredResources.Limits[v1.ResourceCPU] = *desired.Limits.Cpu()
		changed = true
	}
	// Check memory limits
	if desired.Limits.Memory().Cmp(*current.Limits.Memory()) != 0 {
		desiredResources.Limits[v1.ResourceMemory] = *desired.Limits.Memory()
		changed = true
	}
	// Check CPU requests
	if desired.Requests.Cpu().Cmp(*current.Requests.Cpu()) != 0 {
		desiredResources.Requests[v1.ResourceCPU] = *desired.Requests.Cpu()
		changed = true
	}
	// Check memory requests
	if desired.Requests.Memory().Cmp(*current.Requests.Memory()) != 0 {
		desiredResources.Requests[v1.ResourceMemory] = *desired.Requests.Memory()
		changed = true
	}

	return changed, desiredResources
}

func AreResourcesDifferent(current, desired interface{}) bool {
	var currentContainers []v1.Container
	var desiredContainers []v1.Container

	currentType := reflect.TypeOf(current)
	desiredType := reflect.TypeOf(desired)

	if currentType != desiredType {
		log.Info("Attempting to compare resources for different types", "current", currentType, "desired", desiredType)
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
		log.Info("Attempting to check resources for unmatched type", "type", currentType)
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
