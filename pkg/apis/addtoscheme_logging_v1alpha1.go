package apis

import (
	v1alpha1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme)
}
