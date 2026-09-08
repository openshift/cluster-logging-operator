package ui

import (
	"github.com/openshift/cluster-logging-operator/must-gather/internal/cluster"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// Collect ConsolePlugins
	consolePluginGVR = schema.GroupVersionResource{
		Group:    "console.openshift.io",
		Version:  "v1",
		Resource: "consoleplugins",
	}

	// Collect Console ClusterOperator
	coGVR = schema.GroupVersionResource{
		Group:    cluster.GroupConfig,
		Version:  "v1",
		Resource: "clusteroperators",
	}
)
