package clusterlogforwarder

import (
	"strings"

	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	vErrors "github.com/openshift/cluster-logging-operator/internal/validations/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ValidateOutputTuning(clf loggingv1.ClusterLogForwarder, k8sClient client.Client, extras map[string]bool) (error, *loggingv1.ClusterLogForwarderStatus) {

	for _, oSpec := range clf.Spec.Outputs {
		if tuningErrors := oSpec.Tuning.Validate(); len(tuningErrors) > 0 {
			asString := []string{}
			for _, error := range tuningErrors {
				asString = append(asString, error.Error())
			}
			return vErrors.NewValidationError("Invalid Output tuning(s): %v", strings.Join(asString, ",")), nil
		}
	}

	return nil, nil
}
