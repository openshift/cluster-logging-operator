package utils

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EnsureCanUpdateOwnedResource returns nil when the object has not been persisted yet
// (create path) or when it is already owned by the expected owner.
// If the object exists and is not owned by the expected owner (foreign-owned or
// unowned), it returns an error and callers must not mutate the object.
//
// CLO-managed resources have a single controller ownerReference (the CLF), or no
// owner at all (e.g. dashboard ConfigMaps). Ownership is compared with HasSameOwner.
func EnsureCanUpdateOwnedResource(obj metav1.Object, desiredOwners ...metav1.OwnerReference) error {
	if obj.GetResourceVersion() == "" {
		return nil
	}
	if HasSameOwner(obj.GetOwnerReferences(), desiredOwners) {
		return nil
	}
	return ResourceOwnershipConflictError(obj)
}

// ResourceOwnershipConflictError builds a stable error for ownership conflicts.
func ResourceOwnershipConflictError(obj metav1.Object) error {
	if ns := obj.GetNamespace(); ns != "" {
		return fmt.Errorf(
			"resource %s/%s already exists and is not owned by the expected owner; refusing to overwrite",
			ns, obj.GetName(),
		)
	}
	return fmt.Errorf(
		"resource %s already exists and is not owned by the expected owner; refusing to overwrite",
		obj.GetName(),
	)
}
