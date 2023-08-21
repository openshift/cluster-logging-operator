package logfilemetricsexporter

import (
	"fmt"
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/apis/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/validations/errors"
)

// Check for singleton. Must be named instance and in openshift-logging namespace
func validateName(instance *loggingv1alpha1.LogFileMetricExporter) (error, *loggingv1alpha1.LogFileMetricExporterStatus) {
	failMessage := ""
	fail := false
	if instance.Name != constants.SingletonName {
		failMessage += fmt.Sprintf("Invalid name %q, singleton instance must be named %q ", instance.Name, constants.SingletonName)
		fail = true
	}
	if instance.Namespace != constants.OpenshiftNS {
		failMessage += fmt.Sprintf("Invalid namespace name %q, instance must be in %q namespace ", instance.Namespace, constants.OpenshiftNS)
		fail = true
	}
	if fail {
		return errors.NewValidationError(failMessage), nil
	}
	return nil, nil
}
