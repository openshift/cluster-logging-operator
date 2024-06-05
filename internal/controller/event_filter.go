package controller

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// IgnoreMigratedResources filters and ignores events for the specified annotation
// It also ignores create events
func IgnoreMigratedResources(annotation string) predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore if generation does not change (status updates)
			if e.ObjectOld.GetGeneration() == e.ObjectNew.GetGeneration() {
				return false
			}
			if status, ok := e.ObjectNew.GetAnnotations()[annotation]; ok {
				return !(status == "true")
			}
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			if status, ok := e.Object.GetAnnotations()[annotation]; ok {
				return !(status == "true")
			}
			return true
		},
		// Do not allow create events to trigger
		CreateFunc: func(e event.CreateEvent) bool {
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			if status, ok := e.Object.GetAnnotations()[annotation]; ok {
				return !(status == "true")
			}
			return true
		},
	}
}
