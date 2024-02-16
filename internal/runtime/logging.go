package runtime

import (
	loggingv1 "github.com/openshift/cluster-logging-operator/api/logging/v1"
)

// NewClusterLogForwarder returns a ClusterLogForwarder with default name and namespace.
func NewClusterLogForwarder(namespace, name string) *loggingv1.ClusterLogForwarder {
	clf := &loggingv1.ClusterLogForwarder{}
	Initialize(clf, namespace, name)
	return clf
}

// NewClusterLogging returns a ClusterLogging with default name and namespace.
func NewClusterLogging(namespace, name string) *loggingv1.ClusterLogging {
	cl := &loggingv1.ClusterLogging{}
	Initialize(cl, namespace, name)
	return cl
}
