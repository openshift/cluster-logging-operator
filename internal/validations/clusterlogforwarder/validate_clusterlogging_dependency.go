package clusterlogforwarder

import (
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/validations/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ValidateClusterLoggingDependency(clf loggingv1.ClusterLogForwarder, k8sClient client.Client, extras map[string]bool) (error, *loggingv1.ClusterLogForwarderStatus) {
	if clf.Name == constants.SingletonName && clf.Namespace == constants.OpenshiftNS && !extras[constants.ClusterLoggingAvailable] {
		return errors.NewValidationError("ClusterLogForwarder '%s/%s' is dependent on a ClusterLogging instance", constants.OpenshiftNS, constants.SingletonName), nil
	}
	return nil, nil
}
