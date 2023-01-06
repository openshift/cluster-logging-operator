package runtime

import (
	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
)

// NewClusterLogForwarder returns a ClusterLogForwarder with default name and namespace.
func NewClusterLogForwarder(namespace, name string) *loggingv1.ClusterLogForwarder {
	clf := &loggingv1.ClusterLogForwarder{}
	Initialize(clf, namespace, name)
	return clf
}
