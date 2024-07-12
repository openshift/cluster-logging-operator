package controller

import (
	"github.com/openshift/cluster-logging-operator/internal/constants"

	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// IgnoreMigratedResources filters and ignores events related to migration from logging.openshift.io to observability.openshift.io
// It ignores create events except for the initial start up when existing resources are available.
func IgnoreMigratedResources() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore if generation does not change (status updates)
			if e.ObjectOld.GetGeneration() == e.ObjectNew.GetGeneration() {
				return false
			}

			// Ignore already converted resources
			if status, ok := e.ObjectNew.GetAnnotations()[constants.AnnotationCRConverted]; ok {
				return !(status == "true")
			}

			// Only allow updates to existing resources
			// Example: If a user removes fluentDForward from an existing resource, should convert to the new API
			// Disallows when a user creates a new resource and update it
			if needsMigration, ok := e.ObjectNew.GetAnnotations()[constants.AnnotationNeedsMigration]; ok {
				return needsMigration == "true"
			}

			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			if status, ok := e.Object.GetAnnotations()[constants.AnnotationCRConverted]; ok {
				return !(status == "true")
			}
			return true
		},
		// Do not allow create events to trigger
		CreateFunc: func(e event.CreateEvent) bool {
			// When operator starts up, existing resources generate a "create" event
			// ensure only existing resources get migrated through an annotation
			if needsMigration, ok := e.Object.GetAnnotations()[constants.AnnotationNeedsMigration]; ok {
				return needsMigration == "true"
			}
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			if status, ok := e.Object.GetAnnotations()[constants.AnnotationCRConverted]; ok {
				return !(status == "true")
			}
			return true
		},
	}
}
