package cluster

import "k8s.io/apimachinery/pkg/runtime/schema"

var (
	defaultClusterResources = []schema.GroupVersionResource{

		// Core resources
		{Group: "", Version: "v1", Resource: "nodes"},
		{Group: "", Version: "v1", Resource: "persistentvolumes"},

		// RBAC
		{Group: groupRbac, Version: "v1", Resource: "clusterroles"},
		{Group: groupRbac, Version: "v1", Resource: "clusterrolebindings"},

		// API Extensions
		{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"},

		// OpenShift Config
		{Group: GroupConfig, Version: "v1", Resource: "clusterversions"},
	}
)
