package utils

import (
	log "github.com/ViaQ/logerr/v2/log/static"
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

func AreResourcesDifferent(current, desired *v1.PodSpec) bool {

	if current == nil || desired == nil {
		return false
	}

	currentContainers := current.Containers
	desiredContainers := desired.Containers

	changed := false

	for _, curr := range currentContainers {
		for _, des := range desiredContainers {
			// Only compare the images of containers with the same name
			if curr.Name == des.Name {
				if !AreResourcesSame(&curr.Resources, &des.Resources) {
					changed = true
				}
			}
		}
	}

	return changed
}
