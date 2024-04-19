package runtime

import (
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test"
)

// NewClusterLogForwarder returns a ClusterLogForwarder with default name and namespace.
func NewClusterLogForwarder() *loggingv1.ClusterLogForwarder {
	return runtime.NewClusterLogForwarder(constants.OpenshiftNS, constants.SingletonName)
}

// NewClusterLogging returns a ClusterLogging with default name, namespace and
// collection configuration. No store, visualization or curation are configured,
// see ClusterLoggingDefaultXXX to add them.
func NewClusterLogging() *loggingv1.ClusterLogging {
	cl := runtime.NewClusterLogging(constants.OpenshiftNS, constants.SingletonName)
	test.MustUnmarshal(`
    collection:
      type: fluentd
    managementState: Managed
    `, &cl.Spec)
	return cl
}
