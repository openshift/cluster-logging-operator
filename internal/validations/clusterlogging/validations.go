package clusterlogging

import (
	v1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Validate(cl v1.ClusterLogging, k8sClient client.Client, extras map[string]bool) error {
	for _, validate := range validations {
		if err := validate(cl); err != nil {
			return err
		}
	}
	return nil
}

// validations are the set of admission rules for validating
// a ClusterLogging
var validations = []func(cl v1.ClusterLogging) error{
	validateClusterLoggingSpec,
}
