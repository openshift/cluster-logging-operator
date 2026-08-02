package collection

import "k8s.io/apimachinery/pkg/runtime/schema"

var (
	clfGVR = schema.GroupVersionResource{
		Group:    "observability.openshift.io",
		Version:  "v1",
		Resource: "clusterlogforwarders",
	}
	lfmeGVR = schema.GroupVersionResource{
		Group:    "logging.openshift.io",
		Version:  "v1alpha1",
		Resource: "logfilemetricexporters",
	}
)
