package comparators

import (
	"reflect"

	corev1 "k8s.io/api/core/v1"
)

// check that all of rhs (desired) are contained within lhs (current)
func ContainsSameVolumeMounts(lhs, rhs []corev1.VolumeMount) bool {
	for _, rVolumeMount := range rhs {
		found := false

		for _, lVolumeMount := range lhs {
			if lVolumeMount.Name == rVolumeMount.Name {
				found = true

				if !reflect.DeepEqual(lVolumeMount, rVolumeMount) {
					return false
				}
			}
		}

		if !found {
			return false
		}
	}

	return true
}
