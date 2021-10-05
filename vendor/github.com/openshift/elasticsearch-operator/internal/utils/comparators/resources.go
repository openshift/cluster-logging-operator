package comparators

import (
	v1 "k8s.io/api/core/v1"
)

func AreResourceRequementsSame(lhs, rhs v1.ResourceRequirements) bool {
	if rhs.Limits.Cpu().Cmp(*lhs.Limits.Cpu()) != 0 {
		return false
	}
	// Check memory limits
	if rhs.Limits.Memory().Cmp(*lhs.Limits.Memory()) != 0 {
		return false
	}
	// Check CPU requests
	if rhs.Requests.Cpu().Cmp(*lhs.Requests.Cpu()) != 0 {
		return false
	}
	// Check memory requests
	if rhs.Requests.Memory().Cmp(*lhs.Requests.Memory()) != 0 {
		return false
	}

	return true
}
