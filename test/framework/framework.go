package framework

import (
	"fmt"
	"time"

	testclient "github.com/openshift/cluster-logging-operator/test/client"
	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
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

func newReplicaWaitCondition[T any](cast func(watch.Event) (*T, bool), specReplicas func(*T) *int32, readyReplicas func(*T) int32) testclient.Condition {
	return func(event watch.Event) (bool, error) {
		obj, ok := cast(event)
		if !ok {
			return false, fmt.Errorf("expected %T but got %T", (*T)(nil), event.Object)
		}
		desired := int32(1)
		if r := specReplicas(obj); r != nil {
			desired = *r
		}
		return readyReplicas(obj) == desired, nil
	}
}

func NewDeploymentWaitCondition(_ *apps.Deployment) testclient.Condition {
	return newReplicaWaitCondition(
		func(e watch.Event) (*apps.Deployment, bool) { d, ok := e.Object.(*apps.Deployment); return d, ok },
		func(d *apps.Deployment) *int32 { return d.Spec.Replicas },
		func(d *apps.Deployment) int32 { return d.Status.AvailableReplicas },
	)
}

func NewStatefulSetWaitCondition(_ *apps.StatefulSet) testclient.Condition {
	return newReplicaWaitCondition(
		func(e watch.Event) (*apps.StatefulSet, bool) { s, ok := e.Object.(*apps.StatefulSet); return s, ok },
		func(s *apps.StatefulSet) *int32 { return s.Spec.Replicas },
		func(s *apps.StatefulSet) int32 { return s.Status.ReadyReplicas },
	)
}
