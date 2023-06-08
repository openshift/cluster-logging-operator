package clusterlogforwarder

import (
	v1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/validations/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validateSingleton(clf v1.ClusterLogForwarder, k8sClient client.Client, extras map[string]bool) (error, *v1.ClusterLogForwarderStatus) {
	if clf.Name != constants.SingletonName {
		return errors.NewValidationError("Invalid name %q, singleton instance must be named %q", clf.Name, constants.SingletonName), nil
	}
	return nil, nil
}
