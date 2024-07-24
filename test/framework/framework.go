package framework

import (
	testclient "github.com/openshift/cluster-logging-operator/test/client"
	"k8s.io/client-go/kubernetes"
	"time"
)

const (

	// DefaultRetryInterval is the retry interval that is used when polling
	DefaultRetryInterval = 1 * time.Second
)

type Test interface {

	//AddCleanup registers a function to be called when the test terminates
	AddCleanup(fn func() error)

	// KubeClient returns a kubernetes client for interacting with the cluster under test
	ClientTyped() *kubernetes.Clientset

	// Client is an untyped test client that can read and write from the API server
	Client() *testclient.Client

	//PodExec executes a command in a specific container of a pod
	PodExec(namespace, name, container string, command []string) (string, error)
}
