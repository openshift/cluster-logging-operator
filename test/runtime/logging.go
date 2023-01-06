package runtime

import (
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test"
)

// NewClusterLogForwarder returns a ClusterLogForwarder with default name and namespace.
func NewClusterLogForwarder() *loggingv1.ClusterLogForwarder {
	return runtime.NewClusterLogForwarder(constants.WatchNamespace, constants.SingletonName)
}

// NewClusterLogging returns a ClusterLogging with default name, namespace and
// collection configuration. No store, visualization or curation are configured,
// see ClusterLoggingDefaultXXX to add them.
func NewClusterLogging() *loggingv1.ClusterLogging {
	cl := &loggingv1.ClusterLogging{}
	runtime.Initialize(cl, constants.WatchNamespace, constants.SingletonName)
	test.MustUnmarshal(`
    collection:
      type: fluentd
    managementState: Managed
    `, &cl.Spec)
	return cl
}

// ClusterLoggingDefaultStore sets default store configuration.
func ClusterLoggingDefaultStore(cl *loggingv1.ClusterLogging) {
	test.MustUnmarshal(`
    type: "elasticsearch"
    elasticsearch:
      nodeCount: 1
      redundancyPolicy: "ZeroRedundancy"
      resources:
        limits:
          cpu: 500m
          memory: 4Gi
`, &cl.Spec.LogStore)
}

// ClusterLoggingDefaultVisualization sets default visualization configuration.
func ClusterLoggingDefaultVisualization(cl *loggingv1.ClusterLogging) {
	test.MustUnmarshal(`
    type: "kibana"
    kibana:
      replicas: 1
`, &cl.Spec.Visualization)
}

// ClusterLoggingDefaultCuration sets defautl curation configuration.
func ClusterLoggingDefaultCuration(cl *loggingv1.ClusterLogging) {
	test.MustUnmarshal(`
    type: "curator"
    curator:
      schedule: "30 3,9,15,21 * * *"
`, &cl.Spec.Curation)
}
