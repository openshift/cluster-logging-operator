package client

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/test/runtime"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

var final = map[corev1.PodPhase]bool{
	corev1.PodFailed: true, corev1.PodSucceeded: true, corev1.PodRunning: true,
}

func podInPhase(o runtime.Object, phase corev1.PodPhase) (bool, error) {
	pod := o.(*corev1.Pod)
	switch {
	case pod.Status.Phase == phase:
		return true, nil
	case final[pod.Status.Phase]:
		return false, fmt.Errorf("%w: %v: want phase=%v, got %v", ErrWatchClosed, runtime.ID(pod), phase, pod.Status.Phase)
	}
	for _, container := range pod.Status.ContainerStatuses {
		if container.State.Waiting != nil &&
			container.State.Waiting.Reason == "CreateContainerError" {
			return false, fmt.Errorf("%w: %v container %vhas CreateContainerError", ErrWatchClosed, runtime.ID(pod), container.Name)
		}
	}
	return false, nil
}

// PodSucceeded returns (true,nil) when e.Object is a Pod with phase PodSucceeded.
// Returns an error if pod reaches any other long-lasting state	[failed, succeeded ,running]
func PodSucceeded(e watch.Event) (bool, error) { return podInPhase(e.Object, corev1.PodSucceeded) }

// PodFailed returns (true,nil) when e.Object is a Pod with phase PodFailed.
// Returns an error if pod reaches any other long-lasting state	[failed, succeeded ,running]
func PodFailed(e watch.Event) (bool, error) { return podInPhase(e.Object, corev1.PodFailed) }

// PodRunning returns (true,nil) when e.Object is a Pod with phase PodRunning.
// Returns an error if pod reaches any other long-lasting state	[failed, succeeded ,running]
func PodRunning(e watch.Event) (bool, error) { return podInPhase(e.Object, corev1.PodRunning) }

// DaemonSetIsReady returns (true, nil) when the pods of the set are ready on all nodes
// in the cluster and (false, nil) otherwise.
func IsDaemonSetReady(e watch.Event) (bool, error) {
	ds, ok := e.Object.(*appsv1.DaemonSet)
	if !ok {
		return false, fmt.Errorf("event is not for a daemonset: %v", e)
	}
	return ds.Status.NumberUnavailable == 0, nil
}
