package utils

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EnsureCanUpdateOwnedResource returns nil when the object has not been persisted yet
// (create path) or when it is already owned by all of the desired owners.
// If the object exists and is not owned by the desired owners (foreign-owned or
// unowned), it returns an error and callers must not mutate the object.
//
// Ownership is matched by OwnerReference UID so additional ownerRefs on the
// object do not cause false conflicts (LOG-9591).
func EnsureCanUpdateOwnedResource(obj metav1.Object, desiredOwners []metav1.OwnerReference) error {
	if obj.GetResourceVersion() == "" {
		return nil
	}
	if IsOwnedByDesired(obj.GetOwnerReferences(), desiredOwners) {
		return nil
	}
	return ResourceOwnershipConflictError(obj)
}

// IsOwnedByDesired reports whether every desired owner UID is present among current owners.
// An empty desired owner list only matches when current owners are also empty.
func IsOwnedByDesired(current, desired []metav1.OwnerReference) bool {
	if len(desired) == 0 {
		return len(current) == 0
	}
	currentUIDs := make(map[string]struct{}, len(current))
	for _, ref := range current {
		currentUIDs[string(ref.UID)] = struct{}{}
	}
	for _, want := range desired {
		if _, ok := currentUIDs[string(want.UID)]; !ok {
			return false
		}
	}
	return true
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
