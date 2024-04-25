package observability

import (
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Initializer is a fuction that knows how to initialize a kubernetes runtime object
type Initializer func(o runtime.Object, namespace, name string, visitors ...func(o runtime.Object))

// NewClusterLogForwarder returns a ClusterLogForwarder with name and namespace.
func NewClusterLogForwarder(namespace, name string, initialize Initializer) *obsv1.ClusterLogForwarder {
	clf := &obsv1.ClusterLogForwarder{}
	initialize(clf, namespace, name)
	return clf
}
