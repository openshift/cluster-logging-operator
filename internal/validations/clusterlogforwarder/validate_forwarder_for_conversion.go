package clusterlogforwarder

import (
	"fmt"

	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers/security"
	"github.com/openshift/cluster-logging-operator/internal/migrations/observability/api/outputs"
	"github.com/openshift/cluster-logging-operator/internal/validations/clusterlogforwarder/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

func ValidateClusterLogForwarderForConversion(clfInstance *loggingv1.ClusterLogForwarder, k8sClient client.Client) (map[string]*corev1.Secret, *loggingv1.ClusterLogForwarderStatus, error) {
	if clfInstance == nil {
		return nil, nil, nil
	}

	var outputSecrets map[string]*corev1.Secret
	var missingSecrets *loggingv1.NamedConditions
	var err error

	status := loggingv1.ClusterLogForwarderStatus{}

	if outputs.ReferencesFluentDForward(&clfInstance.Spec) {
		status.Conditions.SetCondition(conditions.CondInvalid("cannot migrate CLF because fluentDForward is referenced as an output."))
		return nil, &status, fmt.Errorf("cannot migrate CLF with FluentDForward as an output")
	}

	outputSecrets, missingSecrets, err = security.MakeOutputSecretMap(k8sClient, clfInstance.Spec.Outputs, clfInstance.Namespace)
	if err != nil && !errors.IsNotFound(err) {
		return nil, nil, err
	}
	// Update status if missing secrets/ return error
	if missingSecrets != nil {
		status.Outputs = *missingSecrets
		status.Conditions.SetCondition(conditions.CondMissing("outputs have defined secrets that are missing"))
		return nil, &status, fmt.Errorf("output/s are missing defined secret/s")
	}
	return outputSecrets, &status, nil
}
