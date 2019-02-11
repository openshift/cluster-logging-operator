package k8shandler

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func newResourceRequirements(limitMem string, limitCPU string, requestMem string, requestCPU string) *v1.ResourceRequirements {
	resources := v1.ResourceRequirements{
		Limits:   v1.ResourceList{},
		Requests: v1.ResourceList{},
	}
	if limitMem != "" {
		resources.Limits[v1.ResourceMemory] = resource.MustParse(limitMem)
	}
	if limitCPU != "" {
		resources.Limits[v1.ResourceCPU] = resource.MustParse(limitCPU)
	}
	if requestMem != "" {
		resources.Requests[v1.ResourceMemory] = resource.MustParse(requestMem)
	}
	if requestCPU != "" {
		resources.Requests[v1.ResourceCPU] = resource.MustParse(requestCPU)
	}
	return &resources
}
