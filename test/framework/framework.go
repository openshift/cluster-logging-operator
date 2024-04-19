package framework

import (
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

	//KubeClient returns a kubernetes client for interacting with the cluster under test
	Client() *kubernetes.Clientset

	//PodExec executes a command in a specific container of a pod
	PodExec(namespace, name, container string, command []string) (string, error)
}
