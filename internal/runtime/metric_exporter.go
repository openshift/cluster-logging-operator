package runtime

import (
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/apis/logging/v1alpha1"
)

// Returns a new default LogFileMetricExporter
func NewLogFileMetricExporter(namespace, name string) *loggingv1alpha1.LogFileMetricExporter {
	lfme := &loggingv1alpha1.LogFileMetricExporter{}
	Initialize(lfme, namespace, name)
	return lfme
}
