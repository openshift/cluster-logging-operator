package e2e

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	clolog "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/test/helpers"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func (tc *E2ETestFramework) WaitFor(component helpers.LogComponentType) error {
	clolog.Info("Waiting for component to be ready", "component", component)
	switch component {
	case helpers.ComponentTypeVisualization:
		return tc.waitForDeployment(constants.OpenshiftNS, "kibana", defaultRetryInterval, defaultTimeout)
	case helpers.ComponentTypeCollector, helpers.ComponentTypeCollectorVector:
		return tc.WaitForDaemonSet(constants.OpenshiftNS, "mycollector")
	case helpers.ComponentTypeStore, helpers.ComponentTypeReceiverElasticsearchRHManaged:
		return tc.waitForElasticsearchPods(defaultRetryInterval, defaultTimeout)
	case helpers.ComponentTypeCollectorDeployment:
		return tc.waitForDeployment(constants.OpenshiftNS, constants.CollectorName, defaultRetryInterval, defaultTimeout)
	}
	return fmt.Errorf("Unable to waitfor unrecognized component: %v", component)
}

func (tc *E2ETestFramework) WaitForDaemonSet(namespace, name string) error {
	// daemonset should have pods running and available on all the nodes for maxtimes * retryInterval
	maxtimes := 5
	times := 0
	return wait.PollUntilContextTimeout(context.TODO(), defaultRetryInterval, defaultTimeout, true, func(cxt context.Context) (done bool, err error) {
		numUnavail, err := oc.Literal().From(fmt.Sprintf("oc -n %s get ds/%s --ignore-not-found -o jsonpath={.status.numberUnavailable}", namespace, name)).Run()
		if err == nil {
			if numUnavail == "" {
				numUnavail = "0"
			}
			value, err := strconv.Atoi(strings.TrimSpace(numUnavail))
			if err != nil {
				times = 0
				return false, err
			}
			if value == 0 {
				times++
			} else {
				times = 0
			}
			if times == maxtimes {
				return true, nil
			}
		}
		return false, nil
	})
}

// WaitForResourceCondition retrieves resource info given a selector and evaluates the jsonPath output against the provided condition.
func (tc *E2ETestFramework) WaitForResourceCondition(namespace, kind, name, selector, jsonPath string, maxtimes int, isSatisfied func(string) (bool, error)) error {
	times := 0
	return wait.PollUntilContextTimeout(context.TODO(), 5*time.Second, defaultTimeout, true, func(cxt context.Context) (done bool, err error) {
		out, err := oc.Get().WithNamespace(namespace).Resource(kind, name).Selector(selector).OutputJsonpath(jsonPath).Run()
		if err != nil {
			clolog.V(3).Error(err, "Error returned from retrieving resources")
			return false, nil
		}
		met, err := isSatisfied(out)
		if err != nil {
			times = 0
			clolog.V(3).Error(err, "Error returned from condition check")
			return false, nil
		}
		if met {
			times++
			clolog.V(4).Info("Condition met", "success", times, "need", maxtimes)
		} else {
			times = 0
		}
		if times == maxtimes {
			clolog.V(3).Info("Condition met", "success", times, "need", maxtimes)
			return true, nil
		}
		return false, nil
	})
}

func (tc *E2ETestFramework) waitForElasticsearchPods(retryInterval, timeout time.Duration) error {
	clolog.V(3).Info("Waiting for elasticsearch")
	return wait.PollUntilContextTimeout(context.TODO(), retryInterval, timeout, true, func(cxt context.Context) (done bool, err error) {

		options := metav1.ListOptions{
			LabelSelector: "component=elasticsearch",
		}
		pods, err := tc.KubeClient.CoreV1().Pods(constants.OpenshiftNS).List(context.TODO(), options)
		if err != nil {
			if apierrors.IsNotFound(err) {
				clolog.V(2).Error(err, "Did not find elasticsearch pods")
				return false, nil
			}
			clolog.Error(err, "Error listing elasticsearch pods")
			return false, nil
		}
		numberOfPods := len(pods.Items)
		if numberOfPods == 0 {
			clolog.V(2).Info("No elasticsearch pods found ", "pods", pods)
			return false, nil
		}
		containersReadyCount := 0
		containersNotReadyCount := 0
		for _, pod := range pods.Items {
			for _, status := range pod.Status.ContainerStatuses {
				clolog.V(3).Info("Checking status of", "PodName", pod.Name, "ContainerID", status.ContainerID, "status", status.Ready)
				if status.Ready {
					containersReadyCount++
				} else {
					containersNotReadyCount++
				}
			}
		}
		if containersReadyCount == 0 || containersNotReadyCount > 0 {
			clolog.V(3).Info("elasticsearch containers are not ready", "pods", numberOfPods, "ready containers", containersReadyCount, "not ready containers", containersNotReadyCount)
			return false, nil
		}

		return true, nil
	})
}

func (tc *E2ETestFramework) waitForDeployment(namespace, name string, retryInterval, timeout time.Duration) error {
	return wait.PollUntilContextTimeout(context.TODO(), retryInterval, timeout, true, func(cxt context.Context) (done bool, err error) {
		deployment, err := tc.KubeClient.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			clolog.Error(err, "Error trying to retrieve Deployment")
			return false, nil
		}
		replicas := int(*deployment.Spec.Replicas)
		if int(deployment.Status.AvailableReplicas) == replicas {
			return true, nil
		}
		return false, nil
	})
}

func (tc *E2ETestFramework) WaitForCleanupCompletion(namespace string, podlabels []string) {
	if !DoCleanup() {
		return
	}
	if err := tc.waitForClusterLoggingPodsCompletion(namespace, podlabels); err != nil {
		clolog.Error(err, "Cleanup completion error")
	}
}

func (tc *E2ETestFramework) waitForClusterLoggingPodsCompletion(namespace string, podlabels []string) error {
	labels := strings.Join(podlabels, ",")
	labelSelector := fmt.Sprintf("component in (%s)", labels)
	options := metav1.ListOptions{
		LabelSelector: labelSelector,
	}
	clolog.Info("waiting for pods to complete with labels in namespace:", "labels", labels, "namespace", namespace, "options", options)

	return wait.PollUntilContextTimeout(context.TODO(), defaultRetryInterval, defaultTimeout, true, func(cxt context.Context) (done bool, err error) {
		pods, err := tc.KubeClient.CoreV1().Pods(namespace).List(context.TODO(), options)
		if err != nil {
			if apierrors.IsNotFound(err) {
				// this is OK because we want them to not exist
				clolog.Error(err, "Did not find pods")
				return true, nil
			}
			clolog.Error(err, "Error listing pods ")
			return false, err //issue with how the command was crafted
		}
		if len(pods.Items) == 0 {
			clolog.Info("No pods found for label selection", "labels", labels)
			return true, nil
		}
		clolog.V(5).Info("pods still running", "num", len(pods.Items))
		return false, nil
	})
}

func (tc *E2ETestFramework) waitForStatefulSet(namespace, name string, retryInterval, timeout time.Duration) error {
	err := wait.PollUntilContextTimeout(context.TODO(), retryInterval, timeout, true, func(cxt context.Context) (done bool, err error) {
		deployment, err := tc.KubeClient.AppsV1().StatefulSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return false, nil
			}
			clolog.Error(err, "Error Getting StatfuleSet")
			return false, nil
		}
		replicas := int(*deployment.Spec.Replicas)
		if int(deployment.Status.ReadyReplicas) == replicas {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	return nil
}
