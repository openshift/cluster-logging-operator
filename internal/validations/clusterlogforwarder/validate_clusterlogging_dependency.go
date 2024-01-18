package clusterlogforwarder

import (
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/validations/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ValidateClusterLoggingDependencyMSG = "is dependent on a ClusterLogging instance with a valid spec.collector configuration"
)

func ValidateClusterLoggingDependency(clf loggingv1.ClusterLogForwarder, k8sClient client.Client, extras map[string]bool) (error, *loggingv1.ClusterLogForwarderStatus) {
	if clf.Name == constants.SingletonName && clf.Namespace == constants.OpenshiftNS && !extras[constants.ClusterLoggingAvailable] {
		return errors.NewValidationError("ClusterLogForwarder '%s/%s' %s", constants.OpenshiftNS, constants.SingletonName, ValidateClusterLoggingDependencyMSG), nil
	}
	return nil, nil
}
