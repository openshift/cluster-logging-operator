package utils

import (
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"

	v1 "k8s.io/api/core/v1"
)

var (
	limitMemory   = resource.MustParse("120Gi")
	requestMemory = resource.MustParse("100Gi")
	requestCPU    = resource.MustParse("500m")
)

func TestAreResourcesEmptyWhenUpdating(t *testing.T) {

	current := v1.ResourceRequirements{
		Limits: v1.ResourceList{v1.ResourceMemory: limitMemory},
		Requests: v1.ResourceList{
			v1.ResourceMemory: requestMemory,
			v1.ResourceCPU:    requestCPU,
		},
	}

	desired := v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceMemory: requestMemory,
			v1.ResourceCPU:    requestCPU,
		},
	}

	different, result := CompareResources(current, desired)

	if !different {
		t.Error("Expected resourceRequirements to evaluate as different")
	}

	if !reflect.DeepEqual(result.Limits, desired.Limits) {
		t.Error("Expected limits to both be empty")
	}
}
