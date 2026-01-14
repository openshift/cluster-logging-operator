package logfilemetricsexporter

import (
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/api/logging/v1alpha1"
)

func Validate(instance *loggingv1alpha1.LogFileMetricExporter) error {
	for _, validate := range validations {
		if err := validate(instance); err != nil {
			return err
		}
	}
	return nil
}

// validations are the set of admission rules for validating a LogFileMetricExporter instance
var validations = []func(instance *loggingv1alpha1.LogFileMetricExporter) error{
	validateName,
}
