package logfilemetricsexporter

import (
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/apis/logging/v1alpha1"
)

func Validate(instance *loggingv1alpha1.LogFileMetricExporter) (error, *loggingv1alpha1.LogFileMetricExporterStatus) {
	returnStatus := loggingv1alpha1.LogFileMetricExporterStatus{}
	for _, validate := range validations {
		if err, status := validate(instance); err != nil {
			return err, status
		} else if status != nil {
			returnStatus.Conditions = append(returnStatus.Conditions, status.Conditions...)
		}
	}
	return nil, &returnStatus
}

// validations are the set of admission rules for validating a LogFileMetricExporter instance
var validations = []func(instance *loggingv1alpha1.LogFileMetricExporter) (error, *loggingv1alpha1.LogFileMetricExporterStatus){
	validateName,
}
