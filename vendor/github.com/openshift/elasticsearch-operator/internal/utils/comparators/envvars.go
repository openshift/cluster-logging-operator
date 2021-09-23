package comparators

import (
	"reflect"

	v1 "k8s.io/api/core/v1"
)

/**
EnvValueEqual - check if 2 EnvValues are equal or not
Notes:
- reflect.DeepEqual does not return expected results if the to-be-compared value is a pointer.
- needs to adjust with k8s.io/api/core/v#/types.go when the types are updated.
**/
func EnvValueEqual(lhs, rhs []v1.EnvVar) bool {
	if len(lhs) != len(rhs) {
		return false
	}

	for _, l := range lhs {
		found := false

		for _, r := range rhs {

			if l.Name != r.Name {
				continue
			}

			found = true
			if !EnvVarEqual(l, r) {
				return false
			}
		}

		if !found {
			return false
		}
	}

	return true
}

func EnvVarEqual(lhs, rhs v1.EnvVar) bool {
	if lhs.ValueFrom != nil {
		if rhs.ValueFrom == nil {
			return false
		}

		// compare ValueFrom here
		return EnvVarSourceEqual(*lhs.ValueFrom, *rhs.ValueFrom)

	} else {
		if rhs.ValueFrom != nil {
			return false
		}

		// compare Value here
		return lhs.Value == rhs.Value
	}
}

func EnvVarSourceEqual(lhs, rhs v1.EnvVarSource) bool {
	if lhs.FieldRef != nil && rhs.FieldRef != nil {
		return EnvFieldRefEqual(*lhs.FieldRef, *rhs.FieldRef)
	}

	if lhs.ResourceFieldRef != nil && rhs.ResourceFieldRef != nil {
		return EnvResourceFieldRefEqual(*lhs.ResourceFieldRef, *rhs.ResourceFieldRef)
	}

	if lhs.ConfigMapKeyRef != nil && rhs.ConfigMapKeyRef != nil {
		return reflect.DeepEqual(*lhs.ConfigMapKeyRef, *rhs.ConfigMapKeyRef)
	}

	if lhs.SecretKeyRef != nil && rhs.SecretKeyRef != nil {
		return reflect.DeepEqual(*lhs.SecretKeyRef, *rhs.SecretKeyRef)
	}

	return false
}

func EnvFieldRefEqual(lhs, rhs v1.ObjectFieldSelector) bool {
	// taken from https://godoc.org/k8s.io/api/core/v1#ObjectFieldSelector
	// this is the default value, so if omitted by us k8s will add this value in
	defaultAPIVersion := "v1"

	if lhs.APIVersion == "" {
		lhs.APIVersion = defaultAPIVersion
	}

	if rhs.APIVersion == "" {
		rhs.APIVersion = defaultAPIVersion
	}

	if lhs.APIVersion != rhs.APIVersion {
		return false
	}

	return lhs.FieldPath == rhs.FieldPath
}

func EnvResourceFieldRefEqual(lhs, rhs v1.ResourceFieldSelector) bool {
	// taken from https://godoc.org/k8s.io/api/core/v1#ResourceFieldSelector
	// divisor's default value is "1"
	if lhs.Divisor.Cmp(rhs.Divisor) != 0 {
		return false
	}

	if lhs.ContainerName != rhs.ContainerName {
		return false
	}

	return lhs.Resource == rhs.Resource
}
